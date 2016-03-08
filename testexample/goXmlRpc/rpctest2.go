package main

//返回结构体不复杂的数据
import (
	"fmt"
	"github.com/kolo/xmlrpc"
)

type Send struct {
	Dbname string
	Uid    string
	Pwd    string
	Model  string
	Method string
	Ids    string
	Fields string
}

func main() {
	client, _ := xmlrpc.NewClient("http://127.0.0.1:8069/xmlrpc/object", nil)
	result := struct {
		Id string `xmlrpc:"id"`
	}{}
	send := &Send{"11", "1", "1989112", "hr.employee.category", "read", "1", "name"}
	fmt.Println(send)
	err := client.Call("execute", send, &result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Version: %x\n", result.Id)
	//返回的json为
	//{'server_version_info': [8, 0, 0, 'final', 0], 'server_serie': '8.0', 'server_version': '8.0', 'protocol_version': 1}

}
