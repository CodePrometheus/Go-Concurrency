package main

import (
	"fmt"
	"runtime"
	"time"
)

// 基于信号的抢占式调度
func main() {
	runtime.GOMAXPROCS(1)
	fmt.Println("starting")
	go func() {
		for {
		}
	}()
	time.Sleep(time.Second)
	fmt.Println("I got scheduled")
}
