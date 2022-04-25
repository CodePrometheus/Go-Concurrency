package main

import (
	"fmt"
	"strconv"
	_ "strconv"
	"sync"
)

var d = make(map[string]int)
var w1 sync.WaitGroup

func Get(key string) int {
	return d[key]
}

func Set(key string, value int) {
	d[key] = value
}

func main() {
	// 模拟内置 map 并发下报错 fatal error: concurrent map writes
	for i := 0; i < 300; i++ {
		w1.Add(1)
		go func(n int) {
			key := strconv.Itoa(n) // 返回一个表示 x 的字符串
			Set(key, n)
			fmt.Printf("k=:%v,v:=%v\n", key, Get(key))
			w1.Done()
		}(i)
	}
	w1.Wait()
}
