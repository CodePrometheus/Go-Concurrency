package main

import (
	"fmt"
	"time"
)

// Ticker 定时触发的计时器
// type Ticker struct {
//	 C <-chan Time // The channel on which the ticks are delivered. 管道，上层应用跟据此管道接收事件
// 	 r runtimeTimer // 定时器，该定时器即系统管理的定时器，对上层应用不可见
// }
// Timer创建时，不指定事件触发周期，事件触发后Timer自动销毁。
// 而Ticker创建时会指定一个事件触发周期，事件会按照这个周期触发，如果不显式停止，定时器永不停止。
func main() {
	ticker := time.NewTicker(time.Second * 1)
	i := 0
	for {
		<-ticker.C
		i++
		fmt.Printf("i:%d\n", i)
		if i == 6 {
			ticker.Stop()
			break
		}
	}

}
