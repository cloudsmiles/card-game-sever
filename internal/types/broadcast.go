package types

// 事件名称枚举（仅用于广播消息）
type Event string

const (
	RoomStateChanged Event = "room_state_changed"
	PlayerJoined     Event = "player_joined"
	PlayerLeft       Event = "player_left"
	GameStarted      Event = "game_started"
	GameFinished     Event = "game_over"
	StateUpdate      Event = "state_update"
	ChatMessage      Event = "chat"
)

type JoinRoomContent struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
	Message  string `json:"message"`
}

type LeftRoomContent struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
	Message  string `json:"message"`
}

type ChatContent struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
	Content  string `json:"content"`
}
