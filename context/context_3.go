package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx := context.WithValue(context.Background(), "parameter", "1")
	go HandelRequest(ctx)
	time.Sleep(10 * time.Second)
}

func HandelRequest(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Done")
			return
		default:
			fmt.Println("running, parameter: ", ctx.Value("parameter"))
			time.Sleep(2 * time.Second)
		}
	}
}
