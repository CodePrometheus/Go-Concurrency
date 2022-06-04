package main

import (
	"fmt"
	"reflect"
)

func main() {
	reflectNum(3.4)
	user := User{"starry", 21}
	reflectObject(user)
}

func reflectNum(arg any) {
	fmt.Printf("type: %v\n", reflect.TypeOf(arg))   // valueOf 用来获取输入参数接口中的数据的值，如果接口为空则返回零值
	fmt.Printf("value: %v\n", reflect.ValueOf(arg)) // TypeOf 用来获取输入参数接口的数据类型，如果接口为空则返回 nil
}

type User struct {
	Name string
	Age  int
}

func reflectObject(arg any) {
	typeOf := reflect.TypeOf(arg)
	fmt.Println("type: ", typeOf.Name())
	valueOf := reflect.ValueOf(arg)
	fmt.Println("value: ", valueOf)

	// 获取结构体中的字段
	for i := 0; i < typeOf.NumField(); i++ {
		fmt.Println(typeOf.Field(i).Name, valueOf.Field(i))
	}
}
