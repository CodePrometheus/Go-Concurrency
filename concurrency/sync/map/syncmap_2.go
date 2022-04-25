package main

import (
	"fmt"
	"strconv"
	"sync"
)

var d2 = make(map[string]int)
var w2 sync.WaitGroup

var rw sync.RWMutex

func get(key string) int {
	rw.RLock()
	defer rw.RUnlock()
	return d2[key]
}

func set(key string, value int) {
	rw.Lock()
	defer rw.Unlock()
	d2[key] = value
}

func main() {
	for i := 0; i < 3000; i++ {
		w2.Add(1)
		go func(n int) {
			key := strconv.Itoa(n)
			set(key, n)
			fmt.Printf("k=:%v,v:=%v\n", key, get(key))
			w2.Done()
		}(i)
	}
	w2.Wait()
}
