package platform

import (
	"io"
	"net/http"
	"strings"
)

// CheckTwitter 检查节点是否能访问 Twitter/X
// 我们访问官方 @twitter 账号页面，并检查页面标题
func CheckTwitter(client *http.Client) (bool, error) {
	// 使用官方账号页面作为测试链接，稳定且公开
	const testURL = "https://x.com/twitter"
	
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return false, err
	}

	// 很多网站会拒绝没有 User-Agent 的请求，所以我们模拟一个浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 只读取少量数据进行判断即可，无需加载整个页面
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2048)) // 读取最多 2KB
	if err != nil {
		return false, err
	}

	// 检查返回的 HTML 中是否包含页面标题的特定部分，这是成功的关键标志
	if strings.Contains(string(body), "(@twitter) / X") {
		return true, nil
	}

	return false, nil
}
