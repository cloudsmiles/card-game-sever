package types

// 事件名称枚举（仅用于广播消息）
type Event string

const (
	// 房间级别事件（包含玩家加入/离开/准备/选座等所有房间状态变化）
	RoomStateChanged Event = "room_state_changed"

	// 游戏级别事件
	GameStarted  Event = "game_started"
	GameFinished Event = "game_over"
	StateUpdate  Event = "state_update" // 游戏状态更新（出牌等）
	ChatMessage  Event = "chat"
)

type BroadcastData struct {
	Event   Event       `json:"event"`
	Content interface{} `json:"content"`
}

// PersonalizedBroadcastData 个性化广播数据（用于通过channel发送个性化消息）
type PersonalizedBroadcastData struct {
	Event       Event
	ContentFunc func(playerID string) interface{}
}

type ChatContent struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
	Content  string `json:"content"`
}

// RoomStateContent 房间状态变更广播（包含玩家加入/离开/准备/选座等所有变化）
type RoomStateContent struct {
	RoomID  string           `json:"room_id"`
	State   RoomState        `json:"state"`
	Players []PlayerSeatInfo `json:"players"`
	Message string           `json:"message"`
}

// PlayerSeatInfo 玩家座位信息
type PlayerSeatInfo struct {
	PlayerID   string `json:"player_id"`
	SeatNumber int    `json:"seat_number"`
	Ready      bool   `json:"ready"`
	IsOffline  bool   `json:"is_offline"` // 是否断线
}
