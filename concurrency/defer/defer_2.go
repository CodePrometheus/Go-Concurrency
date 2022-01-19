package main

import "fmt"

type T struct{}

func (t T) f(n int) T {
	fmt.Println(n)
	return t
}

// defer 延迟调用时，需要保存函数指针和参数，因此链式调用的情况下，除了最后一个函数/方法外的函数/方法都会在调用时直接执行
// 也就是说 t.f(1) 直接执行，然后执行 fmt.Println(3)，最后函数返回时再执行 .f(2)，因此输出是 132.
func main() {
	var t T
	defer t.f(1).f(2)
	fmt.Println(3)
}
