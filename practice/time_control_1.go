package main

import (
	"fmt"
	"time"
)

func Run(taskId, sleepTime, timeout int, ch chan string) {
	chRun := make(chan string)
	go run(taskId, sleepTime, chRun)
	select {
	case re := <-chRun:
		ch <- re
	case <-time.After(time.Duration(timeout) * time.Second):
		re := fmt.Sprintf("task id %d , timeout", taskId)
		ch <- re
	}
}

func run(taskId, sleepTime int, ch chan string) {
	time.Sleep(time.Duration(sleepTime) * time.Second)
	ch <- fmt.Sprintf("task id %d , sleep %d second", taskId, sleepTime)
	return
}

func main() {
	input := []int{3, 2, 1}
	timeout := 2
	chLimit := make(chan bool, 1)
	chs := make([]chan string, len(input))
	limitFunc := func(chLimit chan bool, ch chan string, taskId, sleepTime, timeout int) {
		Run(taskId, sleepTime, timeout, ch)
		<-chLimit
	}
	startTime := time.Now()
	fmt.Println("Multi run start")
	for i, sleepTime := range input {
		chs[i] = make(chan string, 1) // buffered channel，用来解决本示例的 dead lock 问题。
		chLimit <- true
		go limitFunc(chLimit, chs[i], i, sleepTime, timeout)
	}

	for _, ch := range chs {
		fmt.Println(<-ch)
	}
	endTime := time.Now()
	fmt.Printf("Multi ssh finished. Process time %s. Number of task is %d", endTime.Sub(startTime), len(input))
}
