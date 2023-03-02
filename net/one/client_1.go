package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func client() {
	con, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("dial err: ", err)
		return
	}
	defer con.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		res := strings.Trim(input, "\r\n")
		if strings.ToUpper(res) == "Q" {
			return
		}
		_, err := con.Write([]byte(input))
		if err != nil {
			return
		}
		buf := [512]byte{}
		read, err := con.Read(buf[:])
		if err != nil {
			fmt.Println("read err: ", err)
			return
		}
		fmt.Println(string(buf[:read]))
	}
}

func main() {
	con, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("dial: ", err)
		return
	}
	defer con.Close()
	fmt.Println("client start -> ")
	// 模拟粘包
	// tcp 数据传递模式是流式的，在保持长连接的时候可以进行多次的收和发
	for i := 0; i < 10; i++ {
		msg := `hello, starry~`
		con.Write([]byte(msg))
	}
	fmt.Println("send over -> ")
}
