
## requests

一个类似 Python requests 的 Go HTTP 请求库

## 使用go get安装

```
go get -u github.com/losenli/requests
```

## 简单的GET请求

```go
package main

import (
	"github.com/losenli/requests"
	"log"
)

func main (){
	resp,err := requests.Get("https://www.httpbin.org/get")
	if err != nil{
		log.Fatalln(err)
	}
	log.Println(resp.Text())
}
```

## 传递 URL 参数

```go
package main

import (
	"github.com/losenli/requests"
	"log"
)

func main() {
	payload := requests.Params{"key1": "value1", "key2": "value2"}
	resp, _ := requests.Get("http://www.httpbin.org/get", payload)
	log.Println(resp.Text())
}
```

## 响应内容

```go
package main

import (
	"github.com/losenli/requests"
	"log"
)

func main() {
	resp, _ := requests.Get("http://www.httpbin.org/get")
	log.Println(resp.Text())
}
```

## 二进制响应内容

```go
package main

import (
	"github.com/losenli/requests"
	"log"
)

func main() {
	resp, _ := requests.Get("http://www.httpbin.org/get")
	log.Println(resp.Content())
}
```

## JSON 响应内容

```go
package main

import (
	"github.com/losenli/requests"
	"log"
)

func main() {
	resp, _ := requests.Get("http://www.httpbin.org/get")
	log.Println(resp.Json())
}
```

## 引入gjson.Result
```go
package main

import (
	"github.com/losenli/requests"
	"log"
)

func main() {
	resp, _ := requests.Get("http://www.httpbin.org/get")
	log.Println(resp.Result())
	log.Println(resp.Result().Get("headers.User-Agent"))
}
```

## Post请求

```go
package main

import (
	"github.com/losenli/requests"
	"log"
)

func main() {
	data := requests.DataItem{
		"name": "requests_post_test",
	}
	resp, _ := requests.Post("https://www.httpbin.org/post", data)
	log.Println(resp.Text())
}
```

## 定制请求头

## 传递 json 数据的 POST 请求

```go
package main

import (
	"github.com/losenli/requests"
)

func main (){
	jsonStr := "{\"name\":\"requests_post_test\"}"
	resp,_ := requests.PostJson("http://www.httpbin.org/post",jsonStr)
	println(resp.Text())
}
```

