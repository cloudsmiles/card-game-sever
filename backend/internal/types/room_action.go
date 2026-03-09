package types

// RoomActionType 房间内操作类型
type RoomActionType string

const (
	RoomActionReady RoomActionType = "ready" // 准备/取消准备
	RoomActionSit   RoomActionType = "sit"   // 选择座位
)
