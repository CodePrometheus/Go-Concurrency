package main

import "fmt"

func main() {
	// pair<staticType:string, value:"hello">
	var a string
	a = "hello"

	// pair<type:string, value:"hello">
	var allType interface{}
	allType = a
	// 类型断言
	key, value := allType.(string)
	fmt.Println(key, value)
}
