package leakybucket

import (
	"container/heap"
	"sync"
	"sync/atomic"
)

// 订阅 g，等待固定量 token
type subscriber struct {
	tokens   int64
	sequence int64
	ret      chan error
}

type subscribers []*subscriber

var (
	subscriberPool = &sync.Pool{
		New: func() interface{} {
			return &subscriber{}
		},
	}
	subscriberSequence int64
)

func (s subscribers) Len() int {
	return len(s)
}

func (s subscribers) Less(i, j int) bool {
	if s[i].tokens < s[j].tokens {
		return true
	}
	if s[i].tokens > s[j].tokens {
		return false
	}
	if s[i].sequence < s[j].sequence {
		return true
	}
	return false
}

func (s subscribers) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *subscribers) Push(x any) {
	*s = append(*s, x.(*subscriber))
}

func (s *subscribers) Pop() any {
	old := *s
	n := len(old)
	news := old[n-1]
	old[n-1] = nil
	*s = old[0 : n-1]
	return news
}

var _ heap.Interface = new(subscribers)

func NewSubscriber(n int64) *subscriber {
	s := subscriberPool.Get().(*subscriber)
	s.ret = make(chan error, 1)
	s.tokens = n
	s.sequence = atomic.AddInt64(&subscriberSequence, 1)
	return s
}

func SetSubscriber(s *subscriber) {
	s.ret = nil
	subscriberPool.Put(s)
}
