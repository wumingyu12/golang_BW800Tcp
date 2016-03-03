智能母猪饲喂器BW800 羊头测试wifi连接服务器
===========================================

[有道的相关笔记](http://note.youdao.com/share/?id=2653313eba0d99860e54722c7ff45291&type=note)

[go 使用tcp的例子](http://note.youdao.com/share/?id=4069cd11a0ae1495a79e8a62f689470b&type=note)

[go tcp中判断报文尾部获取一条报文](http://note.youdao.com/share/?id=56d0dcacf0a823701042b4addb6ecd42&type=note)

总体架构说明
-----------
1. BW800用的wifi模块为esp8266

2. BW800作为客户端，输入服务器端的ip 端口号(9999)。

3. 将BW800得到的数据放到sqlite数据库中.[sqlite 与go的结合使用](http://note.youdao.com/share/?id=52ad9474de0a5b76ca76928a92ab6e5e&type=note)

连接流程
-------------
1. BW800里面输入服务器ip 端口号。

2. 服务器会收到BW800发过来的登录报文，服务器必须回复该报文，如果不回复3次，BW800会重新发起一个tcp连接。

3. BW800收到登录回复报文后，不再发送登录报文，开始发送心跳包，心跳包也要回复，回复内容可以为任意内容。如果3次没回复后会重新发送登录报文。

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