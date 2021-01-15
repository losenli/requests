/* Copyright（2） 2018 by  Mr.Li .
Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package requests

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
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

// 版本号
var VERSION = "0.7.1"

// 封装请求结构体
type Request struct {
	httpReq *http.Request
	Header  *http.Header
	Client  *http.Client
	Debug   int
	Cookies []*http.Cookie
}

// 封装响应结构体
type Response struct {
	R       *http.Response
	content []byte
	text    string
	req     *Request
}

// 自定义类型
type Header map[string]string
type Params map[string]string
type DataItem map[string]string
type Files map[string]string
type Auth []string

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
	req.httpReq.Header.Set("User-Agent", "Go-Requests: "+VERSION)

	req.Client = &http.Client{}

	// auto with Cookies
	// cookiejar.New source code return jar, nil
	jar, _ := cookiejar.New(nil)

	req.Client.Jar = jar

	return &req
}

func Get(origUrl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()

	resp, err = req.Get(origUrl, args...)
	return resp, err
}

func (req *Request) Get(origUrl string, args ...interface{}) (resp *Response, err error) {

	req.httpReq.Method = http.MethodGet

	// set params ?a=b&b=c
	//set Header
	var params []map[string]string

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

// handle URL params
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

	fmt.Println("===========Go RequestDebug ============")

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

// cookies
// cookies only save to Client.Jar
// req.Cookies is temporary
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

func (resp *Response) Content() []byte {

	defer resp.R.Body.Close()
	var err error

	var Body = resp.R.Body
	if resp.R.Header.Get("Content-Encoding") == "gzip" && resp.req.Header.Get("Accept-Encoding") != "" {
		// fmt.Println("gzip")
		reader, err := gzip.NewReader(Body)
		if err != nil {
			return nil
		}
		Body = reader
	}

	resp.content, err = ioutil.ReadAll(Body)
	if err != nil {
		return nil
	}

	return resp.content
}

func (resp *Response) Text() string {
	if resp.content == nil {
		resp.Content()
	}
	resp.text = string(resp.content)
	return resp.text
}

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

// 对响应反序列化操作
func (resp *Response) Unmarshal(v interface{}) error {
	if resp.content == nil {
		resp.Content()
	}
	return json.Unmarshal(resp.content, v)
}

// 使用Map反序列化响应
func (resp *Response) Json() (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	if err := resp.Unmarshal(&result); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// 引入gjson
func (resp *Response) Result() gjson.Result {
	if resp.text == "" {
		resp.Text()
	}
	return gjson.Parse(resp.text)
}

func (resp *Response) Cookies() (cookies []*http.Cookie) {
	httpReq := resp.req.httpReq
	client := resp.req.Client

	cookies = client.Jar.Cookies(httpReq.URL)

	return cookies

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
