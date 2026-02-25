package interfaces

// Game 是所有卡牌游戏必须实现的通用接口
type Game interface {
	ID() string
	Init(players []string) error                                     // 初始化，回合从第0位玩家开始
	ProcessAction(playerID string, action interface{}) (bool, error) // 处理卡牌指令，返回是否结束本回合
	CurrentTurn() string                                             // 当前回合玩家ID
	AdvanceTurn()                                                    // 切换到下一个玩家
	GetState() interface{}                                           // 返回完整游戏状态（用于同步给所有客户端）
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
