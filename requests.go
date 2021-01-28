package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// 自定义类型
type Header map[string]string
type Params map[string]string
type DataItem map[string]string
type Files map[string]string
type Auth []string

// 封装请求结构体
type Request struct {
	httpReq *http.Request
	Header  *http.Header
	Client  *http.Client
	Debug   int
	Host    string
	Cookies []*http.Cookie
}

// 支持的请求类型
const (
	TypeJSON       = "application/json"
	TypeXML        = "application/xml"
	TypeForm       = "application/x-www-form-urlencoded"
	TypeFormData   = "application/x-www-form-urlencoded"
	TypeUrlencoded = "application/x-www-form-urlencoded"
	TypeHTML       = "text/html"
	TypeText       = "text/plain"
	TypeMultipart  = "multipart/form-data"
)

func Requests() *Request {
	var req Request

	req.httpReq = &http.Request{
		Method:     "GET",
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	req.Header = &req.httpReq.Header
	req.httpReq.Header.Set("User-Agent", "Go-Requests")

	req.Client = &http.Client{}
	jar, _ := cookiejar.New(nil)
	req.Client.Jar = jar

	return &req
}

// 发送HTTP请求
func (req *Request) Send(method, origUrl string, args ...interface{}) (resp *Response, err error) {
	if strings.TrimSpace(method) == "" {
		return nil, errors.New("method can't be empty")
	}
	req.httpReq.Method = strings.ToUpper(method)

	var params []map[string]string
	var dataItem []map[string]string
	var files []map[string]string

	for _, arg := range args {
		switch argType := arg.(type) {
		case string:
			req.setBodyRawBytes(ioutil.NopCloser(strings.NewReader(arg.(string))))
		case Header:
			for k, v := range argType {
				req.Header.Set(k, v)
			}
		case Params:
			params = append(params, argType)
		case DataItem:
			dataItem = append(dataItem, argType)
		case Files:
			files = append(files, argType)
		case Auth:
			req.httpReq.SetBasicAuth(argType[0], argType[1])
		default:
			b := new(bytes.Buffer)
			if err = json.NewEncoder(b).Encode(argType); err != nil {
				return nil, err
			}
			req.setBodyRawBytes(ioutil.NopCloser(b))
		}
	}
	distUrl, _ := buildURLParams(origUrl, params...)

	if req.httpReq.Method != "GET" && len(dataItem) > 0 {
		if len(files) > 0 {
			req.buildFilesAndForms(files, dataItem)

		} else {
			Forms := req.buildForms(dataItem...)
			req.setBodyBytes(Forms)
		}
	}

	URL, err := url.Parse(distUrl)
	if err != nil {
		return nil, err
	}
	req.httpReq.URL = URL
	req.ClientSetCookies()
	startTime := time.Now()
	// 设置响应内容
	if response, err := req.Client.Do(req.httpReq); err == nil {
		resp = &Response{}
		resp.R = response
		resp.req = req
		resp.Time = time.Since(startTime).Milliseconds()
		resp.Length = req.httpReq.ContentLength
		return resp, nil
	}
	return nil, err
}

func Get(origUrl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()
	resp, err = req.Get(origUrl, args...)
	return resp, err
}

// 构造 GET 请求
func (req *Request) Get(origUrl string, args ...interface{}) (resp *Response, err error) {
	resp, err = req.Send(http.MethodGet, origUrl, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 构造 POST 请求
func (req *Request) Post(origUrl string, args ...interface{}) (resp *Response, err error) {
	resp, err = req.Send(http.MethodPost, origUrl, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func buildURLParams(userURL string, params ...map[string]string) (string, error) {
	parsedURL, err := url.Parse(userURL)
	if err != nil {
		return "", err
	}
	parsedQuery, err := url.ParseQuery(parsedURL.RawQuery)

	if err != nil {
		return "", nil
	}
	for _, param := range params {
		for key, value := range param {
			parsedQuery.Add(key, value)
		}
	}
	return addQueryParams(parsedURL, parsedQuery), nil
}

func addQueryParams(parsedURL *url.URL, parsedQuery url.Values) string {
	if len(parsedQuery) > 0 {
		return strings.Join([]string{strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1), parsedQuery.Encode()}, "?")
	}
	return strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1)
}

func (req *Request) RequestDebug() {
	if req.Debug != 1 {
		return
	}
	message, err := httputil.DumpRequestOut(req.httpReq, false)
	if err != nil {
		return
	}
	log.Println(string(message))
	if len(req.Client.Jar.Cookies(req.httpReq.URL)) > 0 {
		fmt.Println("Cookies:")
		for _, cookie := range req.Client.Jar.Cookies(req.httpReq.URL) {
			fmt.Println(cookie)
		}
	}
}

func (req *Request) SetCookie(cookie *http.Cookie) {
	req.Cookies = append(req.Cookies, cookie)
}

func (req *Request) ClearCookies() {
	req.Cookies = req.Cookies[0:0]
}

func (req *Request) ClientSetCookies() {
	if len(req.Cookies) > 0 {
		req.Client.Jar.SetCookies(req.httpReq.URL, req.Cookies)
		req.ClearCookies()
	}
}

// 设置超时时间
func (req *Request) SetTimeout(n time.Duration) {
	req.Client.Timeout = n * time.Second
}

func (req *Request) Proxy(proxyUrl string) {
	URI := url.URL{}
	urlProxy, err := URI.Parse(proxyUrl)
	if err != nil {
		log.Println("Set proxy failed")
		return
	}
	req.Client.Transport = &http.Transport{
		Proxy:           http.ProxyURL(urlProxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

func (resp *Response) ResponseDebug() {
	if resp.req.Debug != 1 {
		return
	}
	message, err := httputil.DumpResponse(resp.R, false)
	if err != nil {
		return
	}
	log.Println(message)
}

func Post(origUrl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()
	resp, err = req.Post(origUrl, args...)
	return resp, err
}

func PostJson(origUrl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()
	resp, err = req.PostJson(origUrl, args...)
	return resp, err
}

// 发送JSON请求
func (req *Request) PostJson(origUrl string, args ...interface{}) (resp *Response, err error) {
	headers := Header{"Content-Type": TypeJSON}
	resp, err = req.Send(http.MethodPost, origUrl, headers, args)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (req *Request) setBodyBytes(Forms url.Values) {
	data := Forms.Encode()
	req.httpReq.Body = ioutil.NopCloser(strings.NewReader(data))
	req.httpReq.ContentLength = int64(len(data))
}

func (req *Request) setBodyRawBytes(read io.ReadCloser) {
	req.httpReq.Body = read
}

func (req *Request) buildFilesAndForms(files []map[string]string, dataItem []map[string]string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for _, file := range files {
		for k, v := range file {
			part, err := w.CreateFormFile(k, v)
			if err != nil {
				panic(err)
			}
			file := openFile(v)
			_, err = io.Copy(part, file)
			if err != nil {
				panic(err)
			}
		}
	}

	for _, data := range dataItem {
		for k, v := range data {
			_ = w.WriteField(k, v)
		}
	}

	_ = w.Close()
	req.httpReq.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))
	req.httpReq.ContentLength = int64(b.Len())
	req.Header.Set("Content-Type", w.FormDataContentType())
}

// 构造 POST FORM 表单
func (req *Request) buildForms(dataItem ...map[string]string) (Forms url.Values) {
	Forms = url.Values{}
	for _, data := range dataItem {
		for key, value := range data {
			Forms.Add(key, value)
		}
	}
	return Forms
}

// open file for post upload files
func openFile(filename string) *os.File {
	r, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return r
}
