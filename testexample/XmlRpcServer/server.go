package main

import (
	//"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/xmlrpc/object", XmlRpcHandler)
	fmt.Println("监听5656端口")
	http.ListenAndServe(":5656", nil)

}

func XmlRpcHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if req.Method == "POST" {
		result, _ := ioutil.ReadAll(req.Body)
		req.Body.Close()
		fmt.Printf("%s\n", result)
	}
}
