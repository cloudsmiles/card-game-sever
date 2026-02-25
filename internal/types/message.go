package types

// 消息类型枚举
type MsgType string

const (
	// 客户端请求
	CreateRoom MsgType = "create_room" // 对应 CreateRoomData
	JoinRoom   MsgType = "join_room"   // 对应 JoinRoomData
	LeaveRoom  MsgType = "leave_room"  // 对应空数据
	GameAction MsgType = "game_action" // 对应 GameActionData
	Chat       MsgType = "chat"        // 对应 ChatData

	// 服务端响应/广播
	Broadcast MsgType = "broadcast" // 对应 BroadcastData
	Error     MsgType = "error"     // 对应 ErrorData
)

// 统一的消息结构
type Message struct {
	Type     MsgType     `json:"type"`
	RoomID   string      `json:"room_id,omitempty"`
	PlayerID string      `json:"player_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

type CreateRoomData struct {
	GameType string `json:"game_type"`
}

type JoinRoomData struct {
	RoomID string `json:"room_id"`
}

type ChatData struct {
	Content string `json:"content"`
}

type BroadcastData struct {
	Event   Event       `json:"event"`
	Content interface{} `json:"content"`
}

type GameActionData struct {
	Action Action      `json:"action"`
	Card   interface{} `json:"card,omitempty"`
}

type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
