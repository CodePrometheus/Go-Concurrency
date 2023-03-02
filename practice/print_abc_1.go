package main

import (
	"fmt"
	"sync"
)

func main() {
	var c1, c2, c3 = make(chan struct{}), make(chan struct{}), make(chan struct{})
	var wg sync.WaitGroup
	n := 10

	wg.Add(3)
	go func(s string) {
		defer wg.Done()
		for i := 1; i <= n; i++ {
			<-c1
			fmt.Print(s)
			c2 <- struct{}{}
		}
		<-c1
	}("A")
	go func(s string) {
		defer wg.Done()
		for i := 1; i <= n; i++ {
			<-c2
			fmt.Print(s)
			c3 <- struct{}{}
		}
	}("B")
	go func(s string) {
		defer wg.Done()
		for i := 1; i <= n; i++ {
			<-c3
			fmt.Print(s)
			c1 <- struct{}{}
		}
	}("C")
	c1 <- struct{}{}
	wg.Wait()
}
