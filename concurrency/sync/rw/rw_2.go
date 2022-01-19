package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		rw   sync.RWMutex
		data = make(map[int]string, 8)
		cnt  int32
	)
	data[0] = "Starry"
	data[1] = "Star"
	data[6] = "Go"

	// write 1
	for i := 0; i < 2; i++ {
		go func(map[int]string) {
			rw.RLock()
			data[1] = fmt.Sprintf("%s%d", "Star", rand.Intn(100))
			time.Sleep(time.Millisecond * 10) // 模拟请求时间
			rw.RUnlock()
		}(data)
	}

	// read 100
	for i := 0; i < 100; i++ {
		go func(map[int]string) {
			for {
				rw.RLock()
				time.Sleep(time.Millisecond)
				rw.RUnlock()
				atomic.AddInt32(&cnt, 1)
			}
		}(data)
	}

	time.Sleep(time.Second * 3)
	fmt.Println(atomic.LoadInt32(&cnt))
}
