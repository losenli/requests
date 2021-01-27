package requests

import (
	"net/http"
	"testing"
)

var req = Requests()

func TestGet(t *testing.T) {
	resp, err := Get("https://www.httpbin.org/get")
	if err != nil {
		t.Error(err)
	}
	if resp.R.StatusCode != http.StatusOK {
		t.Error("状态码异常！")
	}
}

func TestRequest_Get(t *testing.T) {
	resp, err := req.Get("https://www.httpbin.org/get")
	if err != nil {
		t.Error(err)
	}
	if resp.R.StatusCode != http.StatusOK {
		t.Error("状态码异常！")
	}
}
