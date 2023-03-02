package main

import (
	"fmt"
	"os"
	"runtime/trace"
)

func main() {
	// go tool trace trace.txt
	file, err := os.Create("./trace.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = trace.Start(file)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hello, World!")
	trace.Stop()
}
