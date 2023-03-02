package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	msg := make(chan int, 10)
	defer close(msg)
	// producer
	for i := 0; i < 10; i++ {
		msg <- i
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()

	// consumer
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			select {
			case <-ctx.Done():
				fmt.Println("child process interrupt")
				return
			default:
				fmt.Printf("send msg: %d\n", <-msg)
			}
		}
	}(ctx)

	select {
	case <-ctx.Done():
		time.Sleep(time.Second)
		fmt.Println("main process exit")
	}
}
