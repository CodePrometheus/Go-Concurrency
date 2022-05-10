package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// 解决 tcp 粘包问题
// 封包：给数据加上包头，数据包分为包头和包体，根据包头中的固定长度变量
func main() {
	con, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Listen err: ", err)
		return
	}
	defer con.Close()
	for {
		conn, err := con.Accept()
		if err != nil {
			fmt.Println("accept err: ", err)
			continue
		}
		go process(conn)
	}
}

func process(con net.Conn) {
	defer con.Close()
	reader := bufio.NewReader(con)
	for {
		msg, err := decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("decode err: ", err)
			return
		}
		fmt.Println("recv: ", msg)
	}
}

func decode(read *bufio.Reader) (string, error) {
	n, _ := read.Peek(2)
	buf := bytes.NewBuffer(n)
	var len int16
	// 对于大部分的 CPU 来说，字节都是按照 Little 小字端形式存储的  0x55 0x44 0x33 0x22
	err := binary.Read(buf, binary.LittleEndian, &len)
	if err != nil {
		return "", err
	}
	if int16(read.Buffered()) < len+2 {
		return "", err
	}
	real := make([]byte, int(2+len))
	_, err = read.Read(real)
	if err != nil {
		return "", err
	}
	return string(real[2:]), nil
}
