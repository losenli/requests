package requests

import (
	"log"
	"testing"
)

func TestPostJson(t *testing.T) {
	data := `{"name": "lzh"}`
	resp, err := PostJson("http://httpbin.org/post", data)
	if err != nil {
		t.Error(err)
	}
	log.Println(resp.Time)
}

func TestRequest_Send(t *testing.T) {
	req := Requests()
	//data := `{"name": "lzh"}`
	resp, err := req.Send("Put", "http://httpbin.org/put")
	if err != nil {
		t.Error(err)
	}
	log.Println(resp.Time)
}
