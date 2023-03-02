package main

import "fmt"

type Reader interface {
	ReadBook()
}

type Writer interface {
	WriteBook()
}

type Book struct {
}

func (book *Book) ReadBook() {
	fmt.Println("Read Book")
}

func (book *Book) WriteBook() {
	fmt.Println("Write Book")
}

func main() {
	// pair<type:Book, value:"book{}"地址>
	b := &Book{}
	// pair<type:Book, value:"book{}"地址>
	var r Reader
	r = b
	r.ReadBook()

	var w Writer
	w = r.(Writer) // 因为 w r 具体的 type 是 Book，所以可以断言成功
	w.WriteBook()
}
