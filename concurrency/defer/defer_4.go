package main

import "fmt"

func f(n int) {
	defer fmt.Println(n)
	n += 100
}

// defer 语句执行时，会将需要延迟调用的函数和参数保存起来，也就是说，执行到 defer 时，参数 n(此时等于1) 已经被保存了。
// 因此后面对 n 的改动并不会影响延迟函数调用的结果
func main() {
	f(1)
}
