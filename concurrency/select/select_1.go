package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(time.Second)
		ch1 <- "ch1"
	}()

	go func() {
		time.Sleep(time.Second)
		ch2 <- "ch2"
	}()

	select {
	case a1 := <-ch1:
		fmt.Println(a1)
	case a2 := <-ch2:
		fmt.Println(a2)
	}
}
