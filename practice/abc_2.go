package main

import (
	"fmt"
	"time"
)

type Token struct{}

// 1...10 交替打印
func main() {
	channels := make([]chan Token, 10)
	for i := 0; i < 10; i++ {
		channels[i] = make(chan Token)
	}

	for i := 0; i < 10; i++ {
		go func(index int, current chan Token, nextChan chan Token) {
			for {
				<-current
				fmt.Printf("G %d \n", index)
				time.Sleep(time.Second)
				nextChan <- Token{}
			}
		}(i+1, channels[i], channels[(i+1)%10])
	}
	channels[0] <- Token{}
	select {}
}
