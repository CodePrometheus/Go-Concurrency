package main

import (
	"fmt"
	"sync"
)

// 交替打印数字和字母 12AB34C
func main() {
	num, let := make(chan bool), make(chan bool)
	wait := sync.WaitGroup{}
	go func() {
		i := 1
		for {
			select {
			case <-num:
				if i > 26 {
					wait.Done()
					return
				}
				fmt.Printf("%d", i)
				i++
				let <- true
			}
		}
	}()
	wait.Add(1)
	go func(wait *sync.WaitGroup) {
		i := 'A'
		for {
			select {
			case <-let:
				if i > 'Z' {
					wait.Done()
					return
				}
				fmt.Printf("%s", string(i))
				i++
				num <- true
			}
		}
	}(&wait)
	num <- true
	wait.Wait()
}
