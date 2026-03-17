package connection

import (
	"card-game-server/backend/internal/types"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// playerConnection 存储玩家的连接信息
type playerConnection struct {
	conn *websocket.Conn
	send chan struct{} // 关闭此channel通知旧连接断开
}

// connectedPlayers 追踪已连接的玩家
var connectedPlayers = make(map[string]*playerConnection)
var connectedPlayersMu sync.RWMutex

// isPlayerConnected 检查玩家是否已连接
func isPlayerConnected(playerID string) bool {
	connectedPlayersMu.RLock()
	defer connectedPlayersMu.RUnlock()
	return connectedPlayers[playerID] != nil
}

// kickExistingConnection 踢掉已有连接，返回true表示踢掉了旧连接
func kickExistingConnection(playerID string) bool {
	connectedPlayersMu.Lock()
	defer connectedPlayersMu.Unlock()
	if existing, ok := connectedPlayers[playerID]; ok && existing != nil {
		log.Printf("踢掉玩家 [%s] 的旧连接", playerID)
		// 先发送被踢通知，让前端知道不要重连
		existing.conn.WriteJSON(types.Message{
			Type: types.Error,
			Data: types.ErrorData{Code: 4001, Message: "您的账号在其他地方登录"},
		})
		close(existing.send) // 通知旧连接关闭
		existing.conn.Close()
		delete(connectedPlayers, playerID)
		return true
	}
	return false
}

// registerPlayerConnection 注册玩家连接
func registerPlayerConnection(playerID string, conn *websocket.Conn) <-chan struct{} {
	connectedPlayersMu.Lock()
	defer connectedPlayersMu.Unlock()
	kicked := make(chan struct{})
	connectedPlayers[playerID] = &playerConnection{conn: conn, send: kicked}
	return kicked
}

// unregisterPlayerConnection 注销玩家连接（仅当连接匹配时才注销）
func unregisterPlayerConnection(playerID string, conn *websocket.Conn) {
	connectedPlayersMu.Lock()
	defer connectedPlayersMu.Unlock()
	if existing, ok := connectedPlayers[playerID]; ok && existing != nil && existing.conn == conn {
		delete(connectedPlayers, playerID)
	}
}

// heartbeat 心跳检测，60秒无活动则关闭连接
func heartbeat(conn *websocket.Conn, playerID string) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				conn.Close()
				return
			}
		}
	}()

	return done
}
