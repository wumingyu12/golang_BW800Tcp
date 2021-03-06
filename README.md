智能母猪饲喂器BW800 羊头测试wifi连接服务器
===========================================

[有道的相关笔记 有通信协议与例子](http://note.youdao.com/share/?id=2653313eba0d99860e54722c7ff45291&type=note)

[go 使用tcp的例子](http://note.youdao.com/share/?id=4069cd11a0ae1495a79e8a62f689470b&type=note)

[~~go tcp中判断报文尾部获取一条报文~~](http://note.youdao.com/share/?id=56d0dcacf0a823701042b4addb6ecd42&type=note)

事实上tcp包你是不需要判断哪个是包尾的，他发送都是一次一个包这样的
```go
	for {
		i, err := instance.TcpConnect.Read(b)
		log.Printf("从 %s 收到结果：%x\n", ipStr, b[0:i])
		messageHandle(instance, b[0:i])
		if err != nil {
			log.Printf("%sTcp读取错误\n", ipStr)
		}
	} 
```

[判断两个比特数组是否相等](http://note.youdao.com/share/?id=46e0bb9570c6b0b72caa1e72605b0ef8&type=note)

[golang 控制台打印替代前面的内容](http://note.youdao.com/share/?id=d7d9272cf0e8ff26dd43cdb1f7242aba&type=note)

[golang 粘包的问题](http://note.youdao.com/share/?id=bf7107840bba285aa16b8e6f81222113&type=note)

[使用odoo中的xml-rpc](http://note.youdao.com/share/?id=86d4a757bae17096cd6913597475c3dd&type=note)

[golang 调用odoo xml-rpc](http://note.youdao.com/share/?id=179e06e1c98253938293fc7970e3f8c9&type=note)
 工程目录testexample里面有个例子

[go 实例化结构体里面的数组](http://note.youdao.com/share/?id=8ddac16590c1aebc47854dacf33defb4&type=note)

[如果你想查看发给odoo的xml是怎么的数据](http://note.youdao.com/share/?id=6ee1eb2d5fda01a1242231584578868d&type=note)

[odoo 调用xml rpc的问题](http://note.youdao.com/share/?id=b36461cf93c2da90282f3dce5647a0e6&type=note)

[golang  post 请求服务器 可以用来查看xml-rpc的请求 go请求xml-rpc](http://note.youdao.com/share/?id=562cb25ea3a51b309ef28fa3a0920fb8&type=note)

总体架构说明
-----------
1. BW800用的wifi模块为esp8266

2. BW800作为客户端，输入服务器端的ip 端口号(9999)。



主要用法
-------------
```go
	fmt.Println("A client connected : " + tcpConn.RemoteAddr().String())
	var mycontain = &BW800Tcp.Bw800Container{}
	mycontain.AddBW800(tcpConn)
	mes := []byte{0x8A, 0x9B, 0x02, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x02, 0x2E}
	mycontain.BW800S[0].WriteChan <- mes
	fmt.Println(<-mycontain.BW800S[0].ReadChan)
```
一个tcp连接过来后创建一个实例并加到一个容器内，该实例会自动回复心跳包和登录包，如果要发送一些协议报文，只要向这个实例里面的WriteChan写入要发送的报文
在ReadChan里面接收返回的报文就行


连接流程
------------
1. BW800里面输入服务器ip 端口号。

2. 服务器会收到BW800发过来的登录报文，服务器必须回复该报文，如果不回复3次，BW800会重新发起一个tcp连接。

3. BW800收到登录回复报文后，不再发送登录报文，开始发送心跳包，心跳包也要回复，回复内容可以为任意内容。如果3次没回复后会重新发送登录报文。在这里是收到心跳包后直接发送命令，更新结构体

4. 有一个定时任务（每次收到心跳包后运行），调用odoo模块的函数将数据存储到odoo的数据库中。

5. 心跳包启动的定时任务中，每一次读取10条记录，详细记录3200条最多，那么要读320次重复发送后一个循环，如果是日记录64条最多，那么也是如此。如果读到剩余记录了那么就重置，重第1条再次读取再读到最后一条。


类与结构体
------------
### 包BW800
* 结构体 容器类BW800Container.go
	用来存放BW800的实例BW800Instance.go，具有方法

```go
type Bw800Container struct {
	BW800S []*BW800Instance //存放BW800的每一个连接实例
}

func (b *Bw800Container) AddBW800(conn *net.TCPConn) { //向容器中添加BW800的实例
```

* 结构体 BW800Instance
	成员变量
	WriteChan      chan []byte //用于存放将要发送的命令
	ReadChan       chan []byte //用于存放接收到的命令，如果是心跳包，登录包不算会自动回复

	将要发送的命令放到WriteChan中就可以在ReadChan中读取到返回的包

与odoo的xml-rpc通信
-------------------
1.将要发送的结构体转为json发送出去。

2.go要发送过去的json格式里面的变量类型应该与odoo数据库里面的类型保持一致