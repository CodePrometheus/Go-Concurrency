package workerpool

import (
	"context"
	"fmt"
	"go.uber.org/goleak"
	"sync"
	"testing"
	"time"
)

const max = 20

func TestSimple(t *testing.T) {
	defer goleak.VerifyNone(t) // goroutine 泄漏检测

	wp := New(2)
	requests := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	rspChan := make(chan string, len(requests))

	for _, r := range requests {
		r := r
		wp.Submit(func() {
			rspChan <- r
		})
	}

	wp.StopWait()
	close(rspChan)

	rspSet := map[string]struct{}{}
	for rsp := range rspChan {
		rspSet[rsp] = struct{}{}
	}
	if len(rspSet) < len(requests) {
		t.Fatal("Did not handle all requests")
	}
	for _, req := range requests {
		if _, ok := rspSet[req]; !ok {
			t.Fatal("Missing expected values:", req)
		}
	}
}

func TestMaxWorkers(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(0)
	wp.Stop()
	if wp.maxWorkers != 1 {
		t.Fatal("should have created one worker")
	}

	wp = New(max)
	defer wp.Stop()

	if wp.Size() != max {
		t.Fatal("wrong size returned")
	}

	started := make(chan struct{}, max)
	release := make(chan struct{})

	// 启动 workers，在完成前在 channel 中等待
	for i := 0; i < max; i++ {
		wp.Submit(func() {
			started <- struct{}{}
			<-release
		})
	}
	// 等待所有入队的任务分配给 workers
	if wp.waitingQueue.Len() != wp.WaitingQueueSize() {
		t.Fatal("Working Queue size returned should not be 0")
	}

	timeout := time.After(5 * time.Second)
	for startCount := 0; startCount < max; {
		select {
		case <-started:
			startCount++
		case <-timeout:
			t.Fatal("timed out waiting for workers to start")
		}
	}
	close(release)
}

func TestReuseWorkers(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(5)
	defer wp.Stop()

	release := make(chan struct{})
	// 在下一次 task 创建是复用 worker
	for i := 0; i < 10; i++ {
		wp.Submit(func() {
			<-release
		})
		release <- struct{}{}
		time.Sleep(time.Millisecond)
	}
	close(release)
	// 如果同一个 worker 总是被重用，那么只会创建一个 worker，并且应该只有一个 ready
	if countReady(wp) > 1 {
		t.Fatal("Worker not reused")
	}
}

func TestWorkerTimeout(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(max)
	defer wp.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	wp.Pause(ctx)

	if anyReady(wp) {
		t.Fatal("number of ready workers should be zero")
	}

	if wp.killIdleWorker() {
		t.Fatal("should have been no idle workers to kill")
	}
	cancel()

	fmt.Println("count: ", countReady(wp))
	if countReady(wp) != max {
		t.Fatal("Expected", max, "ready workers")
	}

	// 测试 worker 超时情况
	time.Sleep(idleTimeout*2 + idleTimeout/2)
	if countReady(wp) != max-1 {
		t.Fatal("First worker did not timeout")
	}
	time.Sleep(idleTimeout)
	if countReady(wp) != max-2 {
		t.Fatal("Second worker did not timeout")
	}
}

func TestStop(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(max)
	ctx, cancel := context.WithCancel(context.Background())
	wp.Pause(ctx)
	cancel()

	if wp.Stopped() {
		t.Fatal("pool should not be stopped")
	}

	wp.Stop()
	if anyReady(wp) {
		t.Fatal("should have zero workers after stop")
	}

	if !wp.Stopped() {
		t.Fatal("pool should be stopped")
	}

	wp = New(5)
	release := make(chan struct{})
	finished := make(chan struct{}, max)
	for i := 0; i < max; i++ {
		wp.Submit(func() {
			<-release
			finished <- struct{}{}
		})
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(release)
	}()
	wp.Stop()
	var cnt int

Count:
	for cnt < max {
		select {
		case <-finished:
			cnt++
		default:
			break Count
		}
	}

	if cnt > 5 {
		t.Fatal("Should not have completed any queued tasks, did", cnt)
	}
	wp.Stop()
}

func TestStopWait(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(5)
	release := make(chan struct{})
	finished := make(chan struct{}, max)

	for i := 0; i < max; i++ {
		wp.Submit(func() {
			<-release
			finished <- struct{}{}
		})
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(release)
	}()
	wp.StopWait()
	for cnt := 0; cnt < max; cnt++ {
		select {
		case <-finished:
		default:
			t.Fatal("Should have completed all queued tasks")
		}
	}

	if anyReady(wp) {
		t.Fatal("should have zero workers after StopWait")
	}

	if !wp.Stopped() {
		t.Fatal("pool should be stopped")
	}

	wp = New(5)
	wp.StopWait()

	if anyReady(wp) {
		t.Fatal("should have zero workers after StopWait")
	}
	wp.StopWait()
}

func TestSubmitWait(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(1)
	defer wp.Stop()

	wp.Submit(nil)
	wp.SubmitWait(nil)

	done1 := make(chan struct{})
	wp.Submit(func() {
		time.Sleep(100 * time.Millisecond)
		close(done1)
	})
	select {
	case <-done1:
		t.Fatal("Submit did not return immediately")
	default:
	}

	done2 := make(chan struct{})
	wp.SubmitWait(func() {
		time.Sleep(100 * time.Millisecond)
		close(done2)
	})
	select {
	case <-done2:
	default:
		t.Fatal("SubmitWait did not wait for function to execute")
	}
}

func TestOverflow(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(2)
	defer wp.Stop()
	releaseChan := make(chan struct{})
	for i := 0; i < 64; i++ {
		wp.Submit(func() {
			<-releaseChan
		})
	}

	go func() {
		<-time.After(time.Millisecond)
		close(releaseChan)
	}()
	wp.Stop()

	// 现在工作池已经退出，可以安全地检查它的等待队列而不会引起竞争
	len := wp.waitingQueue.Len()
	if len != 62 {
		t.Fatal("Expected 62 tasks in waiting queue, have", len)
	}
}

func TestStopRace(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(max)
	defer wp.Stop()

	workRelChan := make(chan struct{})
	var started sync.WaitGroup
	started.Add(max)

	for i := 0; i < max; i++ {
		wp.Submit(func() {
			started.Done()
			<-workRelChan
		})
	}
	started.Wait()

	const doneCallers = 5
	stopDone := make(chan struct{}, doneCallers)
	for i := 0; i < doneCallers; i++ {
		go func() {
			wp.Stop()
			stopDone <- struct{}{}
		}()
	}

	select {
	case <-stopDone:
		t.Fatal("Stop should not return in any goroutine")
	default:
	}

	close(workRelChan)

	timeout := time.After(time.Second)
	for i := 0; i < doneCallers; i++ {
		select {
		case <-stopDone:
		case <-timeout:
			wp.Stop()
			t.Fatal("timeout waiting for Stop to return")
		}
	}
}

func TestWaitingQueueSizeRace(t *testing.T) {
	defer goleak.VerifyNone(t)
	const (
		goroutines = 10
		tasks      = 20
		workers    = 5
	)
	wp := New(workers)
	defer wp.Stop()
	maxChan := make(chan int)

	for i := 0; i < goroutines; i++ {
		go func() {
			max := 0
			for j := 0; j < tasks; j++ {
				wp.Submit(func() {
					time.Sleep(time.Microsecond)
				})
				waiting := wp.WaitingQueueSize()
				if waiting > max {
					max = waiting
				}
			}
			maxChan <- max
		}()
	}

	maxMax := 0
	for i := 0; i < goroutines; i++ {
		max := <-maxChan
		if max > maxMax {
			maxMax = max
		}
	}
	if maxMax == 0 {
		t.Error("expected to see waiting queue size > 0")
	}
	if maxMax >= goroutines*tasks {
		t.Error("should not have seen all tasks on waiting queue")
	}
}

func TestPause(t *testing.T) {
	defer goleak.VerifyNone(t)

	wp := New(25)
	defer wp.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	ran := make(chan struct{})
	wp.Submit(func() {
		time.Sleep(time.Millisecond)
		close(ran)
	})
	wp.Pause(ctx)

	select {
	case <-ran:
	default:
		t.Error("did not run all tasks before returning from Pause")
	}
	ran = make(chan struct{})
	wp.Submit(func() {
		close(ran)
	})

	select {
	case <-ran:
		t.Error("ran while paused")
	case <-time.After(time.Millisecond):
	}

	if wp.WaitingQueueSize() != 1 {
		t.Error("waiting queue size should be 1")
	}
	cancel()

	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Error("did not run after canceling pause")
	}

	// ---- Test pause while paused

	ctx, cancel = context.WithCancel(context.Background())
	wp.Pause(ctx)

	ctx2, cancel2 := context.WithCancel(context.Background())

	pauseDone := make(chan struct{})
	go func() {
		wp.Pause(ctx2)
		close(pauseDone)
	}()

	// Check that second pause does not return until first pause in canceled
	select {
	case <-pauseDone:
		wp.Stop()
		t.Fatal("second Pause should not have returned")
	case <-time.After(time.Millisecond):
	}

	cancel() // cancel 1st pause

	// Check that second pause returns
	select {
	case <-pauseDone:
	case <-time.After(time.Second):
		wp.Stop()
		t.Fatal("timed out waiting for Pause to return")
	}

	cancel2() // cancel 2nd pause

	// ---- Test concurrent pauses

	ctx, cancel = context.WithCancel(context.Background())
	ctx2, cancel2 = context.WithCancel(context.Background())
	pauseDone = make(chan struct{})
	pause2Done := make(chan struct{})
	go func() {
		wp.Pause(ctx)
		close(pauseDone)
	}()
	go func() {
		wp.Pause(ctx2)
		close(pause2Done)
	}()

	select {
	case <-pauseDone:
		cancel()
		<-pause2Done
		cancel2()
	case <-pause2Done:
		cancel2()
		<-pauseDone
		cancel()
	case <-time.After(time.Second):
		t.Fatal("concurrent pauses deadlocked")
	}

	// ---- Test stopping paused pool ----

	ctx, cancel = context.WithCancel(context.Background())
	ctx2, cancel2 = context.WithCancel(context.Background())

	// Stack up two pauses
	wp.Pause(ctx)
	go wp.Pause(ctx2)

	ran = make(chan struct{})
	wp.Submit(func() {
		close(ran)
	})

	stopDone := make(chan struct{})
	go func() {
		wp.StopWait()
		close(stopDone)
	}()

	// Check that task was run after calling StopWait
	select {
	case <-stopDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for StopWait to return")
	}

	// Check that task was run after calling StopWait
	select {
	case <-ran:
	default:
		t.Error("did not run after canceling pause")
	}

	defer cancel()
	defer cancel2()

	// ---- Test pause after stop ----

	ctx, cancel = context.WithCancel(context.Background())
	pauseDone = make(chan struct{})
	go func() {
		wp.Pause(ctx)
		close(pauseDone)
	}()
	select {
	case <-pauseDone:
	case <-time.After(time.Second):
		t.Fatal("pause after stop did not return")
	}
	cancel()
}

func TestWorkerLeak(t *testing.T) {
	defer goleak.VerifyNone(t)
	const workerCnt = 100
	wp := New(workerCnt)

	for i := 0; i < workerCnt; i++ {
		wp.Submit(func() {
			time.Sleep(time.Millisecond)
		})
	}
	wp.Stop()
}

func anyReady(wp *WorkerPool) bool {
	release := make(chan struct{})
	wait := func() {
		<-release
	}
	select {
	case wp.workerQueue <- wait:
		close(release)
		return true
	default:
	}
	return false
}

func countReady(wp *WorkerPool) int {
	release := make(chan struct{})
	wait := func() {
		<-release
	}
	var readCount int
	for i := 0; i < max; i++ {
		select {
		case wp.workerQueue <- wait:
			readCount++
		// select 超时时间
		case <-time.After(100 * time.Millisecond):
			i = max
		}
	}
	close(release)
	return readCount
}

// == Run benchmarking with: go test -bench '.'
func BenchmarkEnqueue(b *testing.B) {
	wp := New(1)
	defer wp.Stop()
	releaseChan := make(chan struct{})

	b.ResetTimer()

	// Start workers, and have them all wait on a channel before completing.
	for i := 0; i < b.N; i++ {
		wp.Submit(func() { <-releaseChan })
	}
	close(releaseChan)
}

func BenchmarkEnqueue2(b *testing.B) {
	wp := New(2)
	defer wp.Stop()

	b.ResetTimer()

	// Start workers, and have them all wait on a channel before completing.
	for i := 0; i < b.N; i++ {
		releaseChan := make(chan struct{})
		for i := 0; i < 64; i++ {
			wp.Submit(func() { <-releaseChan })
		}
		close(releaseChan)
	}
}

func BenchmarkExecute1Worker(b *testing.B) {
	benchmarkExecWorkers(1, b)
}

func BenchmarkExecute2Worker(b *testing.B) {
	benchmarkExecWorkers(2, b)
}

func BenchmarkExecute4Workers(b *testing.B) {
	benchmarkExecWorkers(4, b)
}

func BenchmarkExecute16Workers(b *testing.B) {
	benchmarkExecWorkers(16, b)
}

func BenchmarkExecute64Workers(b *testing.B) {
	benchmarkExecWorkers(64, b)
}

func BenchmarkExecute1024Workers(b *testing.B) {
	benchmarkExecWorkers(1024, b)
}

func benchmarkExecWorkers(n int, b *testing.B) {
	wp := New(n)
	defer wp.Stop()
	var allDone sync.WaitGroup
	allDone.Add(b.N * n)

	b.ResetTimer()

	// Start workers, and have them all wait on a channel before completing.
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			wp.Submit(func() {
				//time.Sleep(100 * time.Microsecond)
				allDone.Done()
			})
		}
	}
	allDone.Wait()
}
