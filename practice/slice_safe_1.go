package main

import (
	"fmt"
	"sync"
)

var a []int

func add1() {
	for i := 0; i < 100; i++ {
		go func() {
			a = append(a, 1)
		}()
	}
}

var mu sync.Mutex
var wg sync.WaitGroup

func add2() {
	nums := 100
	wg.Add(nums)
	for i := 0; i < nums; i++ {
		go func() {
			mu.Lock()
			a = append(a, 1)
			mu.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
}

func add3() {
	c := make(chan int)
	nums := 100
	wg.Add(nums)
	for i := 0; i < nums; i++ {
		go func() {
			c <- 1
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	for i := range c {
		a = append(a, i)
	}
}

var cnt = 0

// 使用索引
func add4() {
	nums := 100
	wg.Add(nums)
	b := make([]int, nums, nums)
	for i := 0; i < nums; i++ {
		k := i // 局部变量
		go func(idx int) {
			b[idx] = 1
			wg.Done()
		}(k)
	}
	wg.Wait()
	for i := range b {
		if b[i] != 0 {
			cnt++
		}
	}
}

func main() {
	// add1()
	// fmt.Println("add1: ", len(a))
	// add2()
	// fmt.Println("add2: ", len(a))
	// add3()
	// fmt.Println("add3: ", len(a))
	add4()
	fmt.Println("add4: ", cnt)
}
