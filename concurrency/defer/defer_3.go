package main

import "fmt"

// 匿名函数没有通过传参的方式将 n 传入，因此匿名函数内的 n 和函数外部的 n 是同一个
// 延迟执行时，已经被改变为 101.
func main() {
	n := 1
	defer func() {
		fmt.Println(n)
	}()
	n += 100
}
