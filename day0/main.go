package main

import (
	"fmt"
	"log"
	"net/http"
)

// 实现一个web应用时最先考虑的是使用哪个框架，不同的框架设计理念和提供的功能有很大的差别
// 框架核心位我们解决了什么问题？
func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/count", handler)

	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

// 其中web开发中某些简单的需求http包并不能完全支持，需要手动实现
// 1. 动态路由：例如hello/:name，我们需要解析name的值
// 2. 模板：没有简化统一的HTML机制
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "URL.path = %q\n", r.URL.Path)
	// 字符串格式化为带有引号的字符串
}
