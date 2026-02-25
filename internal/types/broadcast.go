package types

// 事件名称枚举（仅用于广播消息）
type Event string

const (
	RoomCreated      Event = "room_created"
	RoomStateChanged Event = "room_state_changed"
	PlayerJoined     Event = "player_joined"
	PlayerLeft       Event = "player_left"
	GameStarted      Event = "game_started"
	GameFinished     Event = "game_over"
	StateUpdate      Event = "state_update"
	ChatMessage      Event = "chat"
)

// BroadcastData.Content
type CreateRoomContent struct {
	RoomID  string `json:"room_id"`
	Message string `json:"message"`
}

type JoinRoomContent struct {
	RoomID  string `json:"room_id"`
	Message string `json:"message"`
}

type ChatContent struct {
	PlayerID string `json:"player_id"`
	Content  string `json:"content"`
}
