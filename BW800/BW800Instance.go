package BW800

import (
	"fmt"
	"net"
	"time"
)

type BW800Instance struct {
	TcpConnect *net.TCPConn
	IpAndPort  string //用来唯一表示这个实例对应的ip和端口
}

func (b *BW800Instance) construct(conn *net.TCPConn) { //用tcp连接初始化实例
	b.IpAndPort = conn.RemoteAddr().String()
	b.TcpConnect = conn
}

/***********************************************************************
	功能描述：
	1.用BW800Instance结构体中 TcpConnect *net.TCPConn 的TCP连接进行读取与连接
	2.循环读取tcp中的数据，并用判断超时的方式来判断报文尾部，获取一条报文
	3.
 ****************************************************************************/
func (instance *BW800Instance) startListenThread() { //启动线程用来接收数据
	ipStr := instance.TcpConnect.RemoteAddr().String()
	defer func() {
		fmt.Println("disconnected :" + ipStr)
		instance.TcpConnect.Close()
	}()

	b := []byte{0x00} //一个个字节读取直到超时
	result := []byte{}

	//创建2个变量，代表前一次读取是否超时，和这一次读取是否超时，判断如果前一次没有超时，这次超时了就代表读取到报文尾部了
	var timeoutFlag1 = true //前一次读取，是否超时标志位，默认为1，如果默认为无超时第一次的超时将会判断为报文尾部
	var timeoutFlag2 = true //当前读取是否超时标志位，默认有超时
	//循环一直检测发送的数据
	for {
		//每次就读取一个比特
		instance.TcpConnect.SetReadDeadline(time.Now().Add(time.Second * 1)) //设置读取超时1秒
		_, err := instance.TcpConnect.Read(b)
		//fmt.Printf("%x", b)

		//如果超时，分成2种可能，一种为读到报文尾部了，一种为没有任何报文读取超时
		if err != nil {
			fmt.Println("请求超时")
			timeoutFlag1 = timeoutFlag2 //将超时标志赋值
			timeoutFlag2 = true         //设置当前超时标志为yes,代表当前为超时
			//如果这次超时是读取到报文尾部导致的。
			//通过2个标志位判断当前读取的比特是否为报文尾,如果标志1超时没有超时并且标记2超时了，就代表读到报文尾部
			if (timeoutFlag1 == false) && (timeoutFlag2 == true) {
				fmt.Printf("结果：%x\n", result)
			}
			result = []byte{} //清空报文缓存
		} else { //如果没有超时，就将读取到的结果加到报文缓存中
			result = append(result, b[0])
			timeoutFlag1 = timeoutFlag2 //将超时标志赋值
			timeoutFlag2 = false        //设置当前超时标志为no,代表当前为没有超时
		}
	}
}
