package workerpool

import (
	"context"
	"github.com/gammazero/deque"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// 如果 workers 空闲超过该时间，则停止一个 worker
	idleTimeout = 2 * time.Second
)

type WorkerPool struct {
	maxWorkers int // 最大 worker 数量

	taskQueue   chan func() // task 队列
	workerQueue chan func() // worker 队列

	stoppedChan chan struct{} // 停止队列
	stopSignal  chan struct{} // 停止信号

	waitingQueue deque.Deque[func()] // 等待队列

	stopLock sync.Mutex // 停止锁
	stopOnce sync.Once  // 停止 Once
	stopped  bool       // 已停止
	waiting  int32      // 等待数量
	wait     bool       // 是否等待
}

// New 创建并启动一个 worker pool
func New(maxWorkers int) *WorkerPool {
	// 必须至少有一个 worker
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	pool := &WorkerPool{
		maxWorkers:  maxWorkers,
		taskQueue:   make(chan func()),
		workerQueue: make(chan func()),
		stoppedChan: make(chan struct{}),
		stopSignal:  make(chan struct{}),
	}

	// 启动任务调度器
	go pool.dispatch()

	return pool
}

// dispatch 派发任务给可用的 worker
func (p *WorkerPool) dispatch() {
	defer close(p.stoppedChan)
	timeout := time.NewTimer(idleTimeout)
	var workerCount int // 现有的 worker 总数
	var idle bool
	var wg sync.WaitGroup

Loop:
	for {
		// 只要 waitingQueue 有任务，传入的任务就会被放入等待队列，要运行的任务就会从等待队列中取出
		// 一旦等待队列为空，则直接向可用 workers 提交传入任务
		if p.waitingQueue.Len() > 0 {
			if !p.processWaitingQueue() {
				break Loop
			}
			continue
		}

		select {
		// 正常从 taskQueue 取
		case task, ok := <-p.taskQueue:
			if !ok {
				break Loop
			}
			// 取一个 task 执行
			select {
			case p.workerQueue <- task:
			default:
				// 在未达到最大值时，创建一个新的 worker
				if workerCount < p.maxWorkers {
					wg.Add(1)
					go worker(task, p.workerQueue, &wg)
					workerCount++
				} else {
					// 进入队列的 task 将由下一个可用的 worker 执行
					p.waitingQueue.PushBack(task)
					atomic.StoreInt32(&p.waiting, int32(p.waitingQueue.Len()))
				}
			}
			idle = false

		// 等待 work 到来超时，如果 pool 空闲了整整一个超时时间则杀死一个准备好的 worker
		case <-timeout.C:
			if idle && workerCount > 0 {
				if p.killIdleWorker() {
					workerCount--
				}
			}
			idle = true
			timeout.Reset(idleTimeout)
		}
	}

	// 若被告知等待，则运行已经入队的 tasks
	if p.wait {
		p.runQueuedTasks()
	}

	// 停止剩下的所有 workers 当它们准备好时
	for workerCount > 0 {
		p.workerQueue <- nil
		workerCount--
	}

	wg.Wait()
	timeout.Stop()
}

// runQueuedTasks 不断从 waiting queue 里取出 task 执行直到队列为空
func (p *WorkerPool) runQueuedTasks() {
	for p.waitingQueue.Len() != 0 {
		p.workerQueue <- p.waitingQueue.PopFront()
		atomic.StoreInt32(&p.waiting, int32(p.waitingQueue.Len()))
	}
}

// worker 执行 tasks，且当收到 nil task 时停止
func worker(task func(), workerQueue chan func(), wg *sync.WaitGroup) {
	for task != nil {
		task()
		task = <-workerQueue
	}
	wg.Done()
}

// killIdleWorker
func (p *WorkerPool) killIdleWorker() bool {
	select {
	case p.workerQueue <- nil:
		// 给 worker 发送 kill 信号
		return true
	default:
		// 说明无准备好的 workers
		return false
	}
}

// processWaitingQueue 将新的任务放入等待队列，同时当 workers 可用时从 waitingQueue 移除对应的 task
// 返回 false 当 worker pool is stopped
func (p *WorkerPool) processWaitingQueue() bool {
	select {
	case task, ok := <-p.taskQueue:
		if !ok {
			return false
		}
		// ok 则放入
		p.waitingQueue.PushBack(task)
	case p.workerQueue <- p.waitingQueue.Front():
		// 当一个 worker 可用时，派发一个任务
		p.waitingQueue.PopFront()
	}
	// 写入等待队列的长度
	atomic.StoreInt32(&p.waiting, int32(p.waitingQueue.Len()))
	return true
}

// Submit 将 task 入队以供 worker 执行
// 无论有多少个 task 提交，submit 都不会阻塞
// 如果没有可用的 workers 并且 workers 的数量已经是最大值，则会进入等待队列
// 主要没有新任务来，每段时间关闭一个可用的 workers 直到无
func (p *WorkerPool) Submit(task func()) {
	if task != nil {
		p.taskQueue <- task
	}
}

// SubmitWait 将给定的 task 入队并等待它被执行
func (p *WorkerPool) SubmitWait(task func()) {
	if task == nil {
		return
	}
	doneChan := make(chan struct{})
	p.taskQueue <- func() {
		task()
		close(doneChan)
	}
	<-doneChan
}

// StopWait 停止 WorkerPool 并等待所有 task 完全完成
// 不能提交额外的任务，但所有挂起的任务都在该函数返回之前由 workers 执行
func (p *WorkerPool) StopWait() {
	p.stop(true)
}

// Stop 停止工作池并只等待当前正在运行的任务完成
// 当前未运行的挂起任务将被放弃
func (p *WorkerPool) Stop() {
	p.stop(false)
}

// stop 告诉 dispatcher 退出，以及是否完成排队任务
func (p *WorkerPool) stop(wait bool) {
	p.stopOnce.Do(func() {
		// 发出 workerpool 正在停止的信号，以取消暂停任何已经暂停 workers
		close(p.stopSignal)
		// 所有正在进行的暂停都将完成
		p.stopLock.Lock()
		p.stopped = true
		p.stopLock.Unlock()
		p.wait = wait
		close(p.taskQueue)
	})
	<-p.stoppedChan
}

// Size 返回并发 workers 的最大值
func (p *WorkerPool) Size() int {
	return p.maxWorkers
}

// WaitingQueueSize 返回等待队列中的任务数
func (p *WorkerPool) WaitingQueueSize() int {
	return int(atomic.LoadInt32(&p.waiting))
}

// Pause 导致所有 workers 等待给定的上下文，从而使他们无法运行任务
// 当所有 workers 都在等待时返回
// task 可以继续入队 但直到上下文取消或超时才会执行
// 当 WorkerPool 已经暂停时，调用该函数会导致 Pause 等待直到所有先前的暂停都被取消
// 当 workerpool 停止时，worker 将取消暂停，并在 StopWait 期间执行排队的任务
func (p *WorkerPool) Pause(ctx context.Context) {
	p.stopLock.Lock()
	defer p.stopLock.Unlock()
	if p.stopped {
		return
	}
	ready := new(sync.WaitGroup)
	ready.Add(p.maxWorkers)
	for i := 0; i < p.maxWorkers; i++ {
		p.Submit(func() {
			ready.Done()
			select {
			case <-ctx.Done():
			case <-p.stopSignal:
			}
		})
	}
	// 等待所有 workers 都被暂停
	ready.Wait()
}

// Stopped WorkerPool 是否被暂停
func (p *WorkerPool) Stopped() bool {
	p.stopLock.Lock()
	defer p.stopLock.Unlock()
	return p.stopped
}
