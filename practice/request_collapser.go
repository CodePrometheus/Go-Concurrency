package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Msg struct {
	key     string
	content string
}

type Context struct {
	R   *http.Request
	W   http.ResponseWriter
	out chan<- Msg
}

var requests []Context

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var req Context
	req.R = r
	req.W = w
	ch := make(chan Msg, 0) // 申请管道，等待业务数据结果
	req.out = ch
	requests = append(requests, req)
	msg := <-ch // 等待业务数据
	fmt.Fprint(w, msg.key+"_"+msg.content)
}

func batchHandle(s []Context) {
	size := len(s)
	for i := 0; i < size; i++ {
		fmt.Println(s[i].R.URL)
		var msg Msg
		msg.content = "success"
		msg.key = "i: " + strconv.Itoa(i)
		s[i].out <- msg // 将处理好的数据请求，写入对应管道
	}
}

// 收集两秒内接口请求，进行合并请求，批量处理数据，然后分发数据
func task() {
	ticker := time.NewTicker(time.Second * 2)
	for {
		// 开启定时，将收集的请求，放置在新容器中
		<-ticker.C
		var r = requests
		var rNew []Context
		// 清空旧容器数据
		requests = rNew
		batchHandle(r)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	go task()
	log.Println(http.ListenAndServe(":8080", nil))
}
