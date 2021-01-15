
# requests

Requests is an HTTP library  , it is easy to use. Similar to Python requests.

# 使用go get安装

```
go get -u github.com/asmcos/requests
```

# 简单的GET请求

``` go
package main

import "github.com/asmcos/requests"

func main (){

        resp,err := requests.Get("http://www.zhanluejia.net.cn")
        if err != nil{
          return
        }
        println(resp.Text())
}
```

## Post请求

``` go
package main

import "github.com/asmcos/requests"


func main (){

        data := requests.DataItem{
          "name":"requests_post_test",
        }
        resp,_ := requests.Post("https://www.httpbin.org/post",data)
        println(resp.Text())
}

```

     Server return data...

``` json
{
  "args": {},
  "data": "",
  "files": {},
  "form": {
    "name": "requests_post_test"
  },
  "headers": {
    "Accept-Encoding": "gzip",
    "Connection": "close",
    "Content-Length": "23",
    "Content-Type": "application/x-www-form-urlencoded",
    "Host": "www.httpbin.org",
    "User-Agent": "Go-Requests 0.5"
  },
  "json": null,
  "origin": "114.242.34.110",
  "url": "https://www.httpbin.org/post"
}

```


## PostJson

``` go
package main

import "github.com/asmcos/requests"


func main (){

        jsonStr := "{\"name\":\"requests_post_test\"}"
        resp,_ := requests.PostJson("https://www.httpbin.org/post",jsonStr)
        println(resp.Text())
}

```

     Server return data...

``` json
{
  "args": {},
  "data": "",
  "files": {},
  "form": {
    "name": "requests_post_test"
  },
  "headers": {
    "Accept-Encoding": "gzip",
    "Connection": "close",
    "Content-Length": "23",
    "Content-Type": "application/x-www-form-urlencoded",
    "Host": "www.httpbin.org",
    "User-Agent": "Go-Requests 0.5"
  },
  "json": null,
  "origin": "114.242.34.110",
  "url": "https://www.httpbin.org/post"
}

```

# Set header

### example 1

``` go
req := requests.Requests()

resp,err := req.Get("http://www.zhanluejia.net.cn",requests.Header{"Referer":"http://www.jeapedu.com"})
if (err == nil){
  println(resp.Text())
}
```

### example 2

``` go
req := requests.Requests()
req.Header.Set("accept-encoding", "gzip, deflate, br")
resp,_ := req.Get("http://www.zhanluejia.net.cn",requests.Header{"Referer":"http://www.jeapedu.com"})
println(resp.Text())

```

### example 3

``` go
h := requests.Header{
  "Referer":         "http://www.jeapedu.com",
  "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
}
resp,_ := req.Get("http://wwww.zhanluejia.net.cn",h)

h2 := requests.Header{
  ...
  ...
}
h3,h4 ....
// two or more headers ...
resp,_ = req.Get("http://www.zhanluejia.net.cn",h,h2,h3,h4)
```


# Set params

``` go
p := requests.Params{
  "title": "The blog",
  "name":  "file",
  "id":    "12345",
}
resp,_ := req.Get("http://www.cpython.org", p)

```


# Auth

Test with the `correct` user information.

``` go
req := requests.Requests()
resp,_ := req.Get("https://api.github.com/user",requests.Auth{"asmcos","password...."})
println(resp.Text())
```

github return

```
{"login":"asmcos","id":xxxxx,"node_id":"Mxxxxxxxxx==".....
```

# 返回JSON数据

```go
req := requests.Requests()
resp,_ := req.Get("https://httpbin.org/json")
// 返回json数据，使用Map反序列化
result, _ := resp.Json()

for k,v := range result{
	fmt.Println(k,v)
}
```
# 引入gjson.result
```go
req := requests.Requests()
resp,_ := req.Get("https://httpbin.org/json")
// Result()返回gjson.Result
fmt.Println(resp.Result().Get("slideshow.slides.1.title"))
```


# SetTimeout

```
req := Requests()
req.Debug = 1

// 20 Second
req.SetTimeout(20)
req.Get("http://golang.org")
```

# Get Cookies

``` go
resp,_ = req.Get("https://www.httpbin.org")
coo := resp.Cookies()
// coo is [] *http.Cookies
println("********cookies*******")
for _, c:= range coo{
  fmt.Println(c.Name,c.Value)
}
```
