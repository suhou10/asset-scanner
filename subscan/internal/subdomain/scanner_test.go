package subdomain

import (
	"context"
	"net"
	"testing"
	"time"
)

// MockResolver 假解析器，单元测试用，不走真实网络
type MockResolver struct {
	mockData map[string][]string
	mockErr  error
}

func (m *MockResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	if host == "www.example.com" {
		return m.mockData[host], nil
	}
	return nil, &net.DNSError{
		Err:        "no such host",
		IsNotFound: true,
	}
}

func TestScanner_Scan(t *testing.T) {
	// mock数据：www.example.com 返回 1.1.1.1
	mock := &MockResolver{
		mockData: map[string][]string{
			"www.example.com": {"1.1.1.1"},
		},
	}
	scanner := &Scanner{
		Resolver: mock,
		Workers:  2,
		Timeout:  1 * time.Second,
	}
	prefix := []string{"www", "admin"}
	results := scanner.Scan(context.Background(), "example.com", prefix)

	if len(results) != 2 {
		t.Fatalf("预期2条结果，拿到%d", len(results))
	}

}

//TestScan_Dedup去重测试

func TestScanner_Dedup(t *testing.T) {
	mock := &MockResolver{}
	scanner := &Scanner{
		Resolver: mock,
		Workers:  2,
		Timeout:  1 * time.Second,
	}
	results := scanner.Scan(context.Background(), "example.com", []string{"www", "www"})
	if len(results) != 1 {
		t.Fatalf("dedup fail, expect 1,got %d", len(results))
	}
}
