package handler

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(赌博|赌场|博彩|下注|盘口)`),
	regexp.MustCompile(`(?i)(色情|porn|成人|约炮|一夜情|援交)`),
	regexp.MustCompile(`(?i)(枪支|弹药|炸药|炸弹|恐怖)`),
	regexp.MustCompile(`(?i)(毒品|吸毒|贩毒|大麻|海洛因)`),
	regexp.MustCompile(`(?i)(传销|拉人头|下线|金字塔)`),
	regexp.MustCompile(`(?i)(高利贷|裸贷|校园贷|套路贷)`),
	regexp.MustCompile(`(?i)(法轮功|台独|藏独|疆独|港独)`),
}

type HardPolicyResult struct {
	Passed bool
	Reason string
}

func CheckHardPolicy(content string) HardPolicyResult {
	if utf8.RuneCountInString(strings.TrimSpace(content)) < 10 {
		return HardPolicyResult{Passed: false, Reason: "内容过短，请至少输入10个字符"}
	}

	nonSymbol := regexp.MustCompile(`[\p{Han}a-zA-Z]`)
	if !nonSymbol.MatchString(content) {
		return HardPolicyResult{Passed: false, Reason: "内容不能全是数字或符号，请写一段有意义的话"}
	}

	for _, p := range sensitivePatterns {
		if p.MatchString(content) {
			return HardPolicyResult{Passed: false, Reason: "内容包含不适当的关键词，请修改后重新提交"}
		}
	}

	if hasExcessiveRepetition(content) {
		return HardPolicyResult{Passed: false, Reason: "内容包含过多重复字符"}
	}

	return HardPolicyResult{Passed: true}
}

func hasExcessiveRepetition(s string) bool {
	runes := []rune(s)
	if len(runes) < 10 {
		return false
	}
	count := 1
	for i := 1; i < len(runes); i++ {
		if runes[i] == runes[i-1] {
			count++
			if count > 10 {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}
