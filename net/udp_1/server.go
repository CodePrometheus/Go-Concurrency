package main

import (
	"fmt"
	"net"
)

func main() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8080,
	})
	if err != nil {
		fmt.Println("ListenUDP err: ", err)
		return
	}
	defer con.Close()
	for {
		var data [1024]byte
		// 无 Accept
		n, addr, err := con.ReadFromUDP(data[:])
		if err != nil {
			fmt.Println("ReadFromUDP err: ", err)
			return
		}
		fmt.Printf("data: %v, addr: %v, cnt: %v", string(data[:]), addr, n)
		_, err = con.WriteToUDP(data[:n], addr)
		if err != nil {
			fmt.Println("WriteToUDP err: ", err)
			return
		}
	}
}
