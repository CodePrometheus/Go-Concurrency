package main

import "sync"

// 两个 G 交替打印数字
func main() {
	var c = make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i += 2 {
			print(i)
			c <- struct{}{}
			<-c
		}
	}()
	go func() {
		defer wg.Done()
		for i := 2; i <= 10; i += 2 {
			<-c
			print(i)
		}
	}()
}
