package OdooRpc

import (
	"bytes"
	"github.com/divan/gorilla-xmlrpc/xml"
	//"log"
	"net/http"
)

type Send struct {
	Dbname string //pg数据库
	Uid    string //用户id
	Pwd    string //用户密码
	Model  string //调用的模块
	Method string //模块方法
	Args   string //参数，可以是数组
}

//xmlRpc的post组包
func XmlRpcCall(method string, args string) (reply struct{ Message string }, err error) {

	send := Send{"11", "1", "1989112", "nmbw800.mymodel", method, args}
	buf, _ := xml.EncodeClientRequest("execute", &send)

	resp, err := http.Post("http://127.0.0.1:8069/xmlrpc/object", "text/xml", bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = xml.DecodeClientResponse(resp.Body, &reply)
	return
}

// func main() {
// 	reply, err := XmlRpcCall("test", "1")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Printf("Response: %s\n", reply)
// }
