package main

import (
	"log"
	"net/http"

	"card-game-server/internal/connection"
)

func main() {
	http.HandleFunc("/ws", connection.WSHandler)
	log.Println("卡牌游戏服务器启动，监听 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
