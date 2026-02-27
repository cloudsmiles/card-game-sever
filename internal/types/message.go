package types

// 房间状态
type RoomState string

const (
	RoomWaiting   RoomState = "waiting"   // 等待中（可加入、可准备）
	RoomPlaying   RoomState = "playing"   // 游戏中
	RoomPaused    RoomState = "paused"    // 游戏暂停（有玩家断线）
	RoomGameOver  RoomState = "gameover"  // 游戏结束
)

// 消息类型枚举
type MsgType string

const (
	// 房间管理（RoomManager 级别操作）
	RoomCreate MsgType = "room.create" // 创建房间
	RoomJoin   MsgType = "room.join"   // 加入房间
	RoomLeave  MsgType = "room.leave"  // 离开房间

	// 房间内部操作（进入房间后的操作）
	RoomAction MsgType = "room.action" // 房间内操作：准备、选座等
	GameAction MsgType = "game.action" // 游戏操作：出牌、叫分等
	Chat       MsgType = "chat"        // 聊天

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

// GameActionData 统一游戏动作数据结构
type GameActionData struct {
	Action GameActionType `json:"action"`
	Card   interface{}    `json:"card,omitempty"`
}

// RoomActionData 统一房间内操作数据结构
type RoomActionData struct {
	Action RoomActionType `json:"action"` // ready, sit
	Data   interface{}    `json:"data"`   // 根据 action 类型不同
}

// RoomActionReadyData 准备操作数据
type RoomActionReadyData struct {
	Ready bool `json:"ready"` // true=准备, false=取消准备
}

// RoomActionSitData 选座操作数据
type RoomActionSitData struct {
	SeatNumber int `json:"seat_number"` // 座位号 0, 1, 2
}

type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
