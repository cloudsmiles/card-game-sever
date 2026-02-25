package types

type MessageType string

const (
	MsgCreateRoom MessageType = "create_room"
	MsgJoinRoom   MessageType = "join_room"
	MsgLeaveRoom  MessageType = "leave_room"
	MsgChat       MessageType = "chat"
	MsgGameAction MessageType = "game_action"
	MsgBroadcast  MessageType = "broadcast" // 服务端推送
	MsgError      MessageType = "error"
)

type Message struct {
	Type     MessageType `json:"type"`
	RoomID   string      `json:"room_id,omitempty"`
	PlayerID string      `json:"player_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

type CreateRoomData struct {
	GameType string `json:"game_type"` // e.g. "simple"
}

type JoinRoomData struct {
	RoomID string `json:"room_id"`
}

type ChatData struct {
	Content string `json:"content"`
}

// 在 GameActionData 添加
type GameActionData struct {
	Action string      `json:"action"`         // "play_card", "call_landlord", "pass"
	Card   interface{} `json:"card,omitempty"` // 卡牌 or 分数
}

type BroadcastData struct {
	Event   string      `json:"event"`
	Content interface{} `json:"content"`
}
