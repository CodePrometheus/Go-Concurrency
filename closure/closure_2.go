package main

import (
	"fmt"
	"time"
)

func test1() {
	i := 1
	go func() {
		time.Sleep(time.Second)
		fmt.Println("test1: ", i)
	}()
	i++
	time.Sleep(2 * time.Second)
}

// 闭包内捕获外部函数的参数的时候是取的地址,而不是调用闭包时刻的参数值
func test2() {
	i := 1
	go func(i int) {
		time.Sleep(time.Second)
		fmt.Println("test2: ", i)
	}(i) // 通过匿名函数参数将值传入闭包
	i++
	time.Sleep(2 * time.Second)
}

func main() {
	test1()
	test2()
}
