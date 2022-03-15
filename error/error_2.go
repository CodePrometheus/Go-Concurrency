package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	SuccessAssert()
	ErrorAssert()
}

func SuccessAssert() {
	err := &os.PathError{
		Op:   "write",
		Path: "/User/Starry",
		Err:  os.ErrPermission,
	}

	err2 := fmt.Errorf("some context: %w", err)
	var target *os.PathError
	if errors.As(err2, &target) { // 逐层剥离 err2 并检查是否是 os.PathError 类型
		fmt.Printf("operation: %s, path: %s, msg: %v\n", target.Op, target.Path, target.Err)
	}
}

func ErrorAssert() {
	err := &os.PathError{
		Op:   "write",
		Path: "/User/Starry",
		Err:  os.ErrPermission,
	}

	err2 := fmt.Errorf("some context: %w", err)
	if target, ok := err2.(*os.PathError); ok { // 判断失败
		fmt.Printf("operation: %s, path: %s, msg: %v\n", target.Op, target.Path, target.Err)
	}
}
