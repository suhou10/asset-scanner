package subdomain

import (
	"context"
	"errors"
	"net"
	"sort"
	"strings"
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
	//服务器没有目标域名的错误不存储
	Error string `json:"error,omitempty"`
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
				if ctx.Err() != nil {
					return
				}
				sub := pre + "." + domain
				qCtx, cancel := context.WithTimeout(ctx, s.Timeout)
				ips, err := s.Resolver.LookupHost(qCtx, sub)
				cancel()
				if ctx.Err() != nil {
					return
				}
				res := Result{Subdomain: sub}
				if err != nil {
					var dnsErr *net.DNSError
					switch {
					// 子域名不存在，跳过
					case errors.As(err, &dnsErr) && dnsErr.IsNotFound:
						continue
					// 超时 / 上下文取消，跳过（解决你测试失败的核心）
					case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):

						continue
					// 其他DNS错误：你可以选择记录错误，或者continue丢掉
					default:
						// 方案A：保留错误记录（Result里带上Error字段）
						res.Error = err.Error()
						results <- res
						continue

						// 方案B：不记录其他错误，直接 continue
						// continue
					}
				}

				res.IPs = ips

				results <- res
			}
		}()
	}

	// 创建集合seen，用来给prefixes去重
	seen := make(map[string]struct{}, len(prefixes))

	// 投递任务
	go func() {
		for _, pre := range prefixes {
			pre := strings.ToLower(strings.TrimSpace(pre))
			if pre == "" || strings.HasPrefix(pre, "#") {
				continue
			}

			if _, ok := seen[pre]; ok {
				continue
			}
			seen[pre] = struct{}{}

			select {
			case jobs <- pre:

			case <-ctx.Done():
				close(jobs)
				return
			}
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
