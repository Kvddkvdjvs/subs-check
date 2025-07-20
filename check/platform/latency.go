package platform

import (
	"net/http"
	"time"
)

// CheckLatency 测试到 Cloudflare 的延迟
func CheckLatency(client *http.Client) (int, error) {
	// 使用 1.1.1.1 的一个特殊端点，它几乎没有处理时间，非常适合测延迟
	const testURL = "https://1.1.1.1/cdn-cgi/trace"

	startTime := time.Now()

	// 我们只关心能否建立连接和收到响应头，不关心内容，所以用 HEAD 请求最快
	req, err := http.NewRequest("HEAD", testURL, nil)
	if err != nil {
		return -1, err
	}
	// 模拟浏览器 User-Agent，防止被拦截
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	resp.Body.Close()

	// 计算从发出请求到收到响应头的时间差
	latency := time.Since(startTime).Milliseconds()

	return int(latency), nil
}
