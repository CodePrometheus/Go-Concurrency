package main

import "fmt"

func main() {
	ch := make(chan string)
	go func() {
		fmt.Println("Begin")
		ch <- "Mid"
		fmt.Println("End")
	}()
	fmt.Println("main在等新协程运行完毕")
	<-ch // 会在收到新协程写入通道的消息之前一直阻塞
	fmt.Println("main运行结束")
}
