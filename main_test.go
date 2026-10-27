package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望状态码 %d，实际 %d", http.StatusOK, rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	wantCT := "application/json; charset=utf-8"
	if ct != wantCT {
		t.Errorf("Content-Type 期望 %s，得到 %s", wantCT, ct)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("json解析失败: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status 期望 ok，得到 %s", resp["status"])
	}
}

func TestScanHandler_MissingURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/scan", nil)
	rec := httptest.NewRecorder()

	scanHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("期望 %d，实际 %d", http.StatusBadRequest, rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "缺少 url 参数" {
		t.Errorf("error 消息不匹配：%v", resp)
	}
}

func TestScanHandler_Success(t *testing.T) {
	mockHTML := `<html><title>Test Page Title</title></html>`
	mockSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockHTML))
	}))
	defer mockSrv.Close()

	targetURL := mockSrv.URL
	req := httptest.NewRequest("GET", "/scan?url="+targetURL, nil)
	rec := httptest.NewRecorder()

	scanHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望 %d，实际 %d，body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var res ScanResult
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("status code 期望200，得到 %d", res.StatusCode)
	}
	if res.Title != "Test Page Title" {
		t.Errorf("title 期望 'Test Page Title'，得到 '%s'", res.Title)
	}
}
