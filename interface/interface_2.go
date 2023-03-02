package main

import "fmt"

type Animal interface {
	Eat()
	Move()
	Speak()
}

type Cat struct {
	food       string
	locomotion string
	noise      string
}

func (cat *Cat) Eat() {
	fmt.Println(cat.food)
}

func (cat *Cat) Move() {
	fmt.Println(cat.locomotion)
}

func (cat *Cat) Speak() {
	fmt.Println(cat.noise)
}

func main() {
	var animal Animal
	animal = &Cat{
		food:       "Meat",
		locomotion: "Walk",
		noise:      "Meow",
	}
	animal.Eat()
	animal.Move()
	animal.Speak()
}
