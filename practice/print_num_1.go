package main

import (
	"fmt"
	"sync"
)

// 两个 G 交替打印数字
func main() {
	ch := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 1; i < 101; i++ {
			ch <- struct{}{}
			if i%2 == 1 {
				fmt.Println("线程1打印: ", i)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 1; i < 101; i++ {
			<-ch
			if i%2 == 0 {
				fmt.Println("线程2打印: ", i)
			}
		}
	}()
	wg.Wait()
}
