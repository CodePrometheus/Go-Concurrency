package main

import (
	"context"
	"time"
)

func main() {
	ctx := context.Background()
	s := newStore()
	sub1 := newSubscriber("sub1")
	sub2 := newSubscriber("sub2")
	sub3 := newSubscriber("sub3")

	_ = s.subscribe(ctx, sub1)
	_ = s.subscribe(ctx, sub2)
	_ = s.subscribe(ctx, sub3)

	s.publish(ctx, &message{data: []byte("test1")})
	s.publish(ctx, &message{data: []byte("test2")})
	s.publish(ctx, &message{data: []byte("test3")})

	time.Sleep(time.Second)

}
