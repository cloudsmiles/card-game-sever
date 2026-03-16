package main

import (
	"log"
	"net/http"

	"card-game-server/backend/config"
	"card-game-server/backend/internal/auth"
	"card-game-server/backend/internal/connection"
	"card-game-server/backend/internal/model"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("加载配置文件失败，使用默认配置: %v", err)
		// 允许无配置文件启动（向后兼容）
		cfg = &config.Config{
			Server: config.ServerConfig{Port: "8080"},
			JWT:    config.JWTConfig{Secret: "default-dev-secret", ExpireHour: 168},
		}
		config.Global = cfg
	}

	// 初始化数据库（如果配置了 MySQL）
	if cfg.MySQL.DBName != "" {
		if err := model.InitDB(cfg.MySQL.DSN()); err != nil {
			log.Printf("数据库连接失败，用户系统不可用: %v", err)
		}
	} else {
		log.Println("未配置 MySQL，用户系统不可用，仅支持游客模式")
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 配置 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置 HTTP 路由
	connection.SetupHTTPHandlers(r)

	// 设置认证路由（仅在数据库可用时）
	if model.DB != nil {
		auth.SetupRoutes(r)
	}

	// WebSocket 路由（保持原有方式，兼容 token 和 player 参数）
	http.HandleFunc("/ws", connection.WSHandler)

	// 将 gin 引擎作为 HTTP handler
	http.Handle("/", r)

	addr := ":" + cfg.Server.Port
	log.Printf("卡牌游戏服务器启动，监听 %s", addr)
	log.Printf("HTTP API: http://localhost:%s/api/rooms", cfg.Server.Port)
	log.Printf("Auth API: http://localhost:%s/api/auth/*", cfg.Server.Port)
	log.Printf("WebSocket: ws://localhost:%s/ws", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
