package main

import "fmt"

func create() func() int {
	c := 2 // 捕获变量
	return func() int {
		return c
	}
}

func main() {
	f1 := create()
	f2 := create()
	fmt.Println(f1())
	fmt.Println(f2())
}
