package interfaces

// Game 是所有卡牌游戏必须实现的通用接口
type Game interface {
	ID() string
	Init(players []string) error                                     // 初始化，回合从第0位玩家开始
	ProcessAction(playerID string, action interface{}) (bool, error) // 处理卡牌指令，返回是否结束本回合
	CurrentTurn() string                                             // 当前回合玩家ID
	AdvanceTurn()                                                    // 切换到下一个玩家
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
