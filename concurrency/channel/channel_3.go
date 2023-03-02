package main

import "fmt"

var ch = make(chan int)

func f1() {
	fmt.Println("F1() ing...")
	ch <- 1
}

// 保持全部子协程执行完毕,主协程再退出
func main() {
	go f1()
	<-ch
}
