package main

import (
	"fmt"
	"math/rand"
	"sync"
)

// 可以加多个读锁或者一个写锁, 写锁优先级大于读锁，为了防止读锁过多, 写锁一直堵塞的情况发生
// 加写锁之前如果有其他的锁（不论是读锁还是写锁）都会阻塞Lock()方法 适于读次数远多于写次数的场景
// 有写锁的时候，会阻塞其他协程读和写, 此时是独占的, 其他goroutine不可获得读或写锁，直达写锁释放
// 只有读锁的时候，允许其他协程加读锁, 但不允许写, 不能获取写锁
// 读写是互斥的，对于同一个资源读和写是不能同时进行的
// 对于同一个协程，先加了写锁就是不能加读写锁了
// 读写锁的底层实现其实是互斥锁加上计数器
// type RWMutex struct {
//	w           Mutex  // 实现写锁之间的阻塞
//	writerSem   uint32 // 阻塞读锁解锁的写锁信号量
//	readerSem   uint32 // 阻塞写锁解锁的读锁信号量
//	readerCount int32  // 阻塞的读锁的个数
//	readerWait  int32  // 当前读锁的个数，用作实现读锁和写锁之间的互斥
//}

var rw sync.RWMutex
var cnt int

func main() {
	ch := make(chan struct{}, 6)
	// 读写同时进行
	for i := 0; i < 3; i++ {
		go ReadCnt(i, ch)
	}
	for i := 0; i < 3; i++ {
		go WriteCnt(i, ch)
	}
	for i := 0; i < 6; i++ {
		<-ch
	}
}

func WriteCnt(id int, ch chan struct{}) {
	rw.Lock()
	fmt.Printf("协程 %v 进入写操作\n", id)
	cnt := rand.Intn(10)
	fmt.Printf("协程 %v 写结束，新值是: %d\n", id, cnt)
	rw.Unlock()
	ch <- struct{}{}
}

func ReadCnt(id int, ch chan struct{}) {
	rw.RLock()
	fmt.Printf("协程 %v 进入读操作\n", id)
	v := cnt
	fmt.Printf("协程 %v 读取结束, 值是: %d\n", id, v)
	rw.RUnlock()
	ch <- struct{}{}
}
