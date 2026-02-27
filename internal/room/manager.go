package room

import (
	"fmt"
	"log"
	"sync"
)

var (
	ErrRoomNotFound = fmt.Errorf("room not found")
	ErrRoomFull     = fmt.Errorf("room is full")
	ErrNoGame       = fmt.Errorf("game not started")
)

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
	log.Printf("房间 [%s] 已从管理器移除", id)
}
