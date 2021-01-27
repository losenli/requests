package requests

import (
	"compress/gzip"
	"encoding/json"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"os"
)

// 封装响应结构体
type Response struct {
	R          *http.Response
	content    []byte
	text       string
	req        *Request
	StatusCode int
	Time       int64
	Length     int64
}

// 返回二进制响应内容
func (resp *Response) Content() []byte {
	defer resp.R.Body.Close()

	var Body = resp.R.Body
	if resp.R.Header.Get("Content-Encoding") == "gzip" && resp.req.Header.Get("Accept-Encoding") != "" {
		reader, err := gzip.NewReader(Body)
		if err != nil {
			return nil
		}
		Body = reader
	}
	var err error
	resp.content, err = ioutil.ReadAll(Body)
	if err != nil {
		return nil
	}
	return resp.content
}

// 返回字符串响应内容
func (resp *Response) Text() string {
	if resp.content == nil {
		resp.Content()
	}
	resp.text = string(resp.content)
	return resp.text
}

// 响应内容保存到文件
func (resp *Response) SaveFile(filename string) error {
	if resp.content == nil {
		resp.Content()
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(resp.content)
	_ = f.Sync()
	return err
}

// 响应内容反序列化
func (resp *Response) Unmarshal(v interface{}) error {
	if resp.content == nil {
		resp.Content()
	}
	return json.Unmarshal(resp.content, v)
}

// 默认使用Map反序列化响应
func (resp *Response) Json() (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	if err := resp.Unmarshal(&result); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// 转换成gjson.Result,方便解析JSON响应
func (resp *Response) Result() gjson.Result {
	if resp.text == "" {
		resp.Text()
	}
	return gjson.Parse(resp.text)
}

// 获取Cookies
func (resp *Response) Cookies() (cookies []*http.Cookie) {
	httpReq := resp.req.httpReq
	client := resp.req.Client
	cookies = client.Jar.Cookies(httpReq.URL)
	return cookies
}
