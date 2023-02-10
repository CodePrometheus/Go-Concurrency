package main

import (
	"sync"
)

var ch = make(chan struct{})
var waitGroup sync.WaitGroup

func main() {
	waitGroup.Add(2)
	go G1()
	go G2()
	waitGroup.Wait()
}

func G1() {
	defer waitGroup.Done()
	for i := 1; i <= 10; i += 2 {
		println(i)
		ch <- struct{}{}
		<-ch
	}
}

func G2() {
	defer waitGroup.Done()
	for i := 2; i <= 10; i += 2 {
		<-ch
		println(i)
		ch <- struct{}{}
	}
}
