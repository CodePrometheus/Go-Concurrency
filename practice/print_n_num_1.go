package main

import (
	"fmt"
	"sync"
)

var N = 5

// N 个协程交替打印数字
func main() {
	collaborate()
	// race()
}

// 通过 channel 通信协同打印
func collaborate() {
	ch := make([]chan bool, N+1)
	for i := 0; i <= N; i++ {
		ch[i] = make(chan bool)
	}

	for i := 1; i <= N; i++ {
		go func(id int) {
			// 当前协程阻塞，直到收到前一个协程的信号
			<-ch[id-1]
			// 打印自己对应的序号
			println(id)
			// 通知下一个协程干活
			ch[id] <- true
		}(i)
	}
	ch[0] <- true
	// 等待最后一个协程完成工作
	<-ch[N]
}

// 通过竞争互斥锁实现
func race() {
	// 当前执行到的序号
	curSeq := 1
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 1; i <= N; i++ {
		// 将当前协程加入wg
		wg.Add(1)

		go func(id int) {
			for {
				// 尝试获取锁，并打印数据
				mu.Lock()
				if id == curSeq {
					fmt.Println(curSeq)
					curSeq += 1
					// 打印工作完成，解锁并跳出循环
					mu.Unlock()
					break
				}
				// 发现没到自己打印的时候，解锁，把机会留给别人
				mu.Unlock()
			}

			// 子协程工作完成，离开wg
			wg.Done()
		}(i)
	}

	// 主协程等待子协程完成
	wg.Wait()
}
