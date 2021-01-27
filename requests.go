package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
func (req *Request) Send(method, url string, args ...interface{}) (resp *Response, err error) {
	if strings.TrimSpace(method) == "" {
		return nil, errors.New("method can't be empty")
	}
	req.httpReq.Method = strings.ToUpper(method)
	return nil, nil
}

func Get(origUrl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()
	resp, err = req.Get(origUrl, args...)
	return resp, err
}

func (req *Request) Get(origUrl string, args ...interface{}) (resp *Response, err error) {
	req.httpReq.Method = http.MethodGet

	var params []map[string]string

	delete(req.httpReq.Header, "Cookie")

	for _, arg := range args {
		switch a := arg.(type) {
		case Header:
			for k, v := range a {
				req.Header.Set(k, v)
			}
		case Params:
			params = append(params, a)
		case Auth:
			req.httpReq.SetBasicAuth(a[0], a[1])
		}
	}
	distUrl, _ := buildURLParams(origUrl, params...)

	URL, err := url.Parse(distUrl)
	if err != nil {
		return nil, err
	}
	req.httpReq.URL = URL

	req.ClientSetCookies()

	req.RequestDebug()

	res, err := req.Client.Do(req.httpReq)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp = &Response{}
	resp.R = res
	resp.req = req
	resp.ResponseDebug()
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
	fmt.Println(string(message))

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
		// 1. Cookies have content, Copy Cookies to Client.jar
		// 2. Clear  Cookies
		req.Client.Jar.SetCookies(req.httpReq.URL, req.Cookies)
		req.ClearCookies()
	}

}

// set timeout s = second
func (req *Request) SetTimeout(n time.Duration) {
	req.Client.Timeout = n * time.Second
}

func (req *Request) Proxy(proxyUrl string) {

	URI := url.URL{}
	urlProxy, err := URI.Parse(proxyUrl)
	if err != nil {
		fmt.Println("Set proxy failed")
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
	fmt.Println("===========Go ResponseDebug ============")

	message, err := httputil.DumpResponse(resp.R, false)
	if err != nil {
		return
	}

	fmt.Println(string(message))

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

func (req *Request) PostJson(origUrl string, args ...interface{}) (resp *Response, err error) {

	req.httpReq.Method = http.MethodPost
	req.Header.Add("Content-Type", "application/json")

	delete(req.httpReq.Header, "Cookie")

	for _, arg := range args {
		switch a := arg.(type) {
		case Header:
			for k, v := range a {
				req.Header.Set(k, v)
			}
		case string:
			req.setBodyRawBytes(ioutil.NopCloser(strings.NewReader(arg.(string))))
		case Auth:
			req.httpReq.SetBasicAuth(a[0], a[1])
		default:
			b := new(bytes.Buffer)
			err = json.NewEncoder(b).Encode(a)
			if err != nil {
				return nil, err
			}
			req.setBodyRawBytes(ioutil.NopCloser(b))
		}
	}

	URL, err := url.Parse(origUrl)
	if err != nil {
		return nil, err
	}
	req.httpReq.URL = URL

	req.ClientSetCookies()

	req.RequestDebug()

	res, err := req.Client.Do(req.httpReq)
	req.httpReq.Body = nil
	req.httpReq.GetBody = nil
	req.httpReq.ContentLength = 0

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp = &Response{}
	resp.R = res
	resp.req = req
	resp.ResponseDebug()
	return resp, nil
}

func (req *Request) Post(origUrl string, args ...interface{}) (resp *Response, err error) {

	req.httpReq.Method = http.MethodPost

	if req.Header.Get("Content-Type") != "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// set params ?a=b&b=c
	//set Header
	var params []map[string]string
	var dataItem []map[string]string
	var files []map[string]string

	//reset Cookies,
	//Client.Do can copy cookie from client.Jar to req.Header
	delete(req.httpReq.Header, "Cookie")

	for _, arg := range args {
		switch a := arg.(type) {
		// arg is Header , set to request header
		case Header:
			for k, v := range a {
				req.Header.Set(k, v)
			}
		case Params:
			params = append(params, a)
		case DataItem:
			dataItem = append(dataItem, a)
		case Files:
			files = append(files, a)
		case Auth:
			req.httpReq.SetBasicAuth(a[0], a[1])
		}
	}
	distUrl, _ := buildURLParams(origUrl, params...)

	if len(files) > 0 {
		req.buildFilesAndForms(files, dataItem)

	} else {
		Forms := req.buildForms(dataItem...)
		req.setBodyBytes(Forms) // set forms to body
	}
	//prepare to Do
	URL, err := url.Parse(distUrl)
	if err != nil {
		return nil, err
	}
	req.httpReq.URL = URL

	req.ClientSetCookies()

	req.RequestDebug()

	res, err := req.Client.Do(req.httpReq)

	// clear post param
	req.httpReq.Body = nil
	req.httpReq.GetBody = nil
	req.httpReq.ContentLength = 0

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp = &Response{}
	resp.R = res
	resp.req = req
	resp.ResponseDebug()
	return resp, nil
}

// only set forms
func (req *Request) setBodyBytes(Forms url.Values) {

	// maybe
	data := Forms.Encode()
	req.httpReq.Body = ioutil.NopCloser(strings.NewReader(data))
	req.httpReq.ContentLength = int64(len(data))
}

// only set forms
func (req *Request) setBodyRawBytes(read io.ReadCloser) {
	req.httpReq.Body = read
}

// upload file and form
// build to body format
func (req *Request) buildFilesAndForms(files []map[string]string, dataItem []map[string]string) {

	//handle file multipart

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for _, file := range files {
		for k, v := range file {
			part, err := w.CreateFormFile(k, v)
			if err != nil {
				fmt.Printf("Upload %s failed!", v)
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

// build post Form data
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
