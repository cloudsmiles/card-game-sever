package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// WeChatTokenResponse 微信 access_token 响应
type WeChatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// WeChatUserInfo 微信用户信息
type WeChatUserInfo struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetWeChatAccessToken 用 code 换取 access_token
func GetWeChatAccessToken(appID, appSecret, code string) (*WeChatTokenResponse, error) {
	u := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(appID), url.QueryEscape(appSecret), url.QueryEscape(code),
	)
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("请求微信 access_token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取微信响应失败: %w", err)
	}

	var tokenResp WeChatTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信授权失败: [%d] %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}
	return &tokenResp, nil
}

// GetWeChatUserInfo 获取微信用户信息
func GetWeChatUserInfo(accessToken, openID string) (*WeChatUserInfo, error) {
	u := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(accessToken), url.QueryEscape(openID),
	)
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("请求微信用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取微信响应失败: %w", err)
	}

	var userInfo WeChatUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析微信用户信息失败: %w", err)
	}
	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("获取微信用户信息失败: [%d] %s", userInfo.ErrCode, userInfo.ErrMsg)
	}
	return &userInfo, nil
}
