package subdomain

import (
	"context"
	"net"
	"sort"
	"sync"
	"time"
)

// Resolver 接口：抽象DNS查询，方便mock
type Resolver interface {
	LookupHost(ctx context.Context, host string) ([]string, error)
}

// Result 扫描结果
type Result struct {
	Subdomain string   `json:"subdomain"`
	IPs       []string `json:"ips,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// Scanner 扫描器实例
type Scanner struct {
	Resolver Resolver
	Workers  int
	Timeout  time.Duration
}

// RealResolver 真实DNS实现，调用系统net包
type RealResolver struct{}

func (r *RealResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	return net.DefaultResolver.LookupHost(ctx, host)
}

// Scan 核心扫描函数：传入域名、前缀列表，返回排序后的结果
func (s *Scanner) Scan(ctx context.Context, domain string, prefixes []string) []Result {
	jobs := make(chan string, len(prefixes))
	results := make(chan Result, len(prefixes))
	var wg sync.WaitGroup

	// 启动worker
	for i := 0; i < s.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pre := range jobs {
				sub := pre + "." + domain
				qCtx, cancel := context.WithTimeout(ctx, s.Timeout)
				ips, err := s.Resolver.LookupHost(qCtx, sub)
				cancel()
				res := Result{Subdomain: sub}
				if err != nil {
					res.Error = err.Error()
				} else {
					res.IPs = ips
				}
				results <- res
			}
		}()
	}

	// 投递任务
	go func() {
		for _, pre := range prefixes {
			jobs <- pre
		}
		close(jobs)
	}()

	// worker结束关闭结果通道
	go func() {
		wg.Wait()
		close(results)
	}()

	var resList []Result
	for r := range results {
		resList = append(resList, r)
	}

	// 排序
	sort.Slice(resList, func(i, j int) bool {
		return resList[i].Subdomain < resList[j].Subdomain
	})
	return resList
}
