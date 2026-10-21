package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// 匹配 <title>...</title>；(?is) 表示忽略大小写，并让 . 能匹配换行
var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// ScanResult 是 /scan 返回的 JSON 结构
type ScanResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Title      string `json:"title"`
	Error      string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/scan", scanHandler)

	log.Println("服务启动：http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// /health：只返回 {"status":"ok"}，用来确认服务活着
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// /scan?url=http://...：请求目标，返回状态码和网页标题
func scanHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("url")
	if target == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 url 参数"})
		return
	}
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(target)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ScanResult{URL: target, Error: err.Error()})
		return
	}
	defer resp.Body.Close()

	// 最多读 1MB，防止目标返回超大页面把内存吃满
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	title := ""
	if m := titleRe.FindSubmatch(body); len(m) > 1 {
		title = strings.TrimSpace(string(m[1]))
	}

	writeJSON(w, http.StatusOK, ScanResult{
		URL:        target,
		StatusCode: resp.StatusCode,
		Title:      title,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
