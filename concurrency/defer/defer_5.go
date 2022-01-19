package main

import "fmt"

func mq() int {
	i := 0
	defer func() {
		fmt.Println("defer 1")
	}()

	defer func() {
		fmt.Println("defer 2")
	}()
	return i
}

// defer 的执行顺序：后进先出。但是返回值并没有被修改，这是由于 Go 的返回机制决定的，执行 return 语句后
// Go 会创建一个临时变量保存返回值，因此，defer 语句修改了局部变量 i，并没有修改返回值。
func main() {
	// defer 2
	// defer 1
	// return:  0
	fmt.Println("return: ", mq())
}
