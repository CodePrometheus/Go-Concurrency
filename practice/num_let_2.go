package main

import (
	"fmt"
)

// 无缓冲
var num = make(chan struct{})
var let = make(chan struct{})
var down = make(chan struct{})

func f1() {
	for i := 1; i <= 26; i++ {
		fmt.Print(i)
		num <- struct{}{} // num 写
		<-let             // let 阻塞
	}
}

func f2() {
	for i := 'A'; i <= 'Z'; i++ {
		<-num // num 阻塞
		fmt.Print(string(i))
		let <- struct{}{} // let 写
	}
	down <- struct{}{}
}

func main() {
	go f1()
	go f2()
	<-down
}
