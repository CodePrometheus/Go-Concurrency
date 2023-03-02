package main

import (
	"io"
	"net/http"
	"sync"
	"time"
)

type RequestLimitService struct {
	Interval time.Duration
	MaxCount int
	Lock     sync.Mutex
	ReqCount int
}

func NewRequestLimitService(interval time.Duration, maxCnt int) *RequestLimitService {
	reqLimit := &RequestLimitService{
		Interval: interval,
		MaxCount: maxCnt,
	}
	go func() {
		ticker := time.NewTicker(interval)
		for {
			<-ticker.C
			reqLimit.Lock.Lock()
			reqLimit.ReqCount = 0
			reqLimit.Lock.Unlock()
		}
	}()

	return reqLimit
}

func (reqLimit *RequestLimitService) Increase() {
	reqLimit.Lock.Lock()
	defer reqLimit.Lock.Unlock()

	reqLimit.ReqCount += 1
}

func (reqLimit *RequestLimitService) IsAvailable() bool {
	reqLimit.Lock.Lock()
	defer reqLimit.Lock.Unlock()

	return reqLimit.ReqCount < reqLimit.MaxCount
}

func main() {
	// 10s 最大请求数 5
	var RequestLimit = NewRequestLimitService(10*time.Second, 2)
	helloHandler := func(w http.ResponseWriter, r *http.Request) {
		if RequestLimit.IsAvailable() {
			RequestLimit.Increase()
			io.WriteString(w, "Hello world!\n")
		} else {
			io.WriteString(w, "Reach request limit!\n")
		}
	}
	http.HandleFunc("/", helloHandler)
	http.ListenAndServe(":8000", nil)
}
