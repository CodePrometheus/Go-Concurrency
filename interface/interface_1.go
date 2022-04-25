package main

import "fmt"

type Test struct{}

func NilOrNot(v interface{}) bool {
	return v == nil
}
func main() {
	var s *Test
	fmt.Println(s == nil)    // true
	fmt.Println(NilOrNot(s)) // false
}
