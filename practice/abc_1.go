package main

import "sync"

var pa = make(chan struct{}, 0)
var pb = make(chan struct{}, 0)

var pc = make(chan struct{}, 0)

var WG = sync.WaitGroup{}

func main() {
	WG.Add(1)
	go PA()
	go PB()
	// go PC()
	pa <- struct{}{}
	WG.Wait()
}

func PA() {
	i := 0
	for {
		select {
		case <-pa:
			i++
			if i > 100 {
				WG.Done()
				return
			}
			print("A")
			pb <- struct{}{}
		}
	}
}

func PB() {
	for {
		select {
		case <-pb:
			print("B")
			//pc <- struct{}{}
			pa <- struct{}{}
		}
	}
}

func PC() {
	for {
		select {
		case <-pc:
			print("C")
			pa <- struct{}{}
		}
	}
}
