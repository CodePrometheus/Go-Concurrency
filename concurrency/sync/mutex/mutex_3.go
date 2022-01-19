package main

import (
	"fmt"
	"sync"
)

// Cnt state int32 互斥锁的状态
// sema uint32 信号量 用于控制锁的状态
type Cnt struct {
	sync.Mutex
	Num uint64
}

func main() {
	cnt := Cnt{}
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cnt.Lock()
				cnt.Num++
				cnt.Unlock()
			}
		}()
	}
	wg.Wait() // 使得main等子协程运行完再结束，wait在值为0才会放行，否则阻塞
	fmt.Println(cnt.Num)
}
