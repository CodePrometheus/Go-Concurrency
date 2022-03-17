package main

import "fmt"

type M[k string, v any] map[k]v // k不支持any，由于底层map不支持，所以使用string

type C[T any] chan T

func main() {
	m1 := M[string, int]{"key": 1}
	m1["key"] = 2

	m2 := M[string, string]{"key": "value"}
	m2["key"] = "new value"
	fmt.Println(m1, m2)

	c1 := make(C[int], 10)
	c1 <- 1
	c1 <- 2

	c2 := make(C[string], 10)
	c2 <- "hello"
	c2 <- "world"

	fmt.Println(<-c1, <-c2)
	fmt.Println(<-c1, <-c2)
}
