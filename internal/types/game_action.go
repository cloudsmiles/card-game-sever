package types

// 动作名称枚举（仅用于游戏动作）
type Action string

const (
	PlayCard     Action = "play_card"     // 简单游戏出牌
	PlayCards    Action = "play_cards"    // 斗地主出牌
	Pass         Action = "pass"          // 跳过
	CallLandlord Action = "call_landlord" // 叫地主
)
