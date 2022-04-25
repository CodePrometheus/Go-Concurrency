package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var cond sync.Cond // 定义全局条件变量

func product(out chan<- int, index int) {
	for {
		// 先加锁
		cond.L.Lock()
		// 判断缓冲区是否满
		for len(out) == 5 {
			// a)阻塞等待条件变量满足
			// b)释放已掌握的互斥锁相当于cond.L.Unlock()。 注意：两步为一个原子操作。
			// c)当被唤醒，Wait()函数返回时，解除阻塞并重新获取互斥锁。相当于cond.L.Lock()
			cond.Wait()
		}
		num := rand.Intn(1000)
		out <- num
		fmt.Printf("生产者%dth,生产:%d\n", index, num)
		// 访问公共区结束，并且打印结束，解锁
		cond.L.Unlock()
		// 唤醒阻塞在条件变量上的 消费者
		cond.Signal()
		time.Sleep(time.Millisecond * 300)
	}
}

func consumer(in <-chan int, index int) {
	for {
		// 先加锁
		cond.L.Lock()
		// 判断 缓冲区是否为空
		for len(in) == 0 {
			// a)阻塞等待条件变量满足
			// b)释放已掌握的互斥锁相当于cond.L.Unlock()。 注意：两步为一个原子操作。
			// c)当被唤醒，Wait()函数返回时，解除阻塞并重新获取互斥锁。相当于cond.L.Lock()
			cond.Wait()
		}
		num := <-in
		fmt.Printf("----消费者%dth,消费:%d\n", index, num)
		// 访问公共区结束后，解锁
		cond.L.Unlock()
		// 唤醒 阻塞在条件变量上的 生产者
		cond.Signal()
		time.Sleep(time.Millisecond * 200)
	}
}

func main() {
	ch := make(chan int, 5)
	quit := make(chan bool)
	rand.Seed(time.Now().UnixNano())
	// 指定条件变量 使用的锁
	cond.L = new(sync.Mutex)
	for i := 0; i < 5; i++ {
		go product(ch, i+1)
	}
	for i := 0; i < 5; i++ {
		go consumer(ch, i+1)
	}
	<-quit
}
