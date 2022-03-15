package main

import (
	"errors"
	"fmt"
	"sync"
)

func main() {
	TestNew()
	TestV()
	TestW()
}

type Singleton struct {
}

var once sync.Once

var instance *Singleton

func GetInstance() *Singleton {
	once.Do(func() {
		instance = &Singleton{}
	})
	return instance
}

func TestNew() {
	err := errors.New("")
	if err != nil {
		fmt.Printf("error")
	}
}

func TestV() {
	err := errors.New("error")
	basic := fmt.Errorf("ctx: %v", err) // 没有实现 Unwrap 接口
	if _, ok := basic.(interface{ Unwrap() error }); !ok {
		fmt.Println("errorString")
	}
}

func TestW() {
	err := errors.New("error")
	wrap := fmt.Errorf("ctx: %w", err)
	if _, ok := wrap.(interface{ Unwrap() error }); ok {
		fmt.Println("wrapError")
	}
}
