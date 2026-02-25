package room

import (
	"fmt"
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
	m.rooms[id] = NewRoom(id, gameType)
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
