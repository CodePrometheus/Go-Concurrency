package main

import (
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(1) // 任意时刻只允许 1 个 M 执行 Go 代码
	var wg sync.WaitGroup

	for i := 1; i <= 258; i++ {
		wg.Add(1)
		go func(n int) {
			println(n)
			wg.Done()
		}(i)
	}

	wg.Wait() // 3 1 2
}
