package main

import (
	"fmt"
	"sync"
)

// sync.Once可以用于只关闭一次通道，防止对已关闭的通道关闭导致的宕机
var on sync.Once
var wait sync.WaitGroup

func main() {
	ch1 := make(chan int, 10)
	ch2 := make(chan int, 10)
	wait.Add(3)

	go f1(ch1)
	go f2(ch1, ch2)
	go f2(ch1, ch2)

	wait.Wait()
	for res := range ch2 {
		fmt.Println(res)
	}
}

// write
func f1(ch chan<- int) {
	defer wait.Done()
	for i := 0; i < 10; i++ {
		ch <- i
	}
	fmt.Println("关闭ch1")
	close(ch)
}

// gen
func f2(ch1 <-chan int, ch2 chan<- int) {
	defer wait.Done()
	for {
		x, ok := <-ch1
		if !ok {
			break
		}
		ch2 <- 2 * x
	}
	on.Do(func() {
		fmt.Println("关闭ch2")
		close(ch2)
	})
}
