package main

import (
	"context"
	"fmt"
	"time"
)

func cancel() {
	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				fmt.Println("got the stop channel")
				return
			default:
				fmt.Println("still working")
				time.Sleep(time.Second)
			}
		}
	}()
	time.Sleep(3 * time.Second)
	fmt.Println("stop")
	stop <- true
	time.Sleep(3 * time.Second)
}

func withCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("got the stop channel")
				return
			default:
				fmt.Println("still working")
				time.Sleep(time.Second)
			}
		}
	}()

	time.Sleep(3 * time.Second)
	fmt.Println("stop")
	cancel()
	time.Sleep(3 * time.Second)
}

func main() {
	// cancel()
	withCancel()
}
