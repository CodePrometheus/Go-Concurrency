package main

import "fmt"

// Fruit 水果类(抽象接口)
type Fruit interface {
	Show() // 接口的某方法
}

// AbstractFactory 工厂类(抽象接口)
type AbstractFactory interface {
	CreateFruit() Fruit // 生产水果类(抽象)的生产器方法
}

type Apple struct {
	Fruit // 为了易于理解显示继承(此行可以省略)
}

func (apple *Apple) Show() {
	fmt.Println("我是苹果")
}

type Banana struct {
	Fruit
}

func (banana *Banana) Show() {
	fmt.Println("我是香蕉")
}

type Pear struct {
	Fruit
}

func (pear *Pear) Show() {
	fmt.Println("我是梨")
}

// AppleFactory 具体的苹果工厂
type AppleFactory struct {
	AbstractFactory
}

func (fac *AppleFactory) CreateFruit() Fruit {
	var fruit Fruit
	// 生产一个具体的苹果
	fruit = new(Apple)
	return fruit
}

// BananaFactory 具体的香蕉工厂
type BananaFactory struct {
	AbstractFactory
}

func (fac *BananaFactory) CreateFruit() Fruit {
	var fruit Fruit

	//生产一个具体的香蕉
	fruit = new(Banana)

	return fruit
}

// PearFactory 具体的梨工厂
type PearFactory struct {
	AbstractFactory
}

func (fac *PearFactory) CreateFruit() Fruit {
	var fruit Fruit
	// 生产一个具体的梨
	fruit = new(Pear)
	return fruit
}

// ======= 业务逻辑层 =======
func main() {
	// 需求1：需要一个具体的苹果对象
	// 1-先要一个具体的苹果工厂
	var appleFac AbstractFactory
	appleFac = new(AppleFactory)
	// 2-生产相对应的具体水果
	var apple Fruit
	apple = appleFac.CreateFruit()

	apple.Show()

	// 需求2：需要一个具体的香蕉对象
	// 1-先要一个具体的香蕉工厂
	var bananaFac AbstractFactory
	bananaFac = new(BananaFactory)
	// 2-生产相对应的具体水果
	var banana Fruit
	banana = bananaFac.CreateFruit()

	banana.Show()

	// 需求3：需要一个具体的梨对象
	// 1-先要一个具体的梨工厂
	var pearFac AbstractFactory
	pearFac = new(PearFactory)
	// 2-生产相对应的具体水果
	var pear Fruit
	pear = pearFac.CreateFruit()

	pear.Show()
}
