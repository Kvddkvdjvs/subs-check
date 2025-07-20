package platform

import (
	"net/http"
	"strings"
)

// CheckCombo 检查节点满足了哪些平台要求，并返回一个组合标签
func CheckCombo(client *http.Client) string {
	var passed []string // 用于存储通过测试的平台标签

	// 测试 OpenAI
	if ok, err := CheckOpenai(client); err == nil && ok {
		passed = append(passed, "GPT")
	}

	// 测试 Gemini
	if ok, err := CheckGemini(client); err == nil && ok {
		passed = append(passed, "GM")
	}

	// 测试 Twitter
	if ok, err := CheckTwitter(client); err == nil && ok {
		passed = append(passed, "TW")
	}

	return strings.Join(passed, "|")
}
