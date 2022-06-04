package main

import (
	"fmt"
	"reflect"
)

func Func(arg any) {
	fmt.Println(arg)

	// 类型断言
	value, ok := arg.(string)
	if !ok {
		fmt.Println("arg is not string")
	} else {
		fmt.Println("arg is string, value = ", value)
		fmt.Println("value of arg is ", value)
		fmt.Println("type of arg is ", reflect.TypeOf(value))
	}
}

func main() {
	Func("starry")
}
