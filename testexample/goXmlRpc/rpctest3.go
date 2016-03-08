package main

import (
	"bytes"
	"github.com/divan/gorilla-xmlrpc/xml"
	"log"
	"net/http"
)

type Send struct {
	Dbname string
	Uid    string
	Pwd    string
	Model  string
	Method string
	Ids    string
}

func XmlRpcCall(method string, args Send) (reply struct{ Message string }, err error) {
	buf, _ := xml.EncodeClientRequest(method, &args)

	resp, err := http.Post("http://127.0.0.1:8069/xmlrpc/object", "text/xml", bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = xml.DecodeClientResponse(resp.Body, &reply)
	return
}

func main() {
	send := Send{"11", "1", "1989112", "nmbw800.mymodel", "test", "1"}
	reply, err := XmlRpcCall("execute_kw", send)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response: %s\n", reply)
}
