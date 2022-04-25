package main

import (
	"fmt"
	"sync"
)

func main() {
	var m sync.Map
	m.Store("a", 1)
	fmt.Println(m.Load("a"))
	fmt.Println(m.LoadOrStore("b", 1))
	fmt.Println(m.Load("b"))
	m.Delete("a")
	fmt.Println(m.LoadOrStore("a", 1))
	m.Range(func(key, value any) bool {
		fmt.Println(key, value)
		return true // false 只打印一条
	})

}
