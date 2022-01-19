package main

import (
	"fmt"
	"time"
)

// Timer 时间到 执行一次
func main() {
	// Sleep
	timer1 := time.NewTimer(time.Second)
	t1 := time.Now()
	fmt.Printf("t1: %v\n", t1)
	time.Sleep(time.Second * 3)
	t2 := <-timer1.C // 即使Sleep，t2中的时间仍然是t1后的1s
	fmt.Printf("t2: %v\n", t2)

	// timer只能执行一次，多次将会报错 fatal error: all goroutines are asleep - deadlock!
	//timer2 := time.NewTimer(time.Second)
	//for {
	//	<-timer2.C
	//	fmt.Println("Time is Up")
	//}

	// Stop
	timer3 := time.NewTimer(time.Second)
	go func() {
		<-timer3.C
		fmt.Println("没有停止timer3, 还是执行了")
	}()
	stop := timer3.Stop()
	if stop {
		fmt.Println("停止timer3成功")
	}

	// Reset
	timer4 := time.NewTimer(time.Second * 3)
	timer4.Reset(time.Second) // 1s后重置
	fmt.Println(time.Now())
	time.Sleep(time.Second * 4)
	// timer4.Reset(time.Second * 2) 执行完后reset是没用的
	fmt.Println(<-timer4.C)
}
