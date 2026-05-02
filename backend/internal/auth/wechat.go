package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// WechatAccessTokenResp 微信 access_token 响应
type WechatAccessTokenResp struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
	UnionID     string `json:"unionid"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// WechatUserInfo 微信用户信息
type WechatUserInfo struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	Nickname   string `json:"nickname"`
	HeadImgURL string `json:"headimgurl"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// ExchangeCode 用 code 换取 access_token 和 openid
func ExchangeCode(appID, appSecret, code string) (*WechatAccessTokenResp, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(appID),
		url.QueryEscape(appSecret),
		url.QueryEscape(code),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("wechat api error: %w", err)
	}
	defer resp.Body.Close()

	var result WechatAccessTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error [%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// GetUserInfo 获取微信用户信息
func GetUserInfo(accessToken, openID string) (*WechatUserInfo, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		url.QueryEscape(accessToken),
		url.QueryEscape(openID),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("wechat api error: %w", err)
	}
	defer resp.Body.Close()

	var result WechatUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error [%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
