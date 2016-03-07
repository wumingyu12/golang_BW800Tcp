package BW800Tcp

import (
	"log"
	"net"
	"reflect"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

type BW800Instance struct {
	EquAddr        []byte //设备地址，注意是小端排列如 E7 03 00 00 地址就是0x3e7就是999，
	TcpConnect     *net.TCPConn
	IpAndPort      string      //用来唯一表示这个实例对应的ip和端口
	TheReadMessage []byte      //最近的一次读取到的报文，ReadThread线程中更新
	WriteChan      chan []byte //用于存放将要发送的命令
	ReadChan       chan []byte //用于存放接收到的命令，如果是心跳包，登录包不算会自动回复
	IfOnline       bool        //是否在线标志位，当收到心跳报文和登录报文会赋值为yes
}

/*************************用TCP连接实例化一个结构体**********************************************
	功能描述：
 ****************************************************************************/
func (b *BW800Instance) construct(conn *net.TCPConn) { //用tcp连接初始化实例
	b.IpAndPort = conn.RemoteAddr().String()
	b.TcpConnect = conn
	//初始化发送报文通道为有缓冲为1
	b.WriteChan = make(chan []byte, 1)
	b.ReadChan = make(chan []byte, 1)
	b.IfOnline = false
}

/*************************启动BW800具有的线程**********************************************
	功能描述：
		1.启动一个线程用来不断读取发送过来的报文 ReadThread 函数
		2.启动一个线程用来
 ****************************************************************************/
func (instance *BW800Instance) RunThread() {
	go instance.ReadThread()
	go instance.WriteThread() //发送缓存里面的命令

}

/******************************报文发送线程****************************************
	功能描述：
	1.	发送结构体里面的WriteChan 报文
	2.启动线程时要等待在线标志位ok才启动，当收到登录报文和心跳报文会确定为ok
 *****************************************************************************/
func (instance *BW800Instance) WriteThread() {
	for {
		if instance.IfOnline { //启动线程时要等待在线标志位ok才启动
			break
		}
	}
	ipStr := instance.TcpConnect.RemoteAddr().String()
	log.Printf("启动 %s 发送报文线程\n", ipStr)
	//mes := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x02, 0x2E}
	//instance.WriteChan <- mes
	for {
		tempmsg := <-instance.WriteChan
		log.Printf("服务器发送报文：%x 到 %s\n", tempmsg, ipStr)
		instance.TcpConnect.Write(tempmsg)
		time.Sleep(time.Second * 1) //避免和另一个线程的write粘包，见文档  golang 粘包的问题
	}
}

/*************************报文读取线程函数**********************************************
	功能描述：
	1.用BW800Instance结构体中 TcpConnect *net.TCPConn 的TCP连接进行读取与连接
	2.循环读取tcp中的数据,获取一条报文
	4.将获取到的报文，交给报文处理函数messageHandle处理。
 ****************************************************************************/
func (instance *BW800Instance) ReadThread() { //启动线程用来接收数据

	ipStr := instance.TcpConnect.RemoteAddr().String()
	log.Printf("启动 %s 接收报文线程\n", ipStr)
	defer func() {
		log.Println("disconnected :" + ipStr)
		instance.TcpConnect.Close()
	}()

	//mes := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x02, 0x2E}
	//instance.TcpConnect.Write(mes)

	b := make([]byte, 128)
	//计时

	//循环一直检测发送的数据
	for {
		i, err := instance.TcpConnect.Read(b)
		log.Printf("从 %s 收到结果：%x\n", ipStr, b[0:i])
		messageHandle(instance, b[0:i])
		if err != nil {
			log.Printf("%sTcp读取错误\n", ipStr)
		}
	}
}

/*************************报文处理函数********************************************
	调用它的函数：
		1.ReadThread，报文获取函数
	依赖的函数：
		1.sumCheck计算和校验
	存入参数：
		1.BW800的结构体
		2.要处理的接收到的报文
	功能：
		根据获取的报文，进行处理。
		1.如果为登录包，就回复登录确认报文，并退出该函数
		2.如果为心跳包就回复任意东西（并用心跳报文里面的设备地址赋值到结构体中的地址变量里面）并退出该函数。
		3.如果登录包和心跳包都不是的话就是一条响应回复报文,我们将该信息放到结构体中的ReadChan管道中。
	登录报文：
			8A 9B 02 00 00 00 00 0A 10 00 00 05 02 00 00 00 00 48
			|   登录头 |类型 |    设备地址          | 长度 |报文|                        |设备地址              |和校验 |
			0x8a, 0x9b, 0x02, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x10, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x00, 0x00, 0x48
			服务器回复
			        | 设备地址  |
			8A 9B 02 00 00 00 00 06 90 00 00 01 00 BE
	登录成功后会发送心跳报文(每分钟一条)
			|包头|   |设备地址  |                 |校验
			8A 9B 02 00 00 00 00 06 10 01 00 01 00 3F
			服务器回复
			0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x90, 0x01, 0x00, 0x01, 0x00, 0xBF
 ******************************************************************/
func messageHandle(instance *BW800Instance, msg []byte) {
	//设置一条登录报文例子
	//接收到的登录报文
	var logingMessageExample = []byte{0x8a, 0x9b, 0x02, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x10, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x00, 0x00, 0x48}
	if len(msg) == len(logingMessageExample) { //如果这条信息与登录示范报文的长度一样，可以进一步判断是否为登录报文
		eq1 := reflect.DeepEqual(logingMessageExample[0:3], msg[0:3])   //判断第一处地方是否相等就是头与类型码是否相等
		eq2 := reflect.DeepEqual(logingMessageExample[7:13], msg[7:13]) //判断第二处地方是否相等就是寄存器地址
		if eq1 && eq2 {
			confirmExample := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x90, 0x00, 0x00, 0x01, 0x00, 0xBE}
			//登录包回复组包
			confirmLogin := append(confirmExample[0:3], msg[3:7]...) //在报文头后面加上设备地址
			confirmLogin = append(confirmLogin, confirmExample[7:13]...)
			confirmLogin = append(confirmLogin, sumCheck(confirmLogin))

			instance.TcpConnect.Write(confirmLogin)
			time.Sleep(time.Second * 1) //避免和另一个线程的write粘包，见文档  golang 粘包的问题
			log.Printf("服务器回复登录报文：%x\n", confirmLogin)
			instance.IfOnline = true //确认在线

			return
		}
	}
	//如果是一条心跳报文
	var pollingMessageExample = []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x01, 0x00, 0x3F}
	if len(msg) == len(pollingMessageExample) { //如果这条信息与心跳示范报文的长度一样，可以进一步判断是否为心跳报文
		eq1 := reflect.DeepEqual(pollingMessageExample[0:3], msg[0:3])   //判断第一处地方是否相等就是头与类型码是否相等
		eq2 := reflect.DeepEqual(pollingMessageExample[7:13], msg[7:13]) //判断第一处地方是否相等就是头与类型码是否相等
		if eq1 && eq2 {
			confirmPollingExample := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x90, 0x01, 0x00, 0x01, 0x00, 0xBF}
			//登录包回复组包
			confirmPolling := append(confirmPollingExample[0:3], msg[3:7]...) //在报文头后面加上设备地址
			confirmPolling = append(confirmPolling, confirmPollingExample[7:13]...)
			confirmPolling = append(confirmPolling, sumCheck(confirmPolling)) //加上和校验

			instance.TcpConnect.Write(confirmPolling) //发送心跳回复报文
			time.Sleep(time.Second * 1)               ////避免和另一个线程的write粘包，见文档  golang 粘包的问题
			log.Printf("服务器回复心跳报文：%x\n", confirmPolling)
			//将心跳包里面的设备地址赋值到结构体中
			instance.EquAddr = msg[3:7]
			instance.IfOnline = true //确认在线
			return
		}
	}
	//如果都不是心跳包和登录包可能就是一条用户发送协议后的回复响应
	instance.ReadChan <- msg
	//fmt.Printf("%x\n", <-instance.ReadChan)
}
