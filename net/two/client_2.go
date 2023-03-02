package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func encode(msg string) ([]byte, error) {
	// int16类型（占2个字节）
	var len = int16(len(msg))
	var buf = new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, len)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte(msg))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	dial, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("dial err: ", err)
		return
	}
	defer dial.Close()
	fmt.Println("client start -> ")
	for i := 0; i < 10; i++ {
		msg := `hello, starry~`
		real, err := encode(msg)
		if err != nil {
			fmt.Println("Encode err: ", err)
			return
		}
		dial.Write(real)
	}
}
