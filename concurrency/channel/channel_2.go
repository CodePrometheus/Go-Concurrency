package main

import (
	"fmt"
	"time"
)

// 操作	  nil的channel	正常channel	已关闭channel
// <- ch	  阻塞	    成功或阻塞	    读到零值
// ch <-	  阻塞	       成功或阻塞      panic
// close(ch)  panic	       成功	         panic
// 特别的：当nil的通道在select的某个case中时，这个case会阻塞，但不会造成死锁
func main() {
	c := make(chan int)
	go send(c)
	go recv(c)
	time.Sleep(time.Second * 2)
}

func send(c chan<- int) {
	for i := 0; i < 10; i++ {
		c <- i
	}
}

func recv(c <-chan int) {
	for i := range c {
		fmt.Println(i)
	}
}
