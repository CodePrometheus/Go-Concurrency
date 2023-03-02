package main

import "fmt"

// N 个协程交替打印英文字母
func main() {
	gNum := 4 // 协程数
	chanQueue := make([]chan struct{}, gNum)
	ret := 0
	exitChan := make(chan struct{})

	for i := 0; i < gNum; i++ {
		chanQueue[i] = make(chan struct{})
		if i == gNum-1 {
			go func(i int) {
				chanQueue[i] <- struct{}{}
			}(i)
		}
	}

	for i := 0; i < gNum; i++ {
		var lastChan, curChan chan struct{}
		if i == 0 {
			lastChan = chanQueue[gNum-1]
		} else {
			lastChan = chanQueue[i-1]
		}
		curChan = chanQueue[i]
		go func(i byte, lastChan, curChan chan struct{}) {
			for {
				if ret > 20 {
					exitChan <- struct{}{}
				}
				<-lastChan
				fmt.Printf("%c\n", i)
				ret++
				curChan <- struct{}{}
			}
		}('A'+byte(i), lastChan, curChan)
	}
	<-exitChan
	fmt.Println("Done")
}
