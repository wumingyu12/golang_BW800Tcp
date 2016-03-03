package BW800

import (
	"fmt"
	"net"
	"reflect"
	"time"
)

type BW800Instance struct {
	TcpConnect      *net.TCPConn
	IpAndPort       string //用来唯一表示这个实例对应的ip和端口
	TheReadMessage  []byte //最近的一次读取到的报文，ReadThread线程中更新
	TheWriteMessage chan []byte
}

/*************************用TCP连接实例化一个结构体**********************************************
	功能描述：
 ****************************************************************************/
func (b *BW800Instance) construct(conn *net.TCPConn) { //用tcp连接初始化实例
	b.IpAndPort = conn.RemoteAddr().String()
	b.TcpConnect = conn
	//初始化发送报文通道为有缓冲为1
	b.TheWriteMessage = make(chan []byte, 1)
}

/*************************启动BW800具有的线程**********************************************
	功能描述：
		1.启动一个线程用来不断读取发送过来的报文 ReadThread 函数
		2.启动一个线程用来
 ****************************************************************************/
func (instance *BW800Instance) RunThread() {
	go instance.ReadThread()
	//go instance.WriteThread() //发送缓存里面的命令
}

/*************************报文读取线程函数**********************************************
	功能描述：
	1.用BW800Instance结构体中 TcpConnect *net.TCPConn 的TCP连接进行读取与连接
	2.循环读取tcp中的数据，并用判断超时的方式来判断报文尾部，获取一条报文
	3.将读取到的报文，放到结构体中的TheReadMessage中
	4.将获取到的报文，交给报文处理函数messageHandle处理。
 ****************************************************************************/
func (instance *BW800Instance) ReadThread() { //启动线程用来接收数据
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
			fmt.Println("请求超时" + ipStr)
			timeoutFlag1 = timeoutFlag2 //将超时标志赋值
			timeoutFlag2 = true         //设置当前超时标志为yes,代表当前为超时
			//如果这次超时是读取到报文尾部导致的。
			//通过2个标志位判断当前读取的比特是否为报文尾,如果标志1超时没有超时并且标记2超时了，就代表读到报文尾部
			if (timeoutFlag1 == false) && (timeoutFlag2 == true) {
				fmt.Printf("结果：%x\n", result)
				//将获取到的报文缓存到结构体中
				instance.TheReadMessage = result
				//运行报文处理函数
				messageHandle(instance)
			}
			result = []byte{} //清空报文缓存
		} else { //如果没有超时，就将读取到的结果加到报文缓存中
			result = append(result, b[0])
			timeoutFlag1 = timeoutFlag2 //将超时标志赋值
			timeoutFlag2 = false        //设置当前超时标志为no,代表当前为没有超时
		}
	}
}

/******************************报文发送线程****************************************
	功能描述：
	1.	发送结构体里面的TheWriteMessage 报文
 *****************************************************************************/
func (instance *BW800Instance) WriteThread() {
	fmt.Println("启动发送命令线程")
	mes := <-instance.TheWriteMessage
	instance.TcpConnect.Write(mes)
}

/*************************报文处理函数********************************************
	调用它的函数：
		1.ReadThread，报文获取函数

	功能：
		根据获取的报文，进行处理。
		1.如果为登录包，就回复登录确认报文。
		2.如果为心跳包就回复任意东西。
		3.如果登录包和心跳包都不是的话就是一条响应回复报文。
	登录报文：
			8A 9B 02 00 00 00 00 0A 10 00 00 05 02 00 00 00 00 48
			|   登录头 |类型 |    设备地址          | 长度 |报文|                        |设备地址              |和校验 |
			0x8a, 0x9b, 0x02, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x10, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x00, 0x00, 0x48
			服务器回复
			        | 设备地址  |
			8A 9B 02 00 00 00 00 06 90 00 00 01 BE
	登录成功后会发送心跳报文(每分钟一条)
			|包头|   |设备地址  |                 |校验
			8A 9B 02 00 00 00 00 06 10 01 00 01 00 3F
 ******************************************************************/
func messageHandle(instance *BW800Instance) {
	//设置一条登录报文例子
	var logingMessageExample = []byte{0x8a, 0x9b, 0x02, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x10, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x00, 0x00, 0x48}
	if len(instance.TheReadMessage) == len(logingMessageExample) { //如果这条信息与登录示范报文的长度一样，可以进一步判断是否为登录报文
		eq1 := reflect.DeepEqual(logingMessageExample[0:3], instance.TheReadMessage[0:3])   //判断第一处地方是否相等就是头与类型码是否相等
		eq2 := reflect.DeepEqual(logingMessageExample[7:13], instance.TheReadMessage[7:13]) //判断第二处地方是否相等就是寄存器地址
		if eq1 && eq2 {
			fmt.Println("11") //如果两处都相等就认为是一条登录报文
			confirmLogin := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x90, 0x00, 0x00, 0x01, 0x00, 0xBE}
			instance.TcpConnect.Write(confirmLogin) //
		}
	}
	//如果是一条心跳报文
	var pollingMessageExample = []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x01, 0x00, 0x3F}
	if len(instance.TheReadMessage) == len(pollingMessageExample) { //如果这条信息与心跳示范报文的长度一样，可以进一步判断是否为心跳报文
		eq1 := reflect.DeepEqual(pollingMessageExample[0:3], instance.TheReadMessage[0:3])   //判断第一处地方是否相等就是头与类型码是否相等
		eq2 := reflect.DeepEqual(pollingMessageExample[7:13], instance.TheReadMessage[7:13]) //判断第一处地方是否相等就是头与类型码是否相等
		if eq1 && eq2 {
			fmt.Println("22222")
			confirmPolling := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x90, 0x01, 0x00, 0x01, 0x00, 0xBF}
			instance.TcpConnect.Write(confirmPolling) //
		}
	}
}
