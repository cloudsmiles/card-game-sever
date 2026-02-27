package types

// 动作名称枚举（仅用于游戏动作）
type GameActionType string

const (
	PlayCard     GameActionType = "play_card"     // 简单游戏出牌
	PlayCards    GameActionType = "play_cards"    // 斗地主出牌
	Pass         GameActionType = "pass"          // 跳过
	CallLandlord GameActionType = "call_landlord" // 叫地主
)
