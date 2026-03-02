package durian

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"card-game-server/internal/game/interfaces"
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
}

// PlayedCardRecord 已打出的牌记录
type PlayedCardRecord struct {
	PlayerName  string      `json:"player_name"`
	Card        Card        `json:"card"`
	ChosenSide  string      `json:"chosen_side"`
	Timestamp   string      `json:"timestamp"`
	RoundNumber int         `json:"round_number"`
}

// 确保实现接口
var _ interfaces.Game = (*DurianGame)(nil)

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
}

// ProcessAction 处理玩家行动
// action 实际类型为 interfaces.Action
func (g *DurianGame) ProcessAction(playerID string, action interface{}) (bool, error) {
	if g.gameOver {
		return false, fmt.Errorf("游戏已结束")
	}

	if g.CurrentTurn() != playerID {
		return false, fmt.Errorf("不是你的回合，当前轮到 %s", g.CurrentTurn())
	}

	act, ok := action.(interfaces.Action)
	if !ok {
		return false, fmt.Errorf("无效的 action 类型，期望 interfaces.Action")
	}

	switch act.Type {
	case "take_order":
		return g.processTakeOrder(playerID, act)
	case "ring_bell":
		return g.processRingBell(playerID)
	default:
		return false, fmt.Errorf("未知的行动类型: %s，支持 take_order 或 ring_bell", act.Type)
	}
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

	// 将上一轮的牌架卡放入废牌堆
	for _, card := range g.holderCards {
		if card != nil {
			g.deck.Discard(card)
		}
	}

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
	return nil
}

// processTakeOrder 处理"接新订单"行动
// 自动翻开一张牌，根据 chosen_side 选择水果加入订单区
func (g *DurianGame) processTakeOrder(playerID string, act interfaces.Action) (bool, error) {
	if g.phase != "playing" {
		return false, fmt.Errorf("当前阶段 %s 不能接订单", g.phase)
	}

	// 解析 chosen_side
	chosenSide, err := parseChosenSide(act.Data)
	if err != nil {
		return false, err
	}

	// 从牌堆翻一张牌
	card, ok := g.deck.Draw()
	if !ok {
		return false, fmt.Errorf("牌堆已空，无法翻牌")
	}
	g.flippedCard = card

	// 处理翻出的牌
	switch c := card.(type) {
	case *FruitCard:
		// 根据选择的面添加订单
		switch chosenSide {
		case "left":
			g.orders[c.LeftFruit] += c.LeftCount
		case "right":
			g.orders[c.RightFruit] += c.RightCount
		}
		// 翻开的牌废弃（不放入牌架）
		g.deck.Discard(card)

	case *GorillaCard:
		// 翻出猩猩牌：没有水果面可选，强制结算（参考PRD 9.2节）
		// 将此牌废弃并触发强制摇铃
		g.deck.Discard(card)
		// 强制结算，摇铃者视为当前玩家
		return g.processRingBell(playerID)
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
		
	// 返回 true，由 room.go 调用 AdvanceTurn() 切换到下一位玩家
	return true, nil
}

// processRingBell 处理"摇铃"行动，触发结算
func (g *DurianGame) processRingBell(playerID string) (bool, error) {
	if g.phase != "playing" {
		return false, fmt.Errorf("当前阶段 %s 不能摇铃", g.phase)
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

	// 游戏未结束，开始新一轮
	// 从受罚玩家的下一位开始
	punishedIdx := indexOfPlayer(g.players, result.PunishedPlayer)
	nextFirstIdx := (punishedIdx + 1) % len(g.players)

	if err := g.startNewRound(nextFirstIdx); err != nil {
		return false, fmt.Errorf("开始新一轮失败: %w", err)
	}

	// ring_bell 后内部已完成轮次处理，返回 false 告知 room.go 不要再调用 AdvanceTurn
	return false, nil
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
	if !g.gameOver {
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
	}

	// 附加最近一次结算信息（结算/轮次结束时有意义）
	if g.lastSettlement != nil {
		state["last_settlement"] = settlementToMap(g.lastSettlement)
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
func settlementToMap(r *SettlementResult) map[string]interface{} {
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

	// 转换所有牌架卡
	allHolderCards := make([]map[string]interface{}, 0, len(r.AllHolderCards))
	for pid, card := range r.AllHolderCards {
		allHolderCards = append(allHolderCards, map[string]interface{}{
			"player_id": pid,
			"card":      cardToMap(card),
		})
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
		}
	}
	return result
}
