package main

import (
	"fmt"
	"sync"
)

// go run -race mutex_1.go
func main() {
	cnt := 0
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cnt++
			}
		}()
	}
	wg.Wait()
	fmt.Printf("cnt: %d\n", cnt)
}
