package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// 为什么性能差别如此之大
//
func main() {
	var (
		rw   sync.Mutex
		data = make(map[int]string, 8)
		cnt  int32
	)
	data[0] = "Starry"
	data[1] = "Star"
	data[6] = "Go"

	// write 1
	for i := 0; i < 2; i++ {
		go func(map[int]string) {
			rw.Lock()
			data[1] = fmt.Sprintf("%s%d", "Star", rand.Intn(100))
			time.Sleep(time.Millisecond * 10) // 模拟请求时间
			rw.Unlock()
		}(data)
	}

	// read 100
	for i := 0; i < 100; i++ {
		go func(map[int]string) {
			for {
				rw.Lock()
				time.Sleep(time.Millisecond)
				rw.Unlock()
				atomic.AddInt32(&cnt, 1)
			}
		}(data)
	}

	time.Sleep(time.Second * 3)
	fmt.Println(atomic.LoadInt32(&cnt))
}
