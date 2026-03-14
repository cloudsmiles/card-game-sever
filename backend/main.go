package main

import (
	"log"
	"net/http"

	"card-game-server/backend/internal/connection"

	"github.com/gin-gonic/gin"
)

func main() {
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

	// 使用 gin 处理 HTTP 请求
	httpHandler := r

	// WebSocket 路由
	http.HandleFunc("/ws", connection.WSHandler)

	// 将 gin 引擎作为 HTTP handler
	http.Handle("/", httpHandler)

	log.Println("卡牌游戏服务器启动，监听 :8080")
	log.Println("HTTP API: http://localhost:8080/api/rooms")
	log.Println("WebSocket: ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
