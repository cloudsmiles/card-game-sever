package interfaces

// TimeoutHandler 超时处理接口（游戏层实现，房间层调用）
type TimeoutHandler interface {
	HandleTurnTimeout(playerID string) error                                          // 处理回合操作超时
	HandlePendingTimeout() error                                                      // 处理等待响应超时
	SetTimeoutCallbacks(onTurnTimeout func(playerID string), onPendingTimeout func()) // 设置超时回调
}

// BotPlayer 机器人行为接口（游戏层可选实现）
// 实现此接口的游戏可以自定义机器人的决策逻辑，房间层无需关心具体游戏类型
type BotPlayer interface {
	// NeedsBotCheckAfterAction 是否在每次操作后都需要检查机器人
	// 返回 true 表示任何操作后都应检查所有机器人（如麻将的pending响应阶段）
	// 返回 false 表示只在轮到机器人时检查（如斗地主）
	NeedsBotCheckAfterAction() bool

	// GetBotAction 获取指定机器人的操作
	// 返回 nil 表示该机器人当前无需操作
	GetBotAction(botID string) *Action
}

// Game 是所有卡牌游戏必须实现的通用接口
type Game interface {
	ID() string
	Init(players []string) error                                     // 初始化，回合从第 0 位玩家开始
	ProcessAction(playerID string, action interface{}) (bool, error) // 处理卡牌指令，返回是否结束本回合
	CurrentTurn() string                                             // 当前回合玩家 ID
	GetState() interface{}                                           // 返回游戏状态（不包含敏感信息如其他玩家手牌）
	GetStateForPlayer(playerID string) interface{}                   // 返回指定玩家的游戏状态（包含该玩家的手牌等私密信息）
	IsGameOver() bool
	Winner() string
	MaxPlayers() int
	MinPlayers() int
}

// Action 统一包装卡牌指令
type Action struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
