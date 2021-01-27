package main

import "requests"

/**
* @Author : CarpLi
* @Date   : 2021/1/16 10:33
* @Desc   :
 */

func main() {
	resp, _ := requests.Get("http://www.baidu.com")
	println(resp.Text())
	_ = resp.SaveFile("1.txt")
}
