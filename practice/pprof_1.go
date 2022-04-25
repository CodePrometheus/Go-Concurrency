package main

import (
	"encoding/json"
	"fmt"
	_ "net/http/pprof"
)

var res []string

func main() {
	//go func() {
	//	for {
	//		log.Printf("len:%d", Add("starry"))
	//		time.Sleep(time.Millisecond + 10)
	//	}
	//}()
	//// http://localhost:8989/debug/pprof/
	//_ = http.ListenAndServe(":8989", nil)
	// go tool pprof http://localhost:8989/debug/pprof/profile?seconds=60
	// top 10 查看对应资源开销
	ret := `{"name":"1"}`
	a := &name{
		Name: "aa",
	}
	e := json.Unmarshal([]byte(ret), a)
	fmt.Println(a)
	fmt.Println(e)
}

type name struct {
	Name string
}

func Add(str string) int {
	data := []byte(str)
	res = append(res, string(data))
	return len(res)
}
