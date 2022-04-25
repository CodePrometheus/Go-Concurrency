// 生产者消费者模型之中转物流
package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 定义producer
func producer(chanStorage chan int) {
	for i := 0; i < 10; i++ {
		product := rand.Intn(1000)
		chanStorage <- product
		fmt.Println("生产了商品", product)
	}
	close(chanStorage)
}

// 定义中转效果
func mid(pro chan int, con chan int) {
	for p := range pro {
		con <- p
		fmt.Println("完成了中转托运", p)
	}
	fmt.Println("商品转运完成! 商店关闭!")
	close(con)
}

// 定义消费者
func consumer(chanSell chan int) {
	for c := range chanSell {
		fmt.Println("消费了商品", c)
	}
	fmt.Println("商品已经全部消费完成！")
}

func main() {
	// 定义中转物流仓库,定义存储100个上商品
	chanStorage := make(chan int, 100)
	// 定义消费渠道，由消费者进行消费
	chanSell := make(chan int, 100)

	// 制造商品,producer
	go producer(chanStorage)

	// 进行中转托运
	go mid(chanStorage, chanSell)

	// 进行售卖商品, consumer
	go consumer(chanSell)

	// 主流程
	for {
		time.Sleep(time.Second)
	}
}
