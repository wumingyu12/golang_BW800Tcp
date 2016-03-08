package main

import (
	"fmt"
	"github.com/kolo/xmlrpc"
)

func main() {
	client, _ := xmlrpc.NewClient("http://127.0.0.1:8069/xmlrpc/2/common", nil)
	result := struct {
		ServerVersion string `xmlrpc:"server_version"`
		//因为返回的json，'server_version_info': [8, 0, 0, 'final', 0],是数组没空解释到结构体中所以忽略算了，否则整个结构体都无法解析
		//ServerVersionInfo string `xmlrpc:"server_version_info"`
		ServerSerie     string `xmlrpc:"server_serie"`
		ProtocolVersion string `xmlrpc:"protocol_version"`
	}{}
	client.Call("version", nil, &result)
	fmt.Printf("Version: %s\n", result.ServerVersion)
	//返回的json为
	//{'server_version_info': [8, 0, 0, 'final', 0], 'server_serie': '8.0', 'server_version': '8.0', 'protocol_version': 1}

}
