package main

import "fmt"

func order() (i int) {
	i = 0
	defer func() {
		i += 1
		fmt.Println("defer 2")
	}()
	return i
}

func main() {
	fmt.Println("return: ", order())
}
