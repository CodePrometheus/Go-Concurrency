package main

import "fmt"

const (
	m, b = iota + 1, iota + 2 // 1 2 | iota = 0
	c, d                      // 2 3
	e, f                      // 3 4

	g, h = iota * 2, iota * 3 // 6 9 | iota = 3
	i, k                      // 8 12
)

func main() {
	fmt.Println(m, b, c, d, e, f)
	fmt.Println(g, h, i, k)
}
