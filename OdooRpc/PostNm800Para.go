package OdooRpc

import (
	"bytes"
	"github.com/divan/gorilla-xmlrpc/xml"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

//常量
const (
	XML_RPC_METHOD               = "execute"
	XML_RPC_URL                  = "http://127.0.0.1:8069/xmlrpc/object"
	DB_NAME                      = "11" //数据库名字
	UID                          = 1    //登录的用户名id 1 代表admin
	PWD                          = "1989112"
	ODOO_PostParaStruct_MODEL    = "nmbw800.mymodel"
	PostParaStruct_Method        = "addIntance"           //推送结构体时接收的方法
	Odoo_PostDayRecord_Model     = "nmbw800.dayrecord"    //推送日记录的模型
	Odoo_PostDayRecord_Method    = "addItemByJson"        //推送日记录的方法
	Odoo_PostDetailRecord_Model  = "nmbw800.detailrecord" //推送日记录的模型
	Odoo_PostDetailRecord_Method = "addItemByJson"        //推送日记录的方法
)

type Send struct {
	Dbname string //pg数据库
	Uid    int    //用户id
	Pwd    string //用户密码
	Model  string //调用的模块
	Method string //模块方法
	Args   string //参数，可以是数组,我们发送的是json化的结构体
}

//xmlRpc的post组包
func XmlRpcCall(model string, method string, args string) (reply struct{ Message string }, err error) {

	send := Send{DB_NAME, UID, PWD, model, method, args}
	buf, _ := xml.EncodeClientRequest(XML_RPC_METHOD, &send)

	resp, err := http.Post(XML_RPC_URL, "text/xml", bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = xml.DecodeClientResponse(resp.Body, &reply)
	return
}

//推送parastruct到odoo
func PostParaStruct(jsonstruct string) {
	reply, err := XmlRpcCall(ODOO_PostParaStruct_MODEL, PostParaStruct_Method, jsonstruct)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("添加参数结构体到odoo数据库，id：%s", reply)
}

//推送日下料记录到odoo
func PostDayRecord(jsonRecordList string) {
	reply, err := XmlRpcCall(Odoo_PostDayRecord_Model, Odoo_PostDayRecord_Method, jsonRecordList)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("添加日记录列表结构体到odoo数据库：%s", reply)

}

//推送详细下料记录到odoo
func PostDetailRecord(jsonRecordList string) {
	reply, err := XmlRpcCall(Odoo_PostDetailRecord_Model, Odoo_PostDetailRecord_Method, jsonRecordList)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("添加详细记录列表结构体到odoo数据库：%s", reply)

}

// func main() {
// 	reply, err := XmlRpcCall("test", "1")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Printf("Response: %s\n", reply)
// }
