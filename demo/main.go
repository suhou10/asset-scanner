package main

import (
	"log"
	"net/http"
)

// 一个本地测试靶站：访问 http://127.0.0.1:9090 会返回一个带 title 的页面，
// 用来测试 /scan 接口，避免去扫公网。
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html><head><title>Demo Target</title></head>
<body><h1>这是一个本地测试页面</h1></body></html>`))
	})
	log.Println("测试靶站：http://127.0.0.1:9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
