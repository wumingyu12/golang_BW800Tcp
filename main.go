package main

import (
	"./BW800Tcp"
	"fmt"
	"net"
)

//常量
const (
	IP   = "10.33.51.107" //服务器地址
	PORT = "9999"         //服务端口
)

func tcplisten() {
	var tcpAddr *net.TCPAddr

	tcpAddr, _ = net.ResolveTCPAddr("tcp", IP+":"+PORT)

	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)

	defer tcpListener.Close()

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}

		fmt.Println("A client connected : " + tcpConn.RemoteAddr().String())
		var mycontain = &BW800Tcp.Bw800Container{}
		mycontain.AddBW800(tcpConn)
		for i := 0; i < 1000; i++ {
			mes := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x02, 0x2E}
			mycontain.BW800S[0].WriteChan <- mes
			fmt.Println(<-mycontain.BW800S[0].ReadChan)
		}
	}
}

func main() {
	tcplisten()
}
