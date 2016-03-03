package BW800

import (
	"net"
)

type Bw800Container struct {
	BW800S []*BW800Instance //存放BW800的每一个连接实例
}

func (b *Bw800Container) AddBW800(conn *net.TCPConn) { //向容器中添加BW800的实例
	var instance = &BW800Instance{}
	instance.construct(conn) //实例化该tcp连接
	instance.RunThread()     //启动实例里面的线程
	b.BW800S = append(b.BW800S, instance)
}
