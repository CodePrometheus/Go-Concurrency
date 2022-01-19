package main

import (
	"fmt"
	"sync"
	"time"
)

// sync.Once
// done uint32 记录是否已完成过初始化，保证值进行一次初始化
// m Mutex 互斥锁，保证初始化操作的并发安全
// 用于解决一次性初始化问题，作用类似init函数，只执行一次，哪怕函数中发生了 panic
// 不同：init函数是在文件包首次被加载的时候执行；sync.Once是在代码运行中需要的时候执行
// 高并发场景中，需要确保某些操作只执行一次，比如只加载一次配置文件，只关闭一次通道

var once sync.Once

func main() {
	for i, v := range make([]string, 10) {
		once.Do(onces)
		fmt.Println("count:", v, "---", i)
	}
	for i := 0; i < 10; i++ {
		go func() {
			once.Do(onced)
			fmt.Println("213")
		}()
	}
	time.Sleep(4000)
}
func onces() {
	fmt.Println("onces")
}
func onced() {
	fmt.Println("onced")
}
