package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

func f1() {
	defer wg.Done()
	fmt.Println("F1() ing...")
}

// 保持全部子协程执行完毕,主协程再退出
func main() {
	wg.Add(1)
	go f1()
	wg.Wait()
}
