package main

import (
	"fmt"
	"sync"
)

// no data race
func main() {
	mu := sync.Mutex{}
	cnt := 0
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				mu.Lock()
				cnt++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("cnt: %d\n", cnt)
}
