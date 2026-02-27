package connection

import "sync"

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
