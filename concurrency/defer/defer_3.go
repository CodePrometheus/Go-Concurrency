package main

import "fmt"

// 匿名函数没有通过传参的方式将 n 传入，因此匿名函数内的 n 和函数外部的 n 是同一个
// 延迟执行时，已经被改变为 101.
func d() {
	n := 1
	defer func() {
		fmt.Println(n)
	}()
	n += 100
}

// 匿名返回值函数 0
// 首先函数返回时会自动创建一个返回变量假设为ret，函数返回时要将res赋值给ret，即有ret = res，也就是说ret=0
func anonymousReturnValues() int {
	var res int
	defer func() {
		res++
		fmt.Println("defer")
	}()
	return res
}

// 命名返回值函数 1 - 不会存在这一个问题，因为不需要去创建临时变量
func namedReturnValues() (res int) {
	defer func() {
		res++
		fmt.Println("defer")
	}()
	return
}

func main() {
	v1 := anonymousReturnValues()
	v2 := namedReturnValues()
	fmt.Printf("%v,%v", v1, v2)
}
