package durian

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"card-game-server/backend/internal/game/interfaces"
)

// DurianGame 榴莲忘返游戏，实现 interfaces.Game 接口
type DurianGame struct {
	players            []string                // 玩家ID列表（顺序即行动顺序）
	phase              string                  // 游戏阶段
	deck               *Deck                   // 牌堆
	holderCards        map[string]Card         // 每位玩家的牌架卡 playerID -> Card
	orders             map[FruitType]int       // 当前订单区（每种水果的需求数量）
	tokenPool          *AngerTokenPool         // 公共愤怒标记池
	playerAnger        map[string]*PlayerAnger // 每位玩家的愤怒标记
	currentTurn        int                     // 当前行动玩家在 players 中的索引
	lastOrderPlayerIdx int                     // 上一个接订单的玩家索引（-1 表示无人接过订单）
	roundNumber        int                     // 当前轮次编号
	flippedCard        Card                    // 当前翻开等待处理的牌（take_order 时用）
	gameOver           bool                    // 游戏是否结束
	winner             string                  // 获胜者
	lastSettlement     *SettlementResult       // 最近一次结算结果（用于状态广播）
	playedCardsHistory []PlayedCardRecord      // 已打出的牌历史记录

	// 等待确认继续相关
	waitingForContinue    map[string]bool // 等待哪些玩家确认继续
	pendingNextRoundStart int             // 待开始新一轮的首个玩家索引

	// 猩猩牌交换订单相关
	waitingForSwapOrder bool // 是否正在等待玩家选择交换订单

	// 超时相关
	turnTimer        *time.Timer
	turnDeadline     time.Time
	onTurnTimeout    func(playerID string)
	onPendingTimeout func()
}

// PlayedCardRecord 已打出的牌记录
type PlayedCardRecord struct {
	PlayerName  string `json:"player_name"`
	Card        Card   `json:"card"`
	ChosenSide  string `json:"chosen_side"`
	Timestamp   string `json:"timestamp"`
	RoundNumber int    `json:"round_number"`
	Swapped     bool   `json:"swapped"` // 是否已被猩猩牌翻转过（每张牌只能翻转一次）
}

// 确保实现接口
var _ interfaces.Game = (*DurianGame)(nil)
var _ interfaces.BotPlayer = (*DurianGame)(nil)
var _ interfaces.TimeoutHandler = (*DurianGame)(nil)

// New 创建一个新的榴莲忘返游戏实例
func New() interfaces.Game {
	return &DurianGame{}
}

// ─────────────────────────────────────────────
// interfaces.Game 接口实现
// ─────────────────────────────────────────────

func (g *DurianGame) ID() string       { return "durian" }
func (g *DurianGame) MinPlayers() int  { return 2 }
func (g *DurianGame) MaxPlayers() int  { return 7 }
func (g *DurianGame) IsGameOver() bool { return g.gameOver }
func (g *DurianGame) Winner() string   { return g.winner }

// Init 初始化游戏
func (g *DurianGame) Init(players []string) error {
	// 过滤空字符串玩家（房间可能有空座位）
	validPlayers := make([]string, 0, len(players))
	for _, p := range players {
		if p != "" {
			validPlayers = append(validPlayers, p)
		}
	}

	if len(validPlayers) < g.MinPlayers() || len(validPlayers) > g.MaxPlayers() {
		return fmt.Errorf("榴莲忘返需要 %d~%d 名玩家，当前 %d 名",
			g.MinPlayers(), g.MaxPlayers(), len(validPlayers))
	}

	rand.Seed(time.Now().UnixNano())

	g.players = validPlayers
	g.phase = "dealing"
	g.tokenPool = NewAngerTokenPool()
	g.playerAnger = make(map[string]*PlayerAnger, len(validPlayers))
	for _, pid := range validPlayers {
		g.playerAnger[pid] = &PlayerAnger{PlayerID: pid}
	}
	g.orders = map[FruitType]int{
		FruitDurian:     0,
		FruitBanana:     0,
		FruitGrape:      0,
		FruitStrawberry: 0,
	}
	g.lastOrderPlayerIdx = -1
	g.roundNumber = 0
	g.gameOver = false
	g.winner = ""
	g.lastSettlement = nil
	g.playedCardsHistory = make([]PlayedCardRecord, 0)

	// 创建并洗牌
	g.deck = NewDeck()
	g.deck.Shuffle()

	// 开始第一轮
	return g.startNewRound(0)
}

// CurrentTurn 返回当前行动玩家的 ID
func (g *DurianGame) CurrentTurn() string {
	if len(g.players) == 0 {
		return ""
	}
	return g.players[g.currentTurn]
}

// AdvanceTurn 切换到下一位玩家（由 room.go 在 turnEnded=true 时调用）
func (g *DurianGame) AdvanceTurn() {
	g.currentTurn = (g.currentTurn + 1) % len(g.players)
	// Only start timer during playing phase
	if g.phase == "playing" {
		g.startTurnTimer()
	}
}

// ProcessAction 处理玩家行动
// action 实际类型为 interfaces.Action
func (g *DurianGame) ProcessAction(playerID string, action interface{}) (bool, error) {
	if g.gameOver {
		return false, fmt.Errorf("游戏已结束")
	}

	act, ok := action.(interfaces.Action)
	if !ok {
		return false, fmt.Errorf("无效的 action 类型，期望 interfaces.Action")
	}

	// 处理确认继续（在等待确认阶段，任何玩家都可以操作）
	if act.Type == "continue" {
		turnEnded, err := g.processContinue(playerID)
		// 如果回合结束，游戏层自己推进
		if turnEnded {
			g.AdvanceTurn()
		}
		return turnEnded, err
	}

	// 其他操作需要是当前回合玩家
	if g.CurrentTurn() != playerID {
		return false, fmt.Errorf("不是你的回合，当前轮到 %s", g.CurrentTurn())
	}

	var turnEnded bool
	var err error

	switch act.Type {
	case "take_order":
		turnEnded, err = g.processTakeOrder(playerID, act)
	case "ring_bell":
		turnEnded, err = g.processRingBell(playerID)
	default:
		return false, fmt.Errorf("未知的行动类型: %s，支持 take_order、ring_bell 或 continue", act.Type)
	}

	// 如果回合结束，游戏层自己推进
	if turnEnded {
		g.AdvanceTurn()
	}

	return turnEnded, err
}

// GetState 返回公共游戏状态（不含任何玩家私密信息）
func (g *DurianGame) GetState() interface{} {
	return g.buildPublicState()
}

// GetStateForPlayer 返回指定玩家视角的游戏状态
// 核心安全原则：your_card 永远为 nil，不泄露自己的牌架卡
func (g *DurianGame) GetStateForPlayer(playerID string) interface{} {
	state := g.buildPublicState()

	// 构建其他玩家的牌架卡信息（绝不包含自己的牌）
	othersCards := make([]map[string]interface{}, 0, len(g.players)-1)
	for _, pid := range g.players {
		if pid == playerID {
			continue // 跳过自己，永远不暴露自己的牌
		}
		if card, exists := g.holderCards[pid]; exists {
			othersCards = append(othersCards, map[string]interface{}{
				"player_id": pid,
				"card":      cardToMap(card),
			})
		}
	}

	state["your_card"] = nil // 永远为 nil，自己看不到自己的牌
	state["others_cards"] = othersCards

	return state
}

// ─────────────────────────────────────────────
// 内部游戏逻辑
// ─────────────────────────────────────────────

// startNewRound 开始新一轮游戏
// firstPlayerIdx 为本轮第一个行动玩家的索引
func (g *DurianGame) startNewRound(firstPlayerIdx int) error {
	g.roundNumber++
	g.currentTurn = firstPlayerIdx % len(g.players)
	g.lastOrderPlayerIdx = -1
	g.flippedCard = nil

	// 重置订单区
	g.orders = map[FruitType]int{
		FruitDurian:     0,
		FruitBanana:     0,
		FruitGrape:      0,
		FruitStrawberry: 0,
	}

	// 清空历史记录（新一轮开始）
	g.playedCardsHistory = make([]PlayedCardRecord, 0)

	// 清空上一轮结算结果
	g.lastSettlement = nil

	// 重新洗牌（上一轮的牌架卡放回牌堆后重新洗牌）
	// 先将旧牌架卡放回废牌堆
	if g.holderCards != nil {
		for _, card := range g.holderCards {
			g.deck.Discard(card)
		}
	}
	g.deck.ReshuffleAll()

	// 重新为每位玩家发一张牌架卡
	g.holderCards = make(map[string]Card, len(g.players))
	for _, pid := range g.players {
		card, ok := g.deck.Draw()
		if !ok {
			return fmt.Errorf("牌堆和废牌堆均已耗尽，无法发牌")
		}
		g.holderCards[pid] = card
	}

	g.phase = "playing"
	// 重置游戏结束状态（重要：开始新回合时必须重置）
	g.gameOver = false
	g.winner = ""
	return nil
}

// processTakeOrder 处理"接新订单"行动
// 先翻牌显示给玩家，玩家选择左右面后才真正处理
func (g *DurianGame) processTakeOrder(playerID string, act interfaces.Action) (bool, error) {
	if g.phase != "playing" {
		return false, fmt.Errorf("当前阶段 %s 不能接订单", g.phase)
	}

	// 检查是否正在等待交换订单（猩猩牌效果）
	if g.waitingForSwapOrder {
		return g.processSwapOrder(playerID, act)
	}

	// 检查是否已有翻开的牌等待处理
	if g.flippedCard == nil {
		// 第一步：从牌堆翻一张牌，等待玩家选择
		// 如果没有订单历史，猩猩牌无法触发交换，跳过
		for {
			card, ok := g.deck.Draw()
			if !ok {
				return false, fmt.Errorf("牌堆已空，无法翻牌")
			}

			// 检查是否是猩猩牌且订单历史为空（无牌可交换）
			if _, isGorilla := card.(*GorillaCard); isGorilla && len(g.playedCardsHistory) == 0 {
				// 废弃猩猩牌，继续翻下一张
				g.deck.Discard(card)
				continue
			}

			g.flippedCard = card
			break
		}

		// 清空上一轮的结算结果（新一轮第一个行动时）
		g.lastSettlement = nil

		// 猩猩牌效果：进入交换订单阶段，然后替换牌架卡
		if _, isGorilla := g.flippedCard.(*GorillaCard); isGorilla {
			// 检查是否有可交换的牌（未被翻转过的水果牌）
			hasSwappable := false
			for _, record := range g.playedCardsHistory {
				if _, ok := record.Card.(*FruitCard); ok && !record.Swapped {
					hasSwappable = true
					break
				}
			}
			if hasSwappable {
				// 进入等待交换订单状态
				g.waitingForSwapOrder = true
				return false, nil
			}
			// 没有可交换的牌，直接废弃猩猩牌
			g.deck.Discard(g.flippedCard)
			g.flippedCard = nil
			return true, nil
		}

		// 翻牌成功，等待玩家选择，不结束回合
		return false, nil
	}

	// 第二步：处理已翻开的牌（玩家已选择左右面）
	// 解析 chosen_side
	chosenSide, err := parseChosenSide(act.Data)
	if err != nil {
		return false, err
	}

	card := g.flippedCard

	// 处理翻出的水果牌
	if c, ok := card.(*FruitCard); ok {
		// 根据选择的面添加订单
		switch chosenSide {
		case "left":
			g.orders[c.LeftFruit] += c.LeftCount
		case "right":
			g.orders[c.RightFruit] += c.RightCount
		}
		// 翻开的牌废弃（不放入牌架）
		g.deck.Discard(card)
	}

	// 记录本次接订单的玩家
	g.lastOrderPlayerIdx = g.currentTurn

	// 记录到历史
	record := PlayedCardRecord{
		PlayerName:  playerID,
		Card:        card,
		ChosenSide:  chosenSide,
		Timestamp:   time.Now().Format("15:04:05"),
		RoundNumber: g.roundNumber,
	}
	g.playedCardsHistory = append(g.playedCardsHistory, record)

	// 清空翻开的牌
	g.flippedCard = nil

	// 返回 true，表示回合结束（由 ProcessAction 统一调用 AdvanceTurn）
	return true, nil
}

// processSwapOrder 处理猩猩牌的交换订单效果
// 玩家选择历史记录中的一张牌，交换其左右水果
func (g *DurianGame) processSwapOrder(playerID string, act interfaces.Action) (bool, error) {
	// 解析选择的订单索引
	dataMap, ok := act.Data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("无效的交换订单数据格式")
	}

	orderIndexVal, exists := dataMap["order_index"]
	if !exists {
		return false, fmt.Errorf("缺少 order_index 参数")
	}

	var orderIndex int
	switch v := orderIndexVal.(type) {
	case float64:
		orderIndex = int(v)
	case int:
		orderIndex = v
	default:
		return false, fmt.Errorf("order_index 必须是数字")
	}

	// 验证索引合法性
	if orderIndex < 0 || orderIndex >= len(g.playedCardsHistory) {
		return false, fmt.Errorf("无效的订单索引 %d，有效范围 0~%d", orderIndex, len(g.playedCardsHistory)-1)
	}

	// 获取要交换的记录
	record := &g.playedCardsHistory[orderIndex]
	fruitCard, ok := record.Card.(*FruitCard)
	if !ok {
		return false, fmt.Errorf("只能交换水果牌")
	}

	// 检查是否已被翻转过（每张牌只能翻转一次）
	if record.Swapped {
		return false, fmt.Errorf("这张牌已经被翻转过了")
	}

	// 计算该记录在订单区中的贡献
	oldLeftFruit := fruitCard.LeftFruit
	oldLeftCount := fruitCard.LeftCount
	oldRightFruit := fruitCard.RightFruit
	oldRightCount := fruitCard.RightCount

	// 从订单区中移除原来的贡献
	if record.ChosenSide == "left" {
		g.orders[oldLeftFruit] -= oldLeftCount
	} else {
		g.orders[oldRightFruit] -= oldRightCount
	}

	// 交换牌的左右水果
	fruitCard.LeftFruit, fruitCard.RightFruit = fruitCard.RightFruit, fruitCard.LeftFruit
	fruitCard.LeftCount, fruitCard.RightCount = fruitCard.RightCount, fruitCard.LeftCount

	// 将交换后的另一面加入订单区
	if record.ChosenSide == "left" {
		// 原来选了左面，现在左面变成了原来的右面
		g.orders[fruitCard.LeftFruit] += fruitCard.LeftCount
	} else {
		// 原来选了右面，现在右面变成了原来的左面
		g.orders[fruitCard.RightFruit] += fruitCard.RightCount
	}

	// 确保订单数量不为负
	for fruit := range g.orders {
		if g.orders[fruit] < 0 {
			g.orders[fruit] = 0
		}
	}

	// 标记该牌已被翻转
	record.Swapped = true

	// 猩猩牌废弃（不替换牌架卡）
	g.deck.Discard(g.flippedCard)
	g.flippedCard = nil
	g.waitingForSwapOrder = false

	// 返回 true，表示回合结束（由 ProcessAction 统一调用 AdvanceTurn）
	return true, nil
}

// processContinue 处理玩家确认继续（结算后等待所有人确认）
func (g *DurianGame) processContinue(playerID string) (bool, error) {
	if g.phase != "waiting_continue" {
		return false, fmt.Errorf("当前阶段 %s 不需要确认继续", g.phase)
	}

	// 检查玩家是否在等待列表中
	if !g.waitingForContinue[playerID] {
		return false, nil // 已经确认过了，忽略
	}

	// 标记该玩家已确认
	g.waitingForContinue[playerID] = false

	// 检查是否所有玩家都已确认
	allConfirmed := true
	for _, waiting := range g.waitingForContinue {
		if waiting {
			allConfirmed = false
			break
		}
	}

	// 所有人都确认了，开始新一轮
	if allConfirmed {
		if err := g.startNewRound(g.pendingNextRoundStart); err != nil {
			return false, fmt.Errorf("开始新一轮失败：%w", err)
		}
		// 清空等待状态
		g.waitingForContinue = nil
		g.pendingNextRoundStart = 0

		// 返回 true 表示回合结束（由 ProcessAction 统一调用 AdvanceTurn）
		// 但实际的新回合起始玩家已在 startNewRound 中设置
		return true, nil
	}

	return false, nil
}

// processRingBell 处理"摇铃"行动，触发结算
func (g *DurianGame) processRingBell(playerID string) (bool, error) {
	if g.phase != "playing" {
		return false, fmt.Errorf("当前阶段 %s 不能摇铃", g.phase)
	}

	// 每轮第一个人不能摇铃（必须先有人接过订单）
	if len(g.playedCardsHistory) == 0 {
		return false, fmt.Errorf("本轮还没有人接过订单，不能摇铃")
	}

	g.phase = "settlement"

	// 执行完整结算流程
	result, err := DoSettlement(
		playerID,
		g.lastOrderPlayerIdx,
		g.players,
		g.holderCards,
		g.orders,
		g.tokenPool,
		g.playerAnger,
	)
	if err != nil {
		return false, fmt.Errorf("结算失败: %w", err)
	}

	g.lastSettlement = result
	g.phase = "round_end"

	// 检查是否有玩家愤怒分 >= 7，触发游戏结束
	for _, pid := range g.players {
		if g.playerAnger[pid].IsEliminated() {
			g.gameOver = true
			g.winner = g.determineWinner()
			g.phase = "game_over"
			return false, nil
		}
	}

	// 标记池耗尽时也触发游戏结束（安全边界）
	if g.tokenPool.IsEmpty() {
		g.gameOver = true
		g.winner = g.determineWinner()
		g.phase = "game_over"
		return false, nil
	}

	// 游戏未结束，进入等待确认状态
	// 从受罚玩家开始（记录待开始的位置）
	punishedIdx := indexOfPlayer(g.players, result.PunishedPlayer)
	g.pendingNextRoundStart = punishedIdx % len(g.players)

	// 初始化等待确认列表（所有玩家都需要确认）
	g.waitingForContinue = make(map[string]bool)
	for _, pid := range g.players {
		g.waitingForContinue[pid] = true
	}

	// 进入等待确认阶段
	g.phase = "waiting_continue"

	// 摇铃后回合结束，但新回合的起始玩家将在 continue 阶段处理
	return true, nil
}

// determineWinner 确定获胜者（愤怒分最少的玩家，并列则共同获胜）
func (g *DurianGame) determineWinner() string {
	minScore := -1
	winners := make([]string, 0)

	for _, pid := range g.players {
		score := g.playerAnger[pid].TotalScore()
		if minScore == -1 || score < minScore {
			minScore = score
			winners = []string{pid}
		} else if score == minScore {
			winners = append(winners, pid)
		}
	}

	if len(winners) == 1 {
		return winners[0]
	}
	return strings.Join(winners, ",")
}

// buildPublicState 构建公共状态 map（不含任何玩家私密牌架信息）
func (g *DurianGame) buildPublicState() map[string]interface{} {
	// 构建玩家愤怒信息列表
	playerInfos := make([]map[string]interface{}, 0, len(g.players))
	for _, pid := range g.players {
		anger := g.playerAnger[pid]
		info := map[string]interface{}{
			"player_id":    pid,
			"anger_score":  anger.TotalScore(),
			"anger_tokens": anger.TokenValues(),
		}
		playerInfos = append(playerInfos, info)
	}

	currentPlayer := ""
	if !g.gameOver && g.phase != "waiting_continue" {
		currentPlayer = g.CurrentTurn()
	}

	state := map[string]interface{}{
		"phase":            g.phase,
		"round_number":     g.roundNumber,
		"current":          currentPlayer,
		"deck_remaining":   g.deck.Remaining(),
		"orders":           g.ordersToMap(),
		"anger_token_pool": g.tokenPool.RemainingValues(),
		"players":          playerInfos,
		"game_over":        g.gameOver,
		"winner":           g.winner,
		// 添加当前抽到的牌（如果有）
		"current_drawn_card": g.getFlippedCardForState(),
		// 添加历史记录
		"played_cards_history": g.getPlayedCardsHistoryForState(),
		// 添加是否等待交换订单状态
		"waiting_for_swap_order": g.waitingForSwapOrder,
	}

	// 附加回合截止时间
	if !g.turnDeadline.IsZero() && g.phase == "playing" && !g.gameOver {
		state["turnDeadline"] = g.turnDeadline.UnixMilli()
	}

	// 附加最近一次结算信息（结算/轮次结束时有意义）
	if g.lastSettlement != nil {
		state["last_settlement"] = settlementToMap(g.lastSettlement, g.players)
	}

	// 附加等待确认继续的玩家列表（按座位顺序）
	if g.phase == "waiting_continue" && g.waitingForContinue != nil {
		waitingList := make([]string, 0)
		for _, pid := range g.players {
			if g.waitingForContinue[pid] {
				waitingList = append(waitingList, pid)
			}
		}
		state["waiting_for_continue"] = waitingList
	}

	return state
}

// ordersToMap 将订单转换为可序列化的 map
func (g *DurianGame) ordersToMap() map[string]int {
	result := make(map[string]int, 4)
	for fruit, count := range g.orders {
		result[string(fruit)] = count
	}
	return result
}

// ─────────────────────────────────────────────
// 辅助函数
// ─────────────────────────────────────────────

// parseChosenSide 从 action.Data 中解析 chosen_side 字段
func parseChosenSide(data interface{}) (string, error) {
	if data == nil {
		return "", fmt.Errorf("缺少 chosen_side 参数，必须为 \"left\" 或 \"right\"")
	}

	// data 经过 JSON 反序列化后通常是 map[string]interface{}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// 也可能直接是字符串（兼容处理）
		if s, ok := data.(string); ok {
			return validateSide(s)
		}
		return "", fmt.Errorf("chosen_side 数据格式错误，期望 map 或 string 类型")
	}

	sideVal, exists := dataMap["chosen_side"]
	if !exists {
		return "", fmt.Errorf("缺少 chosen_side 字段，必须为 \"left\" 或 \"right\"")
	}

	side, ok := sideVal.(string)
	if !ok {
		return "", fmt.Errorf("chosen_side 必须是字符串类型")
	}

	return validateSide(side)
}

// validateSide 验证 chosen_side 值合法性
func validateSide(side string) (string, error) {
	if side != "left" && side != "right" {
		return "", fmt.Errorf("chosen_side 非法值 %q，必须为 \"left\" 或 \"right\"", side)
	}
	return side, nil
}

// indexOfPlayer 在玩家列表中查找玩家索引
func indexOfPlayer(players []string, playerID string) int {
	for i, pid := range players {
		if pid == playerID {
			return i
		}
	}
	return 0
}

// cardToMap 将 Card 转换为可序列化的 map
func cardToMap(card Card) map[string]interface{} {
	if card == nil {
		return nil
	}
	switch c := card.(type) {
	case *FruitCard:
		return map[string]interface{}{
			"id":          c.ID,
			"card_type":   string(c.Type),
			"left_fruit":  string(c.LeftFruit),
			"left_count":  c.LeftCount,
			"right_fruit": string(c.RightFruit),
			"right_count": c.RightCount,
		}
	case *GorillaCard:
		return map[string]interface{}{
			"id":          c.ID,
			"card_type":   string(c.Type),
			"ability":     string(c.Ability),
			"description": c.Description,
		}
	}
	return nil
}

// settlementToMap 将结算结果转换为可序列化的 map
func settlementToMap(r *SettlementResult, playerOrder []string) map[string]interface{} {
	if r == nil {
		return nil
	}

	// 转换缺货水果列表
	shortageFruits := make([]string, len(r.ShortageFruits))
	for i, f := range r.ShortageFruits {
		shortageFruits[i] = string(f)
	}

	// 转换猩猩牌特效列表
	gorillaEffects := make([]map[string]interface{}, len(r.GorillaEffects))
	for i, e := range r.GorillaEffects {
		gorillaEffects[i] = map[string]interface{}{
			"player_id":        e.PlayerID,
			"ability":          string(e.Ability),
			"cancelled_orders": e.CancelledOrders,
		}
	}

	// 转换所有牌架卡（按 g.players 座位顺序，确保顺序稳定）
	allHolderCards := make([]map[string]interface{}, 0, len(r.AllHolderCards))
	for _, pid := range playerOrder {
		if card, exists := r.AllHolderCards[pid]; exists {
			allHolderCards = append(allHolderCards, map[string]interface{}{
				"player_id": pid,
				"card":      cardToMap(card),
			})
		}
	}

	// 转换订单 map
	ordersBefore := make(map[string]int)
	for k, v := range r.OrdersBeforeCancel {
		ordersBefore[string(k)] = v
	}
	ordersAfter := make(map[string]int)
	for k, v := range r.OrdersAfterCancel {
		ordersAfter[string(k)] = v
	}
	inventory := make(map[string]int)
	for k, v := range r.Inventory {
		inventory[string(k)] = v
	}

	return map[string]interface{}{
		"bell_ringer":          r.BellRinger,
		"all_holder_cards":     allHolderCards,
		"gorilla_effects":      gorillaEffects,
		"orders_before_cancel": ordersBefore,
		"orders_after_cancel":  ordersAfter,
		"inventory":            inventory,
		"is_shortage":          r.IsShortage,
		"shortage_fruits":      shortageFruits,
		"punished_player":      r.PunishedPlayer,
		"punish_reason":        r.PunishReason,
		"anger_token_given":    r.AngerTokenGiven,
	}
}

// getFlippedCardForState 返回当前翻开的牌用于状态广播
// 如果没有翻开的牌或不是 playing 阶段，则返回 nil
func (g *DurianGame) getFlippedCardForState() interface{} {
	// 只在 playing 阶段且有翻开的牌时才返回
	if g.phase != "playing" || g.flippedCard == nil {
		return nil
	}

	// 将 Card 转换为可序列化的 map
	return cardToMap(g.flippedCard)
}

// getPlayedCardsHistoryForState 返回历史记录用于状态广播
func (g *DurianGame) getPlayedCardsHistoryForState() []interface{} {
	result := make([]interface{}, len(g.playedCardsHistory))
	for i, record := range g.playedCardsHistory {
		result[i] = map[string]interface{}{
			"player_name":  record.PlayerName,
			"card":         cardToMap(record.Card),
			"chosen_side":  record.ChosenSide,
			"timestamp":    record.Timestamp,
			"round_number": record.RoundNumber,
			"swapped":      record.Swapped,
		}
	}
	return result
}

// ─────────────────────────────────────────────
// TimeoutHandler 接口实现
// ─────────────────────────────────────────────

// SetTimeoutCallbacks 设置超时回调函数（由房间层调用）
func (g *DurianGame) SetTimeoutCallbacks(
	onTurnTimeout func(playerID string),
	onPendingTimeout func(),
) {
	g.onTurnTimeout = onTurnTimeout
	g.onPendingTimeout = onPendingTimeout
	// 回调设置完成后，启动第一个回合的定时器
	g.startTurnTimer()
}

// HandleTurnTimeout 处理回合操作超时
func (g *DurianGame) HandleTurnTimeout(playerID string) error {
	if g.gameOver {
		return nil
	}
	if g.CurrentTurn() != playerID {
		log.Printf("[榴莲忘返] 玩家 %s 超时回调触发，但当前回合已变更，忽略", playerID)
		return nil
	}
	log.Printf("[榴莲忘返] 玩家 %s 回合超时，自动处理", playerID)

	if g.phase == "playing" {
		// 等待交换订单（猩猩牌效果）：自动选第一张未翻转的水果牌
		if g.waitingForSwapOrder {
			for i, record := range g.playedCardsHistory {
				if _, ok := record.Card.(*FruitCard); ok && !record.Swapped {
					action := interfaces.Action{
						Type: "take_order",
						Data: map[string]interface{}{"order_index": i},
					}
					_, err := g.ProcessAction(playerID, action)
					return err
				}
			}
			// 没有可交换的牌，不应该到这里
			return nil
		}

		if g.flippedCard != nil {
			// 已翻牌等待选择左右面：自动选左面
			action := interfaces.Action{
				Type: "take_order",
				Data: map[string]interface{}{"chosen_side": "left"},
			}
			_, err := g.ProcessAction(playerID, action)
			return err
		}

		// 未翻牌：自动接新订单
		action := interfaces.Action{Type: "take_order", Data: map[string]interface{}{}}
		_, err := g.ProcessAction(playerID, action)
		return err
	}

	return nil
}

// HandlePendingTimeout 榴莲忘返没有pending阶段，空实现
func (g *DurianGame) HandlePendingTimeout() error {
	return nil
}

// startTurnTimer 启动回合操作超时定时器（30秒）
func (g *DurianGame) startTurnTimer() {
	g.stopTurnTimer()
	if g.gameOver || g.phase != "playing" {
		return
	}
	g.turnDeadline = time.Now().Add(30 * time.Second)
	g.turnTimer = time.AfterFunc(30*time.Second, func() {
		if g.onTurnTimeout != nil {
			g.onTurnTimeout(g.CurrentTurn())
		}
	})
}

// stopTurnTimer 停止回合操作超时定时器
func (g *DurianGame) stopTurnTimer() {
	if g.turnTimer != nil {
		g.turnTimer.Stop()
		g.turnTimer = nil
	}
	g.turnDeadline = time.Time{}
}

// ─────────────────────────────────────────────
// BotPlayer 接口实现
// ─────────────────────────────────────────────

// NeedsBotCheckAfterAction 每次操作后都检查机器人（因为有continue阶段）
func (g *DurianGame) NeedsBotCheckAfterAction() bool {
	return true
}

// GetBotAction 获取指定机器人的操作
func (g *DurianGame) GetBotAction(botID string) *interfaces.Action {
	// 等待确认继续阶段：机器人自动确认
	if g.phase == "waiting_continue" {
		if g.waitingForContinue != nil && g.waitingForContinue[botID] {
			return &interfaces.Action{Type: "continue"}
		}
		return nil
	}

	// 只在 playing 阶段且轮到该机器人时操作
	if g.phase != "playing" || g.CurrentTurn() != botID {
		return nil
	}

	// 等待交换订单（猩猩牌效果）：随机选一张未翻转的水果牌
	if g.waitingForSwapOrder {
		candidates := []int{}
		for i, record := range g.playedCardsHistory {
			if _, ok := record.Card.(*FruitCard); ok && !record.Swapped {
				candidates = append(candidates, i)
			}
		}
		if len(candidates) > 0 {
			idx := candidates[rand.Intn(len(candidates))]
			return &interfaces.Action{
				Type: "take_order",
				Data: map[string]interface{}{"order_index": idx},
			}
		}
		return nil
	}

	// 已翻牌等待选择：随机选左或右
	if g.flippedCard != nil {
		side := "left"
		if rand.Intn(2) == 0 {
			side = "right"
		}
		return &interfaces.Action{
			Type: "take_order",
			Data: map[string]interface{}{"chosen_side": side},
		}
	}

	// 未翻牌：决定接订单还是摇铃
	// 策略：订单区有水果且已有一些历史记录时，有概率摇铃
	totalOrders := 0
	for _, count := range g.orders {
		totalOrders += count
	}

	// 至少有2条历史记录且订单区有水果时，30%概率摇铃
	if len(g.playedCardsHistory) >= 4 && totalOrders > 0 && rand.Intn(10) < 3 {
		return &interfaces.Action{Type: "ring_bell"}
	}

	// 默认接新订单
	return &interfaces.Action{Type: "take_order", Data: map[string]interface{}{}}
}
