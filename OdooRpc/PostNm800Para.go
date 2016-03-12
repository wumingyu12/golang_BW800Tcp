package OdooRpc

import (
	"bytes"
	"github.com/divan/gorilla-xmlrpc/xml"
	//"log"
	"net/http"
)

//常量
const (
	XML_RPC_METHOD = "execute"
	XML_RPC_URL    = "http://127.0.0.1:8069/xmlrpc/object"
	DB_NAME        = "11" //数据库名字
	UID            = 1    //登录的用户名id 1 代表admin
	PWD            = "1989112"
	ODOO_MODEL     = "nmbw800.mymodel"
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

	send := Send{DB_NAME, UID, PWD, ODOO_MODEL, method, args}
	buf, _ := xml.EncodeClientRequest(XML_RPC_METHOD, &send)

	resp, err := http.Post(XML_RPC_URL, "text/xml", bytes.NewBuffer(buf))
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
