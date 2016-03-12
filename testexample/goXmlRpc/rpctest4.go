package main

import (
	"bytes"
	"github.com/divan/gorilla-xmlrpc/xml"
	"log"
	"net/http"
)

type Send struct {
	Dbname string //pg数据库
	Uid    int    //用户id
	Pwd    string //用户密码
	Model  string //调用的模块
	Method string //模块方法
	Args   string //参数，可以是数组
}

//xmlRpc的post组包
func XmlRpcCall(method string, args string) (reply struct{ Message string }, err error) {

	send := Send{"11", 1, "1989112", "nmbw800.mymodel", method, args}
	buf, _ := xml.EncodeClientRequest("execute", &send)

	resp, err := http.Post("http://127.0.0.1:8069/xmlrpc/object", "text/xml", bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = xml.DecodeClientResponse(resp.Body, &reply)
	return
}

func main() {
	reply, err := XmlRpcCall("test", "{\"TotalTable\":[320,320,200,200,100,171,242,314,385,457,528,600,600,800,800,800,800,800,800,800],\"TimeTStart\":[0,0,0,0,0,0,0,0,0,0],\"TimeTEnd\":[2359,0,0,0,0,0,0,0,0,0],\"TimeTAmount\":[100,0,0,0,0,0,0,0,0,0],\"CalTable\":[100,0,0,0,0,0,0,0,0,0],\"CanOver\":10,\"XlPer\":100,\"WaterAuto\":0,\"WaterTime\":2,\"WaterSpace\":3,\"UsePass\":0,\"Addr\":2000001000,\"RouteName\":666666,\"RoutePass\":12345678,\"ServerIP\":[10,33,51,107],\"ServerPort\":9999,\"Stage\":1,\"Day\":3,\"HasEat\":0,\"EatDelay\":10,\"Sons\":10,\"Mday\":11,\"Rev\":0,\"Password\":0,\"Sum\":3560}")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response: %s\n", reply)
}
