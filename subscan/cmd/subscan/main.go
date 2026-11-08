package main

import (
	"asset-scanner/subscan/internal/subdomain"
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var (
		domain   = flag.String("domain", "example.com", "目标域名")
		dictFile = flag.String("dict", "subdomains.txt", "字典路径")
		workers  = flag.Int("workers", 10, "并发worker数量")
		timeout  = flag.Duration("timeout", 2*time.Second, "单次DNS超时")
	)
	flag.Parse()

	//对输入参数的值进行校验

	if *workers <= 0 {
		log.Fatalf("参数错误： -workers 必须大于0，当前值为：%d", *workers)
	}

	if *timeout <= 0 {
		log.Fatalf("参数错误： -timeout 必须大于0，当前值为：%d", *timeout)
	}

	// 读取字典
	file, err := os.Open(*dictFile)
	if err != nil {
		log.Fatalf("打开字典失败: %v\n", err)
		return
	}
	defer file.Close()
	var prefixes []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		txt := sc.Text()
		if txt != "" {
			prefixes = append(prefixes, txt)
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Printf("读取字典错误: %v\n", err)
	}

	// 构造扫描器，注入真实Resolver
	scan := &subdomain.Scanner{
		Resolver: &subdomain.RealResolver{},
		Workers:  *workers,
		Timeout:  *timeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	res := scan.Scan(ctx, *domain, prefixes)
	jsonOut, err := json.MarshalIndent(res, "", "  ")

	if err != nil {
		log.Fatalf("JSON序列化结果失败: %v", err)
	}
	fmt.Println(string(jsonOut))
}
