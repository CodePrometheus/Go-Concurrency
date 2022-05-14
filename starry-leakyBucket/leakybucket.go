package leakybucket

import (
	"container/heap"
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	minTimerResolution = time.Millisecond
)

var (
	ErrClosed           = errors.New("LeakyBucket: already closed")
	ErrNotEnough        = errors.New("LeakyBucket: not enough")
	ErrOverFlow         = errors.New("LeakyBucket: over flow")
	ErrDeadlineExceeded = errors.New("LeakyBucket: deadline exceeded")
)

type LeakyBucket struct {
	mu sync.Mutex

	max   int64
	total int64
	tps   float64 // token 产生速率

	lastUpdated time.Time
	subscribers subscribers

	closed  int32 // 关闭状态，所有接口直接失败
	started int32 // 启动状态，等待检测

	signal chan struct{}
	exit   chan struct{}
}

func (lb *LeakyBucket) Lock(f func() error) error {
	if atomic.LoadInt32(&lb.closed) == 1 {
		return ErrClosed
	}
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if atomic.LoadInt32(&lb.closed) == 1 {
		return ErrClosed
	}
	return f()
}

func (lb *LeakyBucket) UpdateTokens() {
	if lb.tps == 0 {
		return
	}
	now := time.Now()
	diff := now.Sub(lb.lastUpdated).Seconds() * lb.tps
	ret := math.Abs(diff)

	// 将 token 变化量向下取整，并尽可能的将时间改成恰好增加这么多 token 的时间，缓解精度问题
	if ret >= 1.0 {
		added := math.Floor(ret)
		rounded := time.Duration(math.Floor(added * float64(time.Second) / math.Abs(lb.tps)))
		lb.lastUpdated = lb.lastUpdated.Add(rounded)

		if diff > 0 {
			lb.total += int64(added)
		} else {
			lb.total -= int64(added)
		}

		if lb.total < 0 {
			lb.total = 0
		} else if lb.total > lb.max {
			lb.total = lb.max
		}
	}
}

func (lb *LeakyBucket) Trigger() {
	if len(lb.subscribers) == 0 {
		return
	}
	select {
	case lb.signal <- struct{}{}:
	default:
	}
}

// ResizeMax 重新设定上限值
func (lb *LeakyBucket) ResizeMax(n int64) error {
	if n < 0 {
		n = 0
	}
	return lb.Lock(func() error {
		lb.max = n
		if lb.total > lb.max {
			lb.total = lb.max
		}
		return nil
	})
}

// New 创建
func New(config *Config) *LeakyBucket {
	if config.MaxTokens <= 0 {
		config.MaxTokens = math.MaxInt64
	}

	if config.InitTokens < 0 {
		config.InitTokens = 0
	}

	return &LeakyBucket{
		max:   config.MaxTokens,
		total: config.InitTokens,
		tps:   config.TokensPer,

		lastUpdated: time.Now(),

		signal: make(chan struct{}, 1),
		exit:   make(chan struct{}),
	}
}

// Get 获取 n 个 token
func (lb *LeakyBucket) Get(n int64) error {
	if n <= 0 {
		return nil
	}
	return lb.Lock(func() error {
		lb.UpdateTokens()
		next := lb.total - n
		if next < 0 {
			return ErrNotEnough
		}
		lb.total = next
		return nil
	})
}

// Set 放入 n 个 token
func (lb *LeakyBucket) Set(n int64) error {
	if n <= 0 {
		return nil
	}

	return lb.Lock(func() error {
		lb.UpdateTokens()
		next := lb.total + n
		if next > lb.max {
			return ErrOverFlow
		}
		lb.total = next

		// 若新增加 token 可以满足正在等待的 g
		lb.Trigger()
		return nil
	})
}

// Fill 放入 n 个 token
func (lb *LeakyBucket) Fill(n int64) (set int64) {
	if n <= 0 {
		return 0
	}
	lb.Lock(func() error {
		lb.UpdateTokens()
		next := lb.total + n
		if next > lb.max {
			set = lb.max - lb.total
			lb.total = lb.max
		} else {
			set = n
			lb.total = next
		}
		if set > 0 {
			lb.Trigger()
		}
		return nil
	})
	return
}

func (lb *LeakyBucket) Wait(n int64) error {
	if n <= 0 {
		return nil
	}
	ret, err := lb.WaitForToken(n)
	if err != nil {
		return err
	}
	if ret == nil {
		return nil
	}
	return <-ret
}

func (lb *LeakyBucket) WaitForToken(n int64) (ret chan error, err error) {
	err = lb.Lock(func() error {
		lb.UpdateTokens()
		next := lb.total - n
		if next >= 0 {
			lb.total = next
			return nil
		}
		ret = lb.Subscribe(n)
		return nil
	})
	lb.StartMonitor()
	return
}

// Subscribe 订阅者
func (lb *LeakyBucket) Subscribe(n int64) (ret chan error) {
	s := NewSubscriber(n)
	heap.Push(&lb.subscribers, s)
	ret = s.ret
	return
}

func (lb *LeakyBucket) StartMonitor() {
	if atomic.LoadInt32(&lb.started) > 0 {
		// 主动触发
		lb.Trigger()
		return
	}
	atomic.StoreInt32(&lb.started, 1)
	go lb.Monitor()
}

func (lb *LeakyBucket) Monitor() {
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		// 尝试拿出 token 需求最小的订阅者
		var s *subscriber
		var next time.Duration
		err := lb.Lock(func() error {
			// 是否存在已有些许订阅者的需求满足
			lb.UnsafeUnblockSubscribers()
			if len(lb.subscribers) == 0 {
				return nil
			}
			s = lb.subscribers[0]
			// 最短还需要多少时间能得到足够的 Token
			if lb.tps > 0 {
				d := time.Duration(math.Ceil(float64((s.tokens-lb.total)*time.Second.Nanoseconds()) / lb.tps))
				next = d - time.Now().Sub(lb.lastUpdated)
				if next < minTimerResolution {
					next = minTimerResolution
				}
			}
			return nil
		})
		if err != nil {
			return
		}

		// 若无订阅者，阻塞等待下一个订阅者到来
		if s == nil {
			select {
			case <-lb.signal:
				continue
			case <-lb.exit:
				return
			}
		}

		if lb.tps <= 0 {
			select {
			case <-lb.signal:
				lb.UnblockSubscribers()
				continue
			case <-lb.exit:
				return
			}
		}

		// 轮训，等待 next 时间观察是否有足够多的 Token
		if timer == nil {
			timer = time.NewTimer(next)
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(next)
		}

		select {
		case <-timer.C:
			lb.UnblockSubscribers()
		case <-lb.signal:
			lb.UnblockSubscribers()
		case <-lb.exit:
			return
		}
	}
}

func (lb *LeakyBucket) UnblockSubscribers() {
	lb.Lock(lb.UnsafeUnblockSubscribers)
}

func (lb *LeakyBucket) UnsafeUnblockSubscribers() error {
	if len(lb.subscribers) == 0 {
		return nil
	}
	lb.UpdateTokens()
	for len(lb.subscribers) > 0 {
		s := lb.subscribers[0]
		if s.tokens > lb.total {
			break
		}
		select {
		case s.ret <- nil:
			lb.total -= s.tokens
		default:
		}
		heap.Pop(&lb.subscribers)
	}
	return nil
}

// Close 所有接口返回错误， 所有 g 都被释放
func (lb *LeakyBucket) Close() error {
	var pending subscribers
	err := lb.Lock(func() error {
		atomic.StoreInt32(&lb.closed, 1)
		pending = lb.subscribers
		lb.subscribers = nil
		close(lb.exit)
		return nil
	})
	if err != nil && err != ErrClosed {
		return err
	}
	for _, s := range pending {
		select {
		case s.ret <- ErrClosed:
		default:
		}
		SetSubscriber(s)
	}
	return nil
}

// WaitContext 获取 n 个 token，若无足够会一直等待 lb 中可以取出足够的 token 或 ctx 超时
func (lb *LeakyBucket) WaitContext(ctx context.Context, n int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if n <= 0 {
		return nil
	}
	res, err := lb.WaitForToken(n)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}

	select {
	case err := <-res:
		return err
	case <-ctx.Done():
		err := ctx.Err()
		if err == context.DeadlineExceeded {
			err = ErrDeadlineExceeded
		}

		select {
		case res <- err:
		default:
		}
		return <-res
	}
}
