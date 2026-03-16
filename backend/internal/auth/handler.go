package auth

import (
	"card-game-server/backend/config"
	"card-game-server/backend/internal/model"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginResponse 登录响应
type LoginResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

// GuestLoginRequest 游客登录请求
type GuestLoginRequest struct {
	Nickname string `json:"nickname" binding:"required,min=1,max=20"`
}

// WeChatLoginRequest 微信登录请求
type WeChatLoginRequest struct {
	Code     string `json:"code" binding:"required"`
	Platform string `json:"platform"` // "h5" 或 "open"（PC扫码），默认 "h5"
}

// SetupRoutes 注册认证相关路由
func SetupRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/guest", GuestLoginHandler)
		auth.POST("/wechat", WeChatLoginHandler)
		auth.GET("/me", AuthMiddleware(), MeHandler)
	}
}

// GuestLoginHandler 游客登录
func GuestLoginHandler(c *gin.Context) {
	var req GuestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "昵称不能为空且不超过20字符"})
		return
	}

	// 创建游客用户
	user := &model.User{
		Nickname:  req.Nickname,
		LoginType: "guest",
	}
	if err := model.CreateUser(user); err != nil {
		log.Printf("创建游客用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败"})
		return
	}

	// 签发 token
	token, err := GenerateToken(user.ID, user.Nickname, user.LoginType)
	if err != nil {
		log.Printf("签发 token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "登录失败"})
		return
	}

	log.Printf("游客登录成功 [ID: %d, 昵称: %s]", user.ID, user.Nickname)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data":    LoginResponse{Token: token, User: *user},
	})
}

// WeChatLoginHandler 微信登录
func WeChatLoginHandler(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少 code 参数"})
		return
	}

	// 根据平台选择 AppID/AppSecret
	cfg := config.Global
	var appID, appSecret, loginType string
	if req.Platform == "open" {
		appID = cfg.WeChat.OpenAppID
		appSecret = cfg.WeChat.OpenAppSecret
		loginType = "wechat_open"
	} else {
		appID = cfg.WeChat.H5AppID
		appSecret = cfg.WeChat.H5AppSecret
		loginType = "wechat_h5"
	}

	if appID == "" || appSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("微信 %s 登录未配置", req.Platform)})
		return
	}

	// 用 code 换 access_token
	tokenResp, err := GetWeChatAccessToken(appID, appSecret, req.Code)
	if err != nil {
		log.Printf("微信授权失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "微信授权失败: " + err.Error()})
		return
	}

	// 获取用户信息
	userInfo, err := GetWeChatUserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	if err != nil {
		log.Printf("获取微信用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败"})
		return
	}

	// 查找或创建用户
	user, err := model.FindOrCreateWeChatUser(
		tokenResp.OpenID,
		userInfo.UnionID,
		userInfo.Nickname,
		userInfo.HeadImgURL,
		loginType,
	)
	if err != nil {
		log.Printf("创建/更新微信用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户处理失败"})
		return
	}

	// 签发 token
	token, err := GenerateToken(user.ID, user.Nickname, user.LoginType)
	if err != nil {
		log.Printf("签发 token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "登录失败"})
		return
	}

	log.Printf("微信用户登录成功 [ID: %d, 昵称: %s, 类型: %s]", user.ID, user.Nickname, loginType)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data":    LoginResponse{Token: token, User: *user},
	})
}

// MeHandler 获取当前用户信息
func MeHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	user, err := model.FindUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    user,
	})
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
			c.Abort()
			return
		}
		claims, err := ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "token 无效或已过期"})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("nickname", claims.Nickname)
		c.Set("loginType", claims.LoginType)
		c.Next()
	}
}
