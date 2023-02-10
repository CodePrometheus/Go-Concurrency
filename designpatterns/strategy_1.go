package main

import "fmt"

// 策略模式
func main() {
	ctx := Context{}
	ctx.ConcreteStrategy(new(Cat))
	ctx.Action()

	ctx.ConcreteStrategy(new(Dog))
	ctx.Action()
}

type Context struct {
	strategy MethodStrategy
}

func (ctx *Context) ConcreteStrategy(strategy MethodStrategy) {
	ctx.strategy = strategy
}

func (ctx *Context) Action() {
	ctx.strategy.Say()
}

type MethodStrategy interface {
	Say()
}

type Cat struct{}

func (cat *Cat) Say() {
	fmt.Println("Cat Say...")
}

type Dog struct{}

func (dog *Dog) Say() {
	fmt.Println("Dog Say...")
}
