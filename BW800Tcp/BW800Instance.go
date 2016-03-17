package BW800Tcp

import (
	"../OdooRpc"
	"encoding/json"
	"fmt"
	"github.com/wumingyu12/golang_YuGoTool"
	"log"
	"net"
	"reflect"
	"time"
)

const (
	DEBUG_HAVELOGIN = false //调试模式,Bw800已经登录了
	STRING_LONG     = 10    //地址，猪耳标号补零后的总位数，如10 将补为0000000010
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

type BW800Instance struct {
	EquAddr               []byte //设备地址，注意是小端排列如 E7 03 00 00 地址就是0x3e7就是999，
	TcpConnect            *net.TCPConn
	IpAndPort             string      //用来唯一表示这个实例对应的ip和端口
	TheReadMessage        []byte      //最近的一次读取到的报文，ReadThread线程中更新
	WriteChan             chan []byte //用于存放将要发送的命令
	WriteExplainChan      chan string //对应这个个命令的说明，用来打印日志
	ReadChan              chan []byte //用于存放接收到的命令，如果是心跳包，登录包不算会自动回复
	IfOnline              bool        //是否在线标志位，当收到心跳报文和登录报文会赋值为yes
	ParaStruct            *Bw800Para  //结构体对应每个控制器的参数结构体
	PostDayRecordStartNum int16       //日下料记录读取从这里开头作为上传的第一条，每一次心跳触发的任务都会更新
}

/*************************用TCP连接实例化一个结构体**********************************************
	功能描述：
 ****************************************************************************/
func (this *BW800Instance) construct(conn *net.TCPConn) { //用tcp连接初始化实例
	this.IpAndPort = conn.RemoteAddr().String()
	this.TcpConnect = conn
	//初始化发送报文通道为有缓冲为1
	this.WriteChan = make(chan []byte, 1)
	this.ReadChan = make(chan []byte, 1)
	this.WriteExplainChan = make(chan string, 1)
	if DEBUG_HAVELOGIN {
		this.IfOnline = true
	} else {
		this.IfOnline = false
	}
	this.ParaStruct = &Bw800Para{}
	this.PostDayRecordStartNum = 1 //一开始是从第一条开始读
}

/*************************启动BW800具有的线程**********************************************
	功能描述：
		1.启动一个线程用来不断读取发送过来的报文 ReadThread 函数
		2.启动一个线程用来
 ****************************************************************************/
func (this *BW800Instance) RunThread() {
	go this.ReadThread()
	go this.WriteThread() //发送缓存里面的命令
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
		log.Printf("服务器发送报文-%s：%x 到 %s\n", <-instance.WriteExplainChan, tempmsg, ipStr)
		instance.TcpConnect.Write(tempmsg) //这里不加sleep但，有一种很特殊的情况就是心跳包回复时恰好这里也发送，就会粘包
	}
}

/*************************报文读取线程函数**********************************************
	功能描述：
	1.用BW800Instance结构体中 TcpConnect *net.TCPConn 的TCP连接进行读取与连接
	2.循环读取tcp中的数据,获取一条报文
	3.将获取到的报文，交给报文处理函数messageHandle处理。
 ****************************************************************************/
func (this *BW800Instance) ReadThread() { //启动线程用来接收数据

	ipStr := this.TcpConnect.RemoteAddr().String()
	log.Printf("启动 %s 接收报文线程\n", ipStr)
	defer func() {
		log.Println("disconnected :" + ipStr)
		this.TcpConnect.Close()
	}()

	//mes := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x02, 0x2E}
	//this.TcpConnect.Write(mes)

	b := make([]byte, 256)

	//计时

	//循环一直检测发送的数据
	for {
		i, err := this.TcpConnect.Read(b)
		log.Printf("从 %s 收到结果：%x\n", ipStr, b[0:i])
		messageHandle(this, b[0:i])
		if err != nil {
			log.Printf("%sTcp读取错误\n", ipStr)
		}
	}
}

/*************************报文处理函数********************************************
	调用它的函数：
		1.ReadThread，报文获取函数
	依赖的函数：
		1.Fun_SumCheck计算和校验
	存入参数：
		1.BW800的结构体
		2.要处理的接收到的报文
	功能：
		根据获取的报文，进行处理。
		1.如果为登录包，就回复登录确认报文，并退出该函数
		2.如果为心跳包就运行一个定时任务PollingTask，理论上你接收到心跳包后回复任何东西都是可以的
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
func messageHandle(this *BW800Instance, msg []byte) {
	//设置一条登录报文例子
	//接收到的登录报文
	var logingMessageExample = []byte{0x8a, 0x9b, 0x02, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x10, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x00, 0x00, 0x48}
	if len(msg) == len(logingMessageExample) { //如果这条信息与登录示范报文的长度一样，可以进一步判断是否为登录报文
		eq1 := reflect.DeepEqual(logingMessageExample[0:3], msg[0:3])   //判断第一处地方是否相等就是头与类型码是否相等
		eq2 := reflect.DeepEqual(logingMessageExample[7:13], msg[7:13]) //判断第二处地方是否相等就是寄存器地址
		if eq1 && eq2 {
			//将包里面的设备地址赋值到结构体中
			this.EquAddr = msg[3:7]
			confirmExample := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x90, 0x00, 0x00, 0x01, 0x00, 0xBE}
			//登录包回复组包
			confirmLogin := append(confirmExample[0:3], msg[3:7]...) //在报文头后面加上设备地址
			confirmLogin = append(confirmLogin, confirmExample[7:13]...)
			confirmLogin = append(confirmLogin, Fun_SumCheck(confirmLogin))

			this.TcpConnect.Write(confirmLogin)
			time.Sleep(time.Second * 1) //避免和另一个线程的write粘包，见文档  golang 粘包的问题
			log.Printf("服务器回复登录报文：%x\n", confirmLogin)
			this.IfOnline = true //确认在线

			return
		}
	}
	//如果是一条心跳报文
	var pollingMessageExample = []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x01, 0x00, 0x3F}
	if len(msg) == len(pollingMessageExample) { //如果这条信息与心跳示范报文的长度一样，可以进一步判断是否为心跳报文
		eq1 := reflect.DeepEqual(pollingMessageExample[0:3], msg[0:3])   //判断第一处地方是否相等就是头与类型码是否相等
		eq2 := reflect.DeepEqual(pollingMessageExample[7:13], msg[7:13]) //判断第一处地方是否相等就是头与类型码是否相等
		if eq1 && eq2 {
			//将心跳包里面的设备地址赋值到结构体中
			this.EquAddr = msg[3:7]
			this.IfOnline = true //确认在线

			// confirmPollingExample := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x06, 0x90, 0x01, 0x00, 0x01, 0x00, 0xBF}
			//心跳包回复组包
			// confirmPolling := append(confirmPollingExample[0:3], msg[3:7]...) //在报文头后面加上设备地址
			// confirmPolling = append(confirmPolling, confirmPollingExample[7:13]...)
			// confirmPolling = append(confirmPolling, Fun_SumCheck(confirmPolling)) //加上和校验

			// this.TcpConnect.Write(confirmPolling) //发送心跳回复报文
			// time.Sleep(time.Second * 1)               ////避免和另一个线程的write粘包，见文档  golang 粘包的问题
			// log.Printf("服务器回复心跳报文：%x\n", confirmPolling)
			log.Println("收到心跳包回复定时任务")
			//运行一次更新任务,为了不影响这个读线程直接go 一个多线程处理
			//注意你如果这里不用go就会造成死锁
			go this.PollingTask()
			return
		}
	}
	//如果都不是心跳包和登录包可能就是一条用户发送协议后的回复响应
	this.ReadChan <- msg
	//fmt.Printf("%x\n", <-this.ReadChan)
}

/*******************************收到心跳包后运行的定时任务*****************************************
	1.更新参数结构体ParaStruct
	2.将要更新后的结构体转为json，通过xml-rpc发送到odoo，存入数据库
*************************************************************************/
func (this *BW800Instance) PollingTask() {
	this.PostParaStruct()
	this.PostDayRecord(this.PostDayRecordStartNum, 10)
}

/*************************************************************************
	将实例中的参数结构体推送到odoozhong
***************************************************************************/
func (this *BW800Instance) PostParaStruct() {
	//1.更新参数结构体ParaStruct
	//组包                              |类型 | 地址                 | 长度 |C   |起始地址    |D长度|
	getParaExample := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0xB8, 0xE4}
	getParaMsg := append(getParaExample[0:3], this.EquAddr...) //在报文头后面加上设备地址
	getParaMsg = append(getParaMsg, getParaExample[7:12]...)
	getParaMsg = append(getParaMsg, Fun_SumCheck(getParaMsg)) //加上和校验

	//发送获取参数结构体命令
	this.WriteExplainChan <- "获取参数结构体"
	this.WriteChan <- getParaMsg
	result := <-this.ReadChan

	//更新参数结构体
	err1 := this.ParaStruct.Reflash(result)
	if err1 != nil {
		log.Println(err1)
	}
	//log.Println(this.ParaStruct)

	ss, _ := json.Marshal(this.ParaStruct)
	log.Println(string(ss))
	//将结构体推送到odoo中
	OdooRpc.PostParaStruct(string(ss))
}

/***************************************************************
	将日下料记录推送
	startList 从哪一条记录开始读起
	listnum 读多少条记录
***************************************************************/
func (this *BW800Instance) PostDayRecord(startList int16, listnum int16) {
	if listnum > 10 { //最多一次多10条
		log.Println("最多只能一次读10条记录")
		return
	}
	//组包                                         |设备地址              |长度| C    |第几条记录|读几条|Cs
	getDayRecordExample := []byte{0x8A, 0x9B, 0x02, 0xe7, 0x03, 0x00, 0x00, 0x05, 0x04, 0x03, 0x00, 0x0a, 0xff}
	getDayRecordMsg := append(getDayRecordExample[0:3], this.EquAddr...)   //在报文头后面加上设备地址
	getDayRecordMsg = append(getDayRecordMsg, getDayRecordExample[7:9]...) //加长度和C
	sbh, sbl := YuGoTool.Int16_to_twobyte(startList)
	getDayRecordMsg = append(getDayRecordMsg, sbl) //从第几条记录开始读
	getDayRecordMsg = append(getDayRecordMsg, sbh) //从第几条记录开始读
	_, snl := YuGoTool.Int16_to_twobyte(listnum)
	getDayRecordMsg = append(getDayRecordMsg, snl)                           //读多少条记录
	getDayRecordMsg = append(getDayRecordMsg, Fun_SumCheck(getDayRecordMsg)) //添加校验和
	//发送获取日记录
	this.WriteExplainChan <- fmt.Sprintf("获取日下料记录,起始条数:%d,连续读取：%d条\n", startList, listnum)
	this.WriteChan <- getDayRecordMsg
	result := <-this.ReadChan

	//log.Println(result)
	//先看下返回的记录条目数
	dlen := int(result[11])
	if dlen == 0 { //如果读取的结果中没有一条有效记录
		this.PostDayRecordStartNum = 1 //让下次读取从1开始
		return
	}
	dayRecordList := RecordStructList{}      //存放的记录结构体的容器
	resultData := Fun_handle_message(result) //去掉头尾只有数据部分
	for i := 0; i < dlen; i++ {
		//log.Printf("%x\n", resultData[(0+16*i):(16+16*i)])
		ds := RecordStruct{}
		//控制器地址 //赋值将这条记录附加一个设备地址
		ds.Addr = fmt.Sprintf("%d", YuGoTool.Fourbyte_to_uint32(this.EquAddr[3], this.EquAddr[2], this.EquAddr[1], this.EquAddr[0]))

		//如10 将补为0000000010
		if len(ds.Addr) < STRING_LONG { //如果长度不足
			var temp string
			for i := 0; i < STRING_LONG-len(ds.Addr); i++ {
				temp = temp + "0"
			}
			ds.Addr = temp + ds.Addr
		}
		err := ds.Reflash(resultData[0+16*i:16+16*i], "day")
		if err != nil {
			log.Println(err)
		}
		dayRecordList.List = append(dayRecordList.List, ds)
	}
	//将结果list 结构体json化发到odoo
	ss, _ := json.Marshal(dayRecordList)
	//将结构体推送到odoo中
	OdooRpc.PostDayRecord(string(ss))

	log.Println(string(ss))

	if dlen < int(listnum) { //如果,读取到的是最后一组记录了
		this.PostDayRecordStartNum = 1 //让下次读取从1开始
	} else {
		this.PostDayRecordStartNum = this.PostDayRecordStartNum + listnum
	}

}
