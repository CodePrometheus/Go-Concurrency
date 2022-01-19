package main

import (
	"fmt"
)

// defer 的作用域是函数，而不是代码块，因此 if 语句退出时，defer 不会执行，而是等 101 打印后，整个函数返回时，才会执行.
func main() {
	n := 1
	if n == 1 {
		defer fmt.Printf("defer: %v\n", n)
		n += 100
	}
	fmt.Printf("print: %v\n", n)
}
