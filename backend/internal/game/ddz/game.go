package ddz

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"card-game-server/backend/internal/game/interfaces"
)

type Card struct {
	Value int    // 3~15 (3~A), 16=小王, 17=大王
	Suit  string // "1"~ "4" 或 "小王"/"大王"
}

type Hand []Card

type DdzGame struct {
	players     []string       // 按座位号排序的玩家列表 [座位0, 座位1, 座位2]
	playerSeats map[string]int // playerID -> 座位号
	hands       map[string]Hand
	bottomCards Hand
	landlord    string
	calls       map[string]int // 叫分记录
	callTurn    int            // 当前叫地主回合的座位号
	outTurn     int            // 当前出牌回合的座位号
	lastPlay    Play
	lastPlayer  string // 记录最后出牌的玩家
	passCount   int    // 记录连续pass的次数
	phase       string // "call" 或 "play"
	gameOver    bool
	winner      string

	// 每个玩家最近一次出牌/pass记录（用于前端显示）
	lastActions map[string]*PlayerAction

	// 超时相关
	turnTimer        *time.Timer           // 回合操作超时定时器
	turnDeadline     time.Time             // 当前回合截止时间
	onTurnTimeout    func(playerID string) // 回合操作超时回调
	onPendingTimeout func()                // 未使用，但需实现接口
}

// PlayerAction 玩家最近一次操作
type PlayerAction struct {
	Type  string `json:"type"`  // "play" | "pass" | "call"
	Cards Hand   `json:"cards"` // 出的牌（pass时为空）
	Score int    `json:"score"` // 叫分（仅call时有效）
}

type Play struct {
	Type  string
	Cards Hand
	Value int // 用于比较的主牌值（例如炸弹/火箭用最小牌值或特殊标记）
}

func New() interfaces.Game {
	return &DdzGame{}
}

func (g *DdzGame) MaxPlayers() int { return 3 }
func (g *DdzGame) MinPlayers() int { return 3 }

func (g *DdzGame) ID() string { return "ddz" }

func (g *DdzGame) Init(players []string) error {
	if len(players) != 3 {
		return fmt.Errorf("斗地主必须正好3名玩家")
	}
	g.players = players
	g.playerSeats = make(map[string]int)
	for seatNum, playerID := range players {
		if playerID != "" {
			g.playerSeats[playerID] = seatNum
		}
	}
	g.hands = make(map[string]Hand)
	g.calls = make(map[string]int)
	g.phase = "call"
	g.callTurn = 0
	g.outTurn = 0
	g.lastPlay = Play{}
	g.lastPlayer = ""
	g.passCount = 0
	g.landlord = ""
	g.gameOver = false
	g.winner = ""
	g.lastActions = make(map[string]*PlayerAction)
	rand.Seed(time.Now().UnixNano())

	g.dealCards()

	// 注意：定时器在 setupGameTimers 设置回调后由房间层启动
	// 这里先记录截止时间，实际定时器在 SetTimeoutCallbacks 后启动
	return nil
}

func (g *DdzGame) dealCards() {
	// 创建标准54张牌
	deck := make(Hand, 0, 54)
	for suit := 1; suit <= 4; suit++ {
		for v := 3; v <= 15; v++ {
			deck = append(deck, Card{Value: v, Suit: strconv.Itoa(suit)})
		}
	}
	deck = append(deck, Card{Value: 16, Suit: "小王"}, Card{Value: 17, Suit: "大王"})

	// 验证牌数
	if len(deck) != 54 {
		panic(fmt.Sprintf("牌数错误：期望54张，实际%d张", len(deck)))
	}

	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	// 发牌：每人17张，底牌3张（创建独立副本避免引用问题）
	for i, p := range g.players {
		start := i * 17
		end := (i + 1) * 17
		if end > len(deck) {
			panic("发牌越界")
		}
		// 创建手牌的独立副本
		hand := make(Hand, 17)
		copy(hand, deck[start:end])
		g.hands[p] = hand
		sort.Sort(g.hands[p])
	}

	// 底牌也创建独立副本
	g.bottomCards = make(Hand, 3)
	copy(g.bottomCards, deck[51:54])

	// 验证每个玩家手牌数量
	for player, hand := range g.hands {
		if len(hand) != 17 {
			panic(fmt.Sprintf("玩家%s手牌数量错误：期望17张，实际%d张", player, len(hand)))
		}
	}
}

func (g *DdzGame) CurrentTurn() string {
	if g.phase == "call" {
		return g.players[g.callTurn]
	}
	return g.players[g.outTurn]
}

func (g *DdzGame) ProcessAction(playerID string, action interface{}) (bool, error) {
	if g.CurrentTurn() != playerID {
		return false, fmt.Errorf("不是你的回合")
	}

	act, ok := action.(interfaces.Action)
	if !ok {
		return false, fmt.Errorf("无效的 action 类型")
	}

	switch g.phase {
	case "call":
		if act.Type != "call_landlord" {
			return false, fmt.Errorf("叫地主阶段只能 call_landlord")
		}
		scoreF, ok := act.Data.(float64)
		if !ok {
			return false, fmt.Errorf("叫分必须是数字")
		}
		score := int(scoreF)
		if score < 0 || score > 3 {
			return false, fmt.Errorf("叫分范围 0~3")
		}

		g.calls[playerID] = score
		g.lastActions[playerID] = &PlayerAction{Type: "call", Score: score}
		g.callTurn = (g.callTurn + 2) % 3 // 逆时针

		// 判断是否结束叫地主
		if score == 3 || g.callTurn == 0 {
			maxScore := 0
			landlord := ""
			for p, s := range g.calls {
				if s > maxScore {
					maxScore = s
					landlord = p
				}
			}
			if maxScore == 0 {
				// 都0分，流局（这里简化处理，重开或随机）
				return false, fmt.Errorf("无人叫地主，游戏流局")
			}
			g.landlord = landlord
			g.hands[landlord] = append(g.hands[landlord], g.bottomCards...)
			sort.Sort(g.hands[landlord])
			g.phase = "play"
			g.outTurn = indexOf(g.players, landlord)
			g.startTurnTimer() // 地主开始出牌，重启定时器
			return true, nil   // 阶段切换
		}
		g.startTurnTimer() // 下一个人叫地主
		return false, nil  // 继续叫地主

	case "play":
		if act.Type == "pass" {
			// 新回合第一手牌必须出，不能pass
			if g.lastPlay.Type == "" {
				return false, fmt.Errorf("新回合第一手牌必须出牌，不能pass")
			}

			g.lastActions[playerID] = &PlayerAction{Type: "pass"}
			g.passCount++
			g.outTurn = (g.outTurn + 2) % 3 // 逆时针

			// 如果所有人都pass了（连续2次pass），则最后出牌的人重新获得出牌权
			if g.passCount >= 2 && g.lastPlayer != "" {
				g.outTurn = indexOf(g.players, g.lastPlayer)
				g.passCount = 0
				g.lastPlay = Play{} // 清空上一手牌
				// 新回合清空所有玩家的最近操作
				g.lastActions = make(map[string]*PlayerAction)
				g.startTurnTimer() // 新回合，重启定时器
				return true, nil   // 回合结束，重新开始
			}
			g.startTurnTimer() // 下一个人出牌
			return false, nil
		}

		if act.Type != "play_cards" {
			return false, fmt.Errorf("出牌阶段只能 play_cards 或 pass")
		}

		rawCards, ok := act.Data.([]interface{})
		if !ok {
			return false, fmt.Errorf("出牌数据格式错误")
		}

		played := make(Hand, len(rawCards))
		for i, item := range rawCards {
			m, ok := item.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("卡牌格式错误")
			}
			vf, ok := m["value"].(float64)
			if !ok {
				return false, fmt.Errorf("卡牌 value 必须是数字")
			}
			played[i] = Card{Value: int(vf)}
		}

		// 排序手牌（确保王炸等牌型判断正确）
		sort.Sort(played)

		// 添加调试信息
		fmt.Printf("玩家 %s 尝试出牌: ", playerID)
		for _, card := range played {
			fmt.Printf("%s ", valueToStr(card.Value))
		}
		fmt.Printf("\n")

		playType, valid := parsePlayType(played)
		if !valid {
			return false, fmt.Errorf("无效的牌型: %v", played)
		}

		fmt.Printf("识别的牌型: %s\n", playType)

		// 如果是新回合的第一手牌，不需要压制上家
		if g.lastPlay.Type != "" {
			if !g.beatsLast(playType, played) {
				return false, fmt.Errorf("无法压制上家")
			}
		}

		// 移除手牌
		g.hands[playerID] = removeCards(g.hands[playerID], played)

		// 调试：显示剩余手牌数
		fmt.Printf("玩家 %s 剩余手牌数: %d\n", playerID, len(g.hands[playerID]))

		// 更新游戏状态
		g.lastPlay = Play{Type: playType, Cards: played}
		g.lastPlayer = playerID
		g.lastActions[playerID] = &PlayerAction{Type: "play", Cards: played}
		g.passCount = 0                 // 有人出牌，重置pass计数
		g.outTurn = (g.outTurn + 2) % 3 // 逆时针

		remainingCards := len(g.hands[playerID])
		turnEnd := remainingCards == 0
		fmt.Printf("remainingCards=%d, turnEnd=%v, gameOver=%v\n", remainingCards, turnEnd, g.gameOver)

		if turnEnd {
			fmt.Printf("设置 gameOver=true\n")
			g.gameOver = true
			g.stopTurnTimer() // 游戏结束，停止定时器
			if playerID == g.landlord {
				g.winner = "地主胜"
			} else {
				// 检查农民是否都出完
				farmerWin := true
				for _, p := range g.players {
					if p != g.landlord && len(g.hands[p]) > 0 {
						farmerWin = false
						break
					}
				}
				if farmerWin {
					g.winner = "农民胜"
				} else {
					g.winner = "进行中" // 理论上不应发生
				}
			}
			return true, nil
		}
		g.startTurnTimer() // 下一个人出牌
		return false, nil

		// // 测试模式：第一张牌直接获胜
		// fmt.Printf("测试模式：玩家 %s 出第一张牌直接获胜\n", playerID)
		// g.gameOver = true
		// if playerID == g.landlord {
		// 	g.winner = "地主胜"
		// } else {
		// 	g.winner = "农民胜"
		// }
		// return true, nil
	}

	return false, fmt.Errorf("未知游戏阶段")
}

func (g *DdzGame) AdvanceTurn() {
	// 本游戏中 AdvanceTurn 由 ProcessAction 内部控制，不额外调用
}

// ── TimeoutHandler 接口实现 ──

// SetTimeoutCallbacks 设置超时回调函数（由房间层调用）
func (g *DdzGame) SetTimeoutCallbacks(
	onTurnTimeout func(playerID string),
	onPendingTimeout func(),
) {
	g.onTurnTimeout = onTurnTimeout
	g.onPendingTimeout = onPendingTimeout
	// 回调设置完成后，启动第一个回合的定时器
	g.startTurnTimer()
}

// HandleTurnTimeout 处理回合操作超时
func (g *DdzGame) HandleTurnTimeout(playerID string) error {
	// 如果游戏已结束，忽略超时
	if g.gameOver {
		return nil
	}
	// 如果当前回合已经不是该玩家，说明玩家已经操作过了，忽略超时
	if g.CurrentTurn() != playerID {
		log.Printf("[斗地主] 玩家 %s 超时回调触发，但当前回合已变更，忽略", playerID)
		return nil
	}
	log.Printf("[斗地主] 玩家 %s 回合超时，自动处理", playerID)
	if g.phase == "call" {
		// 叫地主超时：自动叫0分（不叫）
		action := interfaces.Action{Type: "call_landlord", Data: float64(0)}
		_, err := g.ProcessAction(playerID, action)
		return err
	}
	// 出牌超时
	if g.lastPlay.Type == "" {
		// 新回合必须出牌，不能pass：自动出最小的一张牌
		hand := g.hands[playerID]
		if len(hand) > 0 {
			smallest := hand[0]
			action := interfaces.Action{
				Type: "play_cards",
				Data: []interface{}{map[string]interface{}{"value": float64(smallest.Value)}},
			}
			_, err := g.ProcessAction(playerID, action)
			return err
		}
	}
	// 有上家出牌：自动pass
	action := interfaces.Action{Type: "pass"}
	_, err := g.ProcessAction(playerID, action)
	return err
}

// HandlePendingTimeout DDZ没有pending阶段，空实现
func (g *DdzGame) HandlePendingTimeout() error {
	return nil
}

// startTurnTimer 启动回合操作超时定时器
func (g *DdzGame) startTurnTimer() {
	g.stopTurnTimer()
	g.turnDeadline = time.Now().Add(30 * time.Second)
	g.turnTimer = time.AfterFunc(30*time.Second, func() {
		if g.onTurnTimeout != nil {
			g.onTurnTimeout(g.CurrentTurn())
		}
	})
}

// stopTurnTimer 停止回合操作超时定时器
func (g *DdzGame) stopTurnTimer() {
	if g.turnTimer != nil {
		g.turnTimer.Stop()
		g.turnTimer = nil
	}
}

// GetState 返回游戏状态（不包含具体手牌，只返回手牌数量）
func (g *DdzGame) GetState() interface{} {
	// 只返回手牌数量，不暴露具体牌面
	handCounts := make(map[string]int)
	for p, hand := range g.hands {
		handCounts[p] = len(hand)
	}

	// 构建按座位号排序的玩家信息
	playerInfos := make([]map[string]interface{}, len(g.players))
	for seatNum, playerID := range g.players {
		if playerID != "" {
			playerInfos[seatNum] = map[string]interface{}{
				"player_id":   playerID,
				"seat_number": seatNum,
				"is_landlord": playerID == g.landlord,
				"hand_count":  handCounts[playerID], // 手牌数量
			}
		}
	}

	return map[string]interface{}{
		"phase":         g.phase,
		"players":       g.players,
		"player_infos":  playerInfos,
		"player_seats":  g.playerSeats,
		"landlord":      g.landlord,
		"current":       g.CurrentTurn(),
		"current_seat":  g.getCurrentSeat(),
		"last_play":     g.lastPlay,
		"last_actions":  g.lastActions,
		"game_over":     g.gameOver,
		"winner":        g.winner,
		"hand_counts":   handCounts, // 各玩家手牌数量
		"bottom":        g.bottomCardsForState(),
		"turn_deadline": g.turnDeadline.UnixMilli(),
	}
}

// GetStateForPlayer 返回指定玩家的游戏状态（包含该玩家的手牌）
func (g *DdzGame) GetStateForPlayer(playerID string) interface{} {
	// 获取基础状态
	baseState := g.GetState().(map[string]interface{})

	// 添加该玩家的手牌
	if hand, exists := g.hands[playerID]; exists {
		vals := make([]int, len(hand))
		for i, c := range hand {
			vals[i] = c.Value
		}
		baseState["my_hand"] = vals
	}

	return baseState
}

// 获取当前回合玩家的座位号
func (g *DdzGame) getCurrentSeat() int {
	if g.phase == "call" {
		return g.callTurn
	}
	return g.outTurn
}

func (g *DdzGame) bottomCardsValues() []int {
	vals := make([]int, len(g.bottomCards))
	for i, c := range g.bottomCards {
		vals[i] = c.Value
	}
	return vals
}

// bottomCardsForState 只在地主确定后才返回底牌信息
func (g *DdzGame) bottomCardsForState() []int {
	if g.landlord == "" {
		// 叫地主阶段，不暴露底牌
		return nil
	}
	return g.bottomCardsValues()
}

func (g *DdzGame) IsGameOver() bool {
	return g.gameOver
}

func (g *DdzGame) Winner() string {
	return g.winner
}

// ────────────────────────────────────────────────
// 辅助函数
// ────────────────────────────────────────────────

func parsePlayType(cards Hand) (string, bool) {
	if len(cards) == 0 {
		return "", false
	}

	countMap := make(map[int]int)
	for _, c := range cards {
		countMap[c.Value]++
	}

	// 调试日志
	fmt.Printf("解析牌型 - 牌数:%d, countMap:%v\n", len(cards), countMap)

	// 王炸（大王17 + 小王16，不依赖顺序）
	if len(cards) == 2 {
		v1, v2 := cards[0].Value, cards[1].Value
		if (v1 == 16 && v2 == 17) || (v1 == 17 && v2 == 16) {
			return "rocket", true
		}
	}

	switch len(cards) {
	case 1:
		return "single", true
	case 2:
		if cards[0].Value == cards[1].Value {
			return "pair", true
		}
	case 3:
		if len(countMap) == 1 {
			return "triple", true
		}
	case 4:
		if len(countMap) == 1 {
			return "bomb", true
		}
		// 检查是否为三带一
		if isTripleWithSingle(countMap) {
			return "triple_single", true
		}
	case 5:
		// 5张牌可能是顺子或三带对子
		if isStraight(cards) {
			return "straight", true
		}
		// 检查是否为三带对子
		if isTripleWithPair(countMap) {
			fmt.Printf("识别为三带对子牌型\n")
			return "triple_pair", true
		}
	case 6:
		// 6张顺子或四带二
		if isStraight(cards) {
			return "straight", true
		}
		if isQuadWithPair(countMap) {
			return "quad_pair", true
		}
	}

	// 顺子、连对、飞机等判断
	if isStraight(cards) {
		return "straight", true
	}

	// 连对判断（双顺）
	if isPairSequence(cards) {
		return "pair_sequence", true
	}

	// 飞机判断（三顺）
	if isTripleSequence(cards) {
		return "triple_sequence", true
	}

	// 飞机带翅膀
	if isTripleSequenceWithWings(cards) {
		return "triple_sequence_wings", true
	}

	fmt.Printf("无法识别的牌型\n")
	return "", false
}

// 判断是否为三带一对子
func isTripleWithPair(countMap map[int]int) bool {
	if len(countMap) != 2 {
		return false
	}

	tripleFound := false
	pairFound := false

	for _, count := range countMap {
		if count == 3 {
			tripleFound = true
		} else if count == 2 {
			pairFound = true
		}
	}

	return tripleFound && pairFound
}

// 判断是否为三带一
func isTripleWithSingle(countMap map[int]int) bool {
	if len(countMap) != 2 {
		return false
	}

	tripleFound := false
	singleFound := false

	for _, count := range countMap {
		if count == 3 {
			tripleFound = true
		} else if count == 1 {
			singleFound = true
		}
	}

	return tripleFound && singleFound
}

// 判断是否为顺子（连续单牌）
func isStraight(cards Hand) bool {
	if len(cards) < 5 {
		return false // 顺子至少5张
	}

	// 检查是否有重复的牌
	countMap := make(map[int]int)
	for _, c := range cards {
		// 2不能参与顺子
		if c.Value == 15 {
			return false
		}
		countMap[c.Value]++
		if countMap[c.Value] > 1 {
			return false
		}
	}

	// 检查是否连续
	values := make([]int, 0, len(cards))
	for value := range countMap {
		values = append(values, value)
	}
	sort.Ints(values)

	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}

	return true
}

// 判断是否为连对（双顺）
func isPairSequence(cards Hand) bool {
	if len(cards)%2 != 0 || len(cards) < 6 {
		return false // 连对必须是偶数张，至少6张（3对）
	}

	// 统计每张牌的数量
	countMap := make(map[int]int)
	for _, c := range cards {
		// 2不能参与连对
		if c.Value == 15 {
			return false
		}
		countMap[c.Value]++
	}

	// 每张牌必须出现2次
	pairs := 0
	for _, count := range countMap {
		if count != 2 {
			return false
		}
		pairs++
	}

	// 检查对子的值是否连续
	values := make([]int, 0, pairs)
	for value := range countMap {
		values = append(values, value)
	}
	sort.Ints(values)

	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}

	return true
}

// 判断是否为飞机（三顺）
func isTripleSequence(cards Hand) bool {
	if len(cards)%3 != 0 || len(cards) < 6 {
		return false // 飞机必须是3的倍数，至少6张（2个三张）
	}

	// 统计每张牌的数量
	countMap := make(map[int]int)
	for _, c := range cards {
		// 2不能参与飞机
		if c.Value == 15 {
			return false
		}
		countMap[c.Value]++
	}

	// 每张牌必须出现3次
	triples := 0
	for _, count := range countMap {
		if count != 3 {
			return false
		}
		triples++
	}

	// 检查三张的值是否连续
	values := make([]int, 0, triples)
	for value := range countMap {
		values = append(values, value)
	}
	sort.Ints(values)

	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}

	return true
}

// 判断是否为四带二
func isQuadWithPair(countMap map[int]int) bool {
	if len(countMap) != 2 && len(countMap) != 3 {
		return false
	}

	quadFound := false
	pairCount := 0

	for _, count := range countMap {
		if count == 4 {
			quadFound = true
		} else if count == 2 {
			pairCount++
		} else if count != 1 {
			return false // 只能有四张和一对，或者四张和两个单张
		}
	}

	// 四带二对 或 四带两单
	return quadFound && (pairCount == 1 || pairCount == 0)
}

// 判断是否为飞机带翅膀
func isTripleSequenceWithWings(cards Hand) bool {
	countMap := make(map[int]int)
	for _, c := range cards {
		countMap[c.Value]++
	}

	// 找出三张序列
	triples := make([]int, 0)
	singles := 0
	pairs := 0

	for value, count := range countMap {
		if count == 3 {
			triples = append(triples, value)
		} else if count == 2 {
			pairs++
		} else if count == 1 {
			singles++
		}
	}

	if len(triples) < 2 {
		return false // 至少需要两个三张组成序列
	}

	// 检查三张是否连续
	sort.Ints(triples)
	for i := 1; i < len(triples); i++ {
		if triples[i] != triples[i-1]+1 {
			return false
		}
	}

	// 检查翅膀是否匹配
	expectedWings := len(triples)
	return singles+pairs*2 == expectedWings
}

func (g *DdzGame) beatsLast(playType string, cards Hand) bool {
	if g.lastPlay.Type == "" {
		return true // 第一手任意合法牌型
	}

	// 王炸无敌
	if playType == "rocket" {
		return true
	}

	// 炸弹可以压非王炸
	if playType == "bomb" && g.lastPlay.Type != "rocket" {
		return true
	}

	// 同类型牌型比较
	if playType == g.lastPlay.Type {
		// 特殊牌型比较
		switch playType {
		case "triple_single", "triple_pair", "quad_pair", "triple_sequence_wings":
			// 取主要牌型进行比较（三张、四张、飞机主体）
			mainValue1 := getMainCardValue(cards, playType)
			mainValue2 := getMainCardValue(g.lastPlay.Cards, playType)
			return mainValue1 > mainValue2
		case "straight", "pair_sequence", "triple_sequence":
			// 长度必须相同才能比较
			if len(cards) != len(g.lastPlay.Cards) {
				return false
			}
			// 比较最小牌的大小
			return getMinCardValue(cards) > getMinCardValue(g.lastPlay.Cards)
		default:
			// 普通牌型（单牌、对子、三张）按长度和首张大小比较
			if len(cards) == len(g.lastPlay.Cards) {
				return cards[0].Value > g.lastPlay.Cards[0].Value
			}
		}
	}

	return false
}

// 获取主要牌值（用于复合牌型比较）
func getMainCardValue(cards Hand, playType string) int {
	countMap := make(map[int]int)
	for _, c := range cards {
		countMap[c.Value]++
	}

	switch playType {
	case "triple_single", "triple_pair":
		// 找到三张的值
		for value, count := range countMap {
			if count == 3 {
				return value
			}
		}
	case "quad_pair":
		// 找到四张的值
		for value, count := range countMap {
			if count == 4 {
				return value
			}
		}
	case "triple_sequence_wings":
		// 找到飞机主体的最小值
		triples := make([]int, 0)
		for value, count := range countMap {
			if count == 3 {
				triples = append(triples, value)
			}
		}
		sort.Ints(triples)
		if len(triples) > 0 {
			return triples[0]
		}
	}

	return 0
}

func getMinCardValue(cards Hand) int {
	if len(cards) == 0 {
		return 0
	}
	return cards[0].Value
}

func removeCards(hand Hand, played Hand) Hand {
	needRemove := make(map[int]int)
	for _, c := range played {
		needRemove[c.Value]++
	}

	var result Hand
	for _, c := range hand {
		if needRemove[c.Value] > 0 {
			needRemove[c.Value]--
		} else {
			result = append(result, c)
		}
	}

	sort.Sort(result)
	return result
}

// 排序：大王 > 小王 > A > K > ... > 3
func (h Hand) Len() int      { return len(h) }
func (h Hand) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h Hand) Less(i, j int) bool {
	return cardValueForSort(h[i].Value) < cardValueForSort(h[j].Value)
}

func cardValueForSort(v int) int {
	if v == 17 {
		return 100 // 大王
	}
	if v == 16 {
		return 99 // 小王
	}
	if v == 15 {
		return 98 // 2
	}
	return v
}

func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}

// 辅助函数：将数值转换为牌面显示
func valueToStr(value int) string {
	switch value {
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	case 14:
		return "A"
	case 15:
		return "2"
	case 16:
		return "小王"
	case 17:
		return "大王"
	default:
		return fmt.Sprintf("%d", value)
	}
}

// === BotPlayer 接口实现 ===

// NeedsBotCheckAfterAction 斗地主只需在轮到机器人时检查
func (g *DdzGame) NeedsBotCheckAfterAction() bool {
	return false
}

// GetBotAction 获取指定机器人的操作
func (g *DdzGame) GetBotAction(botID string) *interfaces.Action {
	state := g.GetStateForPlayer(botID)
	stateMap, ok := state.(map[string]interface{})
	if !ok {
		return nil
	}

	phase, _ := stateMap["phase"].(string)

	switch phase {
	case "call":
		return g.botCallAction()
	case "play":
		return g.botPlayAction(stateMap)
	default:
		return nil
	}
}

// botCallAction 机器人叫地主策略
func (g *DdzGame) botCallAction() *interfaces.Action {
	score := rand.Intn(3) // 0, 1, 2
	return &interfaces.Action{
		Type: "call_landlord",
		Data: float64(score),
	}
}

// botPlayAction 机器人出牌策略
func (g *DdzGame) botPlayAction(stateMap map[string]interface{}) *interfaces.Action {
	myHandRaw, ok := stateMap["my_hand"].([]int)
	if !ok || len(myHandRaw) == 0 {
		return &interfaces.Action{Type: "pass"}
	}
	myHand := myHandRaw

	// 解析上一手牌
	lastPlayRaw := stateMap["last_play"]
	lastType := ""
	var lastCardValues []int

	if lastPlayRaw != nil {
		// 同进程调用，直接类型断言
		if lp, ok := lastPlayRaw.(Play); ok {
			lastType = lp.Type
			for _, c := range lp.Cards {
				lastCardValues = append(lastCardValues, c.Value)
			}
		}
	}

	// 新回合，出最小的单牌
	if lastType == "" {
		return &interfaces.Action{
			Type: "play_cards",
			Data: []interface{}{map[string]interface{}{"value": float64(myHand[0])}},
		}
	}

	lastMaxValue := 0
	if len(lastCardValues) > 0 {
		sort.Ints(lastCardValues)
		lastMaxValue = lastCardValues[0]
	}

	// 尝试压制
	switch lastType {
	case "single":
		for _, v := range myHand {
			if v > lastMaxValue {
				return &interfaces.Action{
					Type: "play_cards",
					Data: []interface{}{map[string]interface{}{"value": float64(v)}},
				}
			}
		}
	case "pair":
		for _, pv := range findGroups(myHand, 2) {
			if pv > lastMaxValue {
				return &interfaces.Action{
					Type: "play_cards",
					Data: []interface{}{
						map[string]interface{}{"value": float64(pv)},
						map[string]interface{}{"value": float64(pv)},
					},
				}
			}
		}
	case "triple":
		for _, tv := range findGroups(myHand, 3) {
			if tv > lastMaxValue {
				cards := make([]interface{}, 3)
				for i := range cards {
					cards[i] = map[string]interface{}{"value": float64(tv)}
				}
				return &interfaces.Action{Type: "play_cards", Data: cards}
			}
		}
	case "bomb":
		for _, qv := range findGroups(myHand, 4) {
			if qv > lastMaxValue {
				cards := make([]interface{}, 4)
				for i := range cards {
					cards[i] = map[string]interface{}{"value": float64(qv)}
				}
				return &interfaces.Action{Type: "play_cards", Data: cards}
			}
		}
	default:
		// 复杂牌型，尝试用炸弹
		quads := findGroups(myHand, 4)
		if len(quads) > 0 {
			cards := make([]interface{}, 4)
			for i := range cards {
				cards[i] = map[string]interface{}{"value": float64(quads[0])}
			}
			return &interfaces.Action{Type: "play_cards", Data: cards}
		}
	}

	return &interfaces.Action{Type: "pass"}
}

// findGroups 在手牌中找到所有 count 张相同的牌值（已排序）
func findGroups(hand []int, count int) []int {
	countMap := make(map[int]int)
	for _, v := range hand {
		countMap[v]++
	}
	var result []int
	for v, c := range countMap {
		if c >= count {
			result = append(result, v)
		}
	}
	sort.Ints(result)
	return result
}
