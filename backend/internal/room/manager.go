package room

import (
	"fmt"
	"sync"
)

var (
	ErrRoomNotFound   = fmt.Errorf("room not found")
	ErrRoomFull       = fmt.Errorf("room is full")
	ErrNoGame         = fmt.Errorf("game not started")
	ErrAlreadyInRoom  = fmt.Errorf("player already in another room")
	ErrPlayerNotFound = fmt.Errorf("player not found in room")
)

// PlayerRoomTracker 追踪玩家所在的房间
type PlayerRoomTracker struct {
	playerRooms map[string]string // playerID -> roomID
	mu          sync.RWMutex
}

var GlobalPlayerTracker = &PlayerRoomTracker{playerRooms: make(map[string]string)}

// GetPlayerRoom 获取玩家当前所在的房间ID
func (pt *PlayerRoomTracker) GetPlayerRoom(playerID string) (string, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	roomID, exists := pt.playerRooms[playerID]
	return roomID, exists
}

// SetPlayerRoom 设置玩家所在的房间
func (pt *PlayerRoomTracker) SetPlayerRoom(playerID, roomID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.playerRooms[playerID] = roomID
}

// RemovePlayer 移除玩家的房间记录
func (pt *PlayerRoomTracker) RemovePlayer(playerID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	delete(pt.playerRooms, playerID)
}

type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

var GlobalManager = &Manager{rooms: make(map[string]*Room)}

func (m *Manager) CreateRoom(gameType string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := fmt.Sprintf("room_%d", len(m.rooms)+1)
	room := NewRoom(id, gameType)
	if room == nil {
		return "" // 返回空字符串表示创建失败
	}
	m.rooms[id] = room
	return id
}

func (m *Manager) GetRoom(id string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[id]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return r, nil
}

// RemoveRoom 从管理器中移除房间（房间为空时调用）
func (m *Manager) RemoveRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

// GetAllRooms 获取所有房间列表
func (m *Manager) GetAllRooms() []*Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}
	return rooms
}

// RoomInfo 房间简要信息（用于HTTP API）
type RoomInfo struct {
	ID        string `json:"id"`
	GameType  string `json:"game_type"`
	State     string `json:"state"`
	Players   int    `json:"players"`
	MaxPlayers int   `json:"max_players"`
}
