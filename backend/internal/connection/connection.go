package connection

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// connectedPlayers 追踪已连接的玩家（防止同一playerID多次连接）
var connectedPlayers = make(map[string]bool)
var connectedPlayersMu sync.RWMutex

// isPlayerConnected 检查玩家是否已连接
func isPlayerConnected(playerID string) bool {
	connectedPlayersMu.RLock()
	defer connectedPlayersMu.RUnlock()
	return connectedPlayers[playerID]
}

// setPlayerConnected 设置玩家连接状态
func setPlayerConnected(playerID string, connected bool) {
	connectedPlayersMu.Lock()
	defer connectedPlayersMu.Unlock()
	if connected {
		connectedPlayers[playerID] = true
	} else {
		delete(connectedPlayers, playerID)
	}
}

// heartbeat 心跳检测，60秒无活动则关闭连接
// 返回一个channel，当心跳检测到超时时会关闭该channel
func heartbeat(conn *websocket.Conn, playerID string) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		// 设置读取超时和心跳检测
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// 每30秒发送一次ping
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				// 发送ping失败，关闭连接并退出
				conn.Close()
				return
			}
		}
	}()

	return done
}
