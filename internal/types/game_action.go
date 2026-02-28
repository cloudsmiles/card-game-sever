package types

// 动作名称枚举（仅用于游戏动作）
type GameActionType string

const (
	PlayCard     GameActionType = "play_card"     // 简单游戏出牌
	PlayCards    GameActionType = "play_cards"    // 斗地主出牌
	Pass         GameActionType = "pass"          // 跳过
	CallLandlord GameActionType = "call_landlord" // 叫地主

	// 麻将动作
	Discard GameActionType = "discard" // 麻将出牌
	Chow    GameActionType = "chow"    // 吃
	Pong    GameActionType = "pong"    // 碰
	Kong    GameActionType = "kong"    // 杠
	Win     GameActionType = "win"     // 胡牌
)
