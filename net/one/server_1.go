package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

func server(con net.Conn) {
	defer con.Close()
	for {
		reader := bufio.NewReader(con)
		var buf [256]byte
		read, err := reader.Read(buf[:])
		if err != nil {
			fmt.Println("reader.Read err: ", err)
			break
		}
		recv := string(buf[:read])
		fmt.Println("receive: ", recv)
		con.Write([]byte(recv))
	}
}

func process(con net.Conn) {
	defer con.Close()
	read := bufio.NewReader(con)
	var buf [1024]byte
	for {
		n, err := read.Read(buf[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read err: ", err)
			break
		}
		recv := string(buf[:n])
		fmt.Printf("recv: %s\n\n", recv)
	}
}

func main() {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Listen err: ", err)
		return
	}
	defer listen.Close()
	fmt.Println("server start -> ")
	for {
		con, err := listen.Accept()
		if err != nil {
			fmt.Println("Accept err: ", err)
			continue
		}
		go process(con)
	}
}
