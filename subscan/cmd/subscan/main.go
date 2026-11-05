package main

import (
	"asset-scanner/subscan/internal/subdomain"
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

	// 读取字典
	file, err := os.Open(*dictFile)
	if err != nil {
		fmt.Printf("打开字典失败: %v\n", err)
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

	res := scan.Scan(context.Background(), *domain, prefixes)
	jsonOut, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(jsonOut))
}
