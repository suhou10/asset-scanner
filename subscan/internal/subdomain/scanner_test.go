package subdomain

import (
	"context"
	"testing"
	"time"
)

// MockResolver 假解析器，单元测试用，不走真实网络
type MockResolver struct {
	mockData map[string][]string
	mockErr  error
}

func (m *MockResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return m.mockData[host], nil
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
