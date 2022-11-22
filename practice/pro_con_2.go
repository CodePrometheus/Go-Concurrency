package main

import (
	"fmt"
	"sync"
)

const maxCnt = 3
const minCnt = 0

type iceCube int
type cup struct {
	iceCubes []iceCube
}

// 使用 sync.Cond 解决 生产者消费者 问题
// 类比Java 加锁 + 循环&&等待 唤醒 在 Go 中也是 加锁 + 循环&&等待 经典范式
// 不加锁情况下， 如果生产冰块同时时还能从杯子中拿出冰块
// 万一生产速率 < 拿取速率， 杯子就空了, 反之，杯子溢出冰块，都是非预期的情况
func main() {
	stopCh := make(chan struct{})

	lc := new(sync.Mutex)
	cond := sync.NewCond(lc)
	cup := cup{
		iceCubes: make([]iceCube, 3, 3),
	}

	// producer
	go func() {
		for {
			cond.L.Lock()
			for len(cup.iceCubes) == maxCnt {
				cond.Wait()
			}
			// 杯子中新添加进一个冰块
			cup.iceCubes = append(cup.iceCubes, 1)
			fmt.Println("producer 1 ice, left iceCubes: ", len(cup.iceCubes))
			cond.Signal()
			cond.L.Unlock()
		}
	}()

	// consumer
	go func() {
		for {
			cond.L.Lock()
			for len(cup.iceCubes) == minCnt {
				cond.Wait()
			}
			// 删除头部的冰块
			cup.iceCubes = cup.iceCubes[1:]
			fmt.Println("consume 1 ice, left iceCubes: ", len(cup.iceCubes))
			cond.Signal()
			cond.L.Unlock()
		}
	}()

	for {
		select {
		case <-stopCh:
			return
		default:
		}
	}
}
