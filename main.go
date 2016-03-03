package main

import (
	"./BW800"
	"fmt"
	"net"
)

//常量
const (
	IP   = "10.33.51.107" //服务器地址
	PORT = "9999"         //服务端口
)

func main() {
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
		var mycontain = &BW800.Bw800Container{}
		mycontain.AddBW800(tcpConn)
		fmt.Println(mycontain)
		//go tcpPipe(tcpConn)
	}

}
