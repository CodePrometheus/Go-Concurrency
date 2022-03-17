package main

import "fmt"

func PrintSlice[T any](s []T) {
	for _, v := range s {
		fmt.Printf("%v \n", v)
	}
}

func main() {
	PrintSlice[int]([]int{1, 2, 3, 4, 56})
	PrintSlice([]string{"starry", "star"})
}
