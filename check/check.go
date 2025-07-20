package check

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"

	"github.com/beck-8/subs-check/check/platform"
	"github.com/beck-8/subs-check/config"
	proxyutils "github.com/beck-8/subs-check/proxy"
	"github.com/juju/ratelimit"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
)

// Result 存储节点检测结果
type Result struct {
	Proxy      map[string]any
	Openai     bool
	Google     bool
	Cloudflare bool
	Gemini     bool
	Twitter    bool
	Combo      bool
	IP         string
	IPRisk     string // [修改] 加回 IPRisk 字段
	Country    string
}

// ... (从这里到 checkProxy 函数之前的所有代码都和之前一样，无需改动)
// ... (为了简洁，我省略了中间不变的代码)

// checkProxy 检测单个代理
func (pc *ProxyChecker) checkProxy(proxy map[string]any) *Result {
	res := &Result{
		Proxy: proxy,
	}

	if os.Getenv("SUB_CHECK_SKIP") != "" {
		return res
	}

	httpClient := CreateClient(proxy)
	if httpClient == nil {
		slog.Debug(fmt.Sprintf("创建代理Client失败: %v", proxy["name"]))
		return nil
	}
	defer httpClient.Close()

	cloudflare, err := platform.CheckCloudflare(httpClient.Client)
	if err != nil || !cloudflare {
		return nil
	}

	google, err := platform.CheckGoogle(httpClient.Client)
	if err != nil || !google {
		return nil
	}

	var speed int
	if config.GlobalConfig.SpeedTestUrl != "" {
		speed, _, err = platform.CheckSpeed(httpClient.Client, Bucket)
		if err != nil || speed < config.GlobalConfig.MinSpeed {
			return nil
		}
	}

	if config.GlobalConfig.MediaCheck {
		for _, plat := range config.GlobalConfig.Platforms {
			switch plat {
			case "openai":
				if ok, _ := platform.CheckOpenai(httpClient.Client); ok {
					res.Openai = true
				}
			case "gemini":
				if ok, _ := platform.CheckGemini(httpClient.Client); ok {
					res.Gemini = true
				}
			case "twitter":
				if ok, _ := platform.CheckTwitter(httpClient.Client); ok {
					res.Twitter = true
				}
			case "combo":
				if ok, _ := platform.CheckCombo(httpClient.Client); ok {
					res.Combo = true
				}
			case "iprisk": // [修改] 加回 IPRisk 检测逻辑
				country, ip := proxyutils.GetProxyCountry(httpClient.Client)
				if ip == "" {
					break
				}
				res.IP = ip
				res.Country = country
				risk, err := platform.CheckIPRisk(httpClient.Client, ip)
				if err == nil {
					res.IPRisk = risk
				} else {
					slog.Debug(fmt.Sprintf("查询IP风险失败: %v", err))
				}
			}
		}
	}
	// 更新代理名称
	pc.updateProxyName(res, httpClient, speed)
	pc.incrementAvailable()
	return res
}

// updateProxyName 更新代理名称
func (pc *ProxyChecker) updateProxyName(res *Result, httpClient *ProxyClient, speed int) {
	// 以节点IP查询位置重命名节点
	if config.GlobalConfig.RenameNode {
		// [修改] 优化逻辑，如果 iprisk 已经查过 IP，就不重复查了
		if res.Country == "" {
			country, ip := proxyutils.GetProxyCountry(httpClient.Client)
			res.IP = ip
			res.Country = country
		}
		res.Proxy["name"] = config.GlobalConfig.NodePrefix + proxyutils.Rename(res.Country)
	}

	name := res.Proxy["name"].(string)
	name = strings.TrimSpace(name)

	var tags []string
	// 获取速度
	if config.GlobalConfig.SpeedTestUrl != "" {
		name = regexp.MustCompile(`\s*\|(?:\s*[\d.]+[KM]B/s)`).ReplaceAllString(name, "")
		var speedStr string
		if speed < 1024 {
			speedStr = fmt.Sprintf("%dKB/s", speed)
		} else {
			speedStr = fmt.Sprintf("%.1fMB/s", float64(speed)/1024)
		}
		tags = append(tags, speedStr)
	}

	if config.GlobalConfig.MediaCheck {
		// [修改] 更新正则表达式，加入 IPRisk 的百分比格式
		name = regexp.MustCompile(`\s*\|(?:GPT|GM|TW|ALL|\d+%)`).ReplaceAllString(name, "")
	}

	for _, plat := range config.GlobalConfig.Platforms {
		switch plat {
		case "openai":
			if res.Openai {
				tags = append(tags, "GPT")
			}
		case "gemini":
			if res.Gemini {
				tags = append(tags, "GM")
			}
		case "twitter":
			if res.Twitter {
				tags = append(tags, "TW")
			}
		case "combo":
			if res.Combo {
				tags = append(tags, "ALL")
			}
		case "iprisk": // [修改] 加回 IPRisk 的标签逻辑
			if res.IPRisk != "" {
				tags = append(tags, res.IPRisk)
			}
		}
	}

	// 将所有标记添加到名称中
	if len(tags) > 0 {
		name += " |" + strings.Join(tags, "|")
	}

	res.Proxy["name"] = name

}


// ... (从这里到文件末尾的所有代码都和之前一样，无需改动)
