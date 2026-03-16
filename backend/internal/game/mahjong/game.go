package mahjong

import (
	"fmt"
	"log"
	"sort"
	"time"

	"card-game-server/backend/internal/game/interfaces"
)

// 游戏阶段
const (
	PhasePlay    = "play"    // 出牌阶段：当前玩家需要出牌
	PhasePending = "pending" // 等待响应：其他玩家决定是否吃碰杠胡
)

// 动作类型
const (
	ActionDiscard = "discard" // 出牌
	ActionChow    = "chow"    // 吃
	ActionPong    = "pong"    // 碰
	ActionKong    = "kong"    // 杠
	ActionWin     = "win"     // 胡牌
	ActionPass    = "pass"    // 过
)

// 动作优先级：胡 > 杠 > 碰 > 吃
var actionPriority = map[string]int{
	ActionWin:  4,
	ActionKong: 3,
	ActionPong: 2,
	ActionChow: 1,
	ActionPass: 0,
}

// PendingResponse 玩家的响应
type PendingResponse struct {
	Action string
	Data   interface{}
}

// WinResult 胡牌结果
type WinResult struct {
	WinnerSeat int        `json:"winner_seat"`
	WinType    string     `json:"win_type"` // "self_drawn" or "discard"
	LoserSeat  int        `json:"loser_seat"`
	FanList    []*FanItem `json:"fan_list"`
	TotalFan   int        `json:"total_fan"`
}

// MahjongGame 麻将游戏核心逻辑
type MahjongGame struct {
	players     []string       // 4个玩家ID，按座位[东南西北]
	playerSeats map[string]int // playerID -> 座位号

	hands    [4]Tiles  // 各座位手牌
	melds    [4][]Meld // 各座位副露
	discards [4]Tiles  // 各座位牌河

	wall Tiles // 牌墙

	dealerSeat  int // 庄家座位（0-3）
	currentSeat int // 当前回合座位

	phase string // 游戏阶段

	// pending 状态
	lastDiscard      *Tile                    // 最后打出的牌
	lastDiscardSeat  int                      // 最后出牌的座位
	pendingActions   map[int][]string         // 座位 -> 可执行的操作列表
	pendingResponses map[int]*PendingResponse // 座位 -> 已选择的响应
	pendingChowData  map[int]interface{}      // 吃牌时的额外数据

	gameOver  bool
	winner    string
	winResult *WinResult

	// 游戏内定时器（游戏层自己管理）
	turnTimer    *time.Timer // 回合操作超时定时器
	pendingTimer *time.Timer // 等待响应超时定时器

	// 截止时间（用于前端倒计时显示）
	turnDeadline    time.Time // 回合操作截止时间
	pendingDeadline time.Time // 等待响应截止时间

	// 超时回调（通知房间层执行自动操作）
	onTurnTimeout    func(playerID string) // 回合操作超时回调
	onPendingTimeout func()                // 等待响应超时回调
}

func New() interfaces.Game {
	return &MahjongGame{}
}

func (g *MahjongGame) ID() string       { return "mahjong" }
func (g *MahjongGame) MaxPlayers() int  { return 4 }
func (g *MahjongGame) MinPlayers() int  { return 4 }
func (g *MahjongGame) IsGameOver() bool { return g.gameOver }
func (g *MahjongGame) Winner() string   { return g.winner }

func (g *MahjongGame) Init(players []string) error {
	if len(players) != 4 {
		return fmt.Errorf("麻将必须4名玩家")
	}

	g.players = players
	g.playerSeats = make(map[string]int)
	for i, pid := range players {
		g.playerSeats[pid] = i
	}

	g.melds = [4][]Meld{}
	g.discards = [4]Tiles{}
	g.pendingActions = make(map[int][]string)
	g.pendingResponses = make(map[int]*PendingResponse)
	g.pendingChowData = make(map[int]interface{})
	g.dealerSeat = 0
	g.gameOver = false
	g.winner = ""
	g.winResult = nil

	// 洗牌发牌
	g.wall = NewDeck()
	ShuffleDeck(g.wall)

	// 庄家14张，闲家各13张
	for seat := 0; seat < 4; seat++ {
		count := 13
		if seat == g.dealerSeat {
			count = 14
		}
		g.hands[seat] = make(Tiles, count)
		copy(g.hands[seat], g.wall[:count])
		g.wall = g.wall[count:]
		SortTiles(g.hands[seat])
	}

	// 庄家先出牌
	g.currentSeat = g.dealerSeat
	g.phase = PhasePlay

	// 注意：定时器在 SetTimeoutCallbacks 设置回调后启动
	log.Printf("[麻将] 游戏初始化完成，庄家座位:%d，牌墙剩余:%d", g.dealerSeat, len(g.wall))
	return nil
}

func (g *MahjongGame) CurrentTurn() string {
	return g.players[g.currentSeat]
}

// SetTimeoutCallbacks 设置超时回调函数（由房间层调用）
func (g *MahjongGame) SetTimeoutCallbacks(
	onTurnTimeout func(playerID string),
	onPendingTimeout func(),
) {
	g.onTurnTimeout = onTurnTimeout
	g.onPendingTimeout = onPendingTimeout
	// 回调设置完成后，启动第一个回合的定时器
	g.startTurnTimer()
}

// HandleTurnTimeout 处理回合操作超时（游戏层自己决定如何处理）
func (g *MahjongGame) HandleTurnTimeout(playerID string) error {
	// 游戏层自己执行自动出牌逻辑
	// ProcessAction 会在 turnEnded=true 时自动调用 AdvanceTurn()
	action := interfaces.Action{Type: ActionDiscard, Data: nil}
	_, err := g.ProcessAction(playerID, action)
	return err
}

// HandlePendingTimeout 处理等待响应超时（游戏层模拟所有人都 pass）
func (g *MahjongGame) HandlePendingTimeout() error {
	// 等待响应超时，相当于所有人都 pass
	// 清理 pending 状态
	g.pendingActions = make(map[int][]string)
	g.pendingResponses = make(map[int]*PendingResponse)
	g.pendingChowData = make(map[int]interface{})

	// 停止当前的等待响应定时器
	g.stopPendingTimer()

	// 直接推进到下一玩家并摸牌（模拟所有人都pass的情况）
	g.advanceToNextPlayer()
	log.Printf("[麻将] 等待响应超时，视为全部 pass，推进到座位%d", g.currentSeat)
	return nil
}

// startTurnTimer 启动回合操作超时定时器（游戏层自己管理）
func (g *MahjongGame) startTurnTimer() {
	g.stopTurnTimer()
	g.turnDeadline = time.Now().Add(30 * time.Second)

	g.turnTimer = time.AfterFunc(30*time.Second, func() {
		// 回合操作超时，通知房间层
		if g.onTurnTimeout != nil {
			g.onTurnTimeout(g.players[g.currentSeat])
		}
	})
}

// stopTurnTimer 停止回合操作超时定时器
func (g *MahjongGame) stopTurnTimer() {
	if g.turnTimer != nil {
		g.turnTimer.Stop()
		g.turnTimer = nil
	}
}

// startPendingTimer 启动等待响应超时定时器（游戏层自己管理）
func (g *MahjongGame) startPendingTimer() {
	g.stopPendingTimer()

	g.pendingDeadline = time.Now().Add(10 * time.Second)
	log.Printf("[麻将] 启动等待响应定时器（10秒）")
	g.pendingTimer = time.AfterFunc(10*time.Second, func() {
		log.Printf("[麻将] 等待响应定时器触发")
		// 等待响应超时，通知房间层
		if g.onPendingTimeout != nil {
			g.onPendingTimeout()
		} else {
			log.Printf("[麻将] 警告：onPendingTimeout 回调为 nil")
		}
	})
}

// stopPendingTimer 停止等待响应超时定时器
func (g *MahjongGame) stopPendingTimer() {
	if g.pendingTimer != nil {
		g.pendingTimer.Stop()
		g.pendingTimer = nil
	}
}

func (g *MahjongGame) AdvanceTurn() {
	// 推进到下一玩家并摸牌（正常回合结束或超时都会调用这里）
	// 注意：advanceToNextPlayer 内部会启动新的定时器，所以这里不需要额外处理
	g.advanceToNextPlayer()
}

// ProcessAction 处理玩家动作
func (g *MahjongGame) ProcessAction(playerID string, action interface{}) (bool, error) {
	act, ok := action.(interfaces.Action)
	if !ok {
		return false, fmt.Errorf("无效的action类型")
	}

	seat, ok := g.playerSeats[playerID]
	if !ok {
		return false, fmt.Errorf("玩家不在游戏中")
	}

	// 验证操作是否有效（是当前回合玩家或pending阶段有权限的玩家）
	if g.phase == PhasePlay && seat != g.currentSeat {
		return false, fmt.Errorf("不是你的回合")
	}
	if g.phase == PhasePending {
		if _, hasAction := g.pendingActions[seat]; !hasAction {
			return false, fmt.Errorf("你没有可执行的操作")
		}
	}

	// 操作有效，停止当前定时器
	g.stopTurnTimer()
	g.stopPendingTimer()

	var turnEnded bool
	var err error

	switch g.phase {
	case PhasePlay:
		turnEnded, err = g.handlePlayAction(seat, act)
	case PhasePending:
		turnEnded, err = g.handlePendingAction(seat, act)
	default:
		// 未知阶段，启动出牌定时器作为 fallback
		g.startTurnTimer()
		return false, fmt.Errorf("未知游戏阶段: %s", g.phase)
	}

	if err != nil {
		// 操作执行失败，根据当前阶段重启相应的定时器
		if g.phase == PhasePending {
			g.startPendingTimer()
		} else {
			g.startTurnTimer()
		}
		return false, err
	}

	// 如果回合已结束（需要推进），游戏层自己推进
	// 注意：AdvanceTurn 内部会处理定时器启动
	if turnEnded {
		g.AdvanceTurn()
		return true, nil
	}

	// 回合未结束，根据游戏状态启动相应的定时器
	// 如果处于等待响应阶段，启动响应定时器；否则启动出牌定时器
	if g.phase == PhasePending {
		g.startPendingTimer()
	} else {
		g.startTurnTimer()
	}

	return false, nil
}

// === 出牌阶段 ===

func (g *MahjongGame) handlePlayAction(seat int, act interfaces.Action) (bool, error) {
	if seat != g.currentSeat {
		return false, fmt.Errorf("不是你的回合")
	}

	switch act.Type {
	case ActionDiscard:
		return g.handleDiscard(seat, act.Data)
	case ActionWin:
		return g.handleSelfDrawWin(seat)
	case ActionKong:
		return g.handleConcealedOrExtendedKong(seat, act.Data)
	default:
		return false, fmt.Errorf("出牌阶段不支持操作: %s", act.Type)
	}
}

func (g *MahjongGame) handleDiscard(seat int, data interface{}) (bool, error) {
	var tile Tile
	var ok bool

	if data == nil {
		// 自动出牌：打出第一张牌（最后一张是刚摸的，不打）
		if len(g.hands[seat]) == 0 {
			return false, fmt.Errorf("没有手牌可出")
		}
		// 打出第一张非刚摸的牌，如果没有则打出第一张
		tile = g.hands[seat][0]
		if len(g.hands[seat]) > 1 {
			// 找到不是最后一张的牌（最后一张是刚摸的）
			tile = g.hands[seat][len(g.hands[seat])-2]
		}
	} else {
		tile, ok = ParseTile(data)
		if !ok {
			return false, fmt.Errorf("无效的牌数据")
		}
	}

	if !g.hands[seat].Contains(tile) {
		return false, fmt.Errorf("手牌中没有这张牌: %s", tile)
	}

	// 出牌
	g.hands[seat] = g.hands[seat].Remove(tile)
	SortTiles(g.hands[seat])
	g.discards[seat] = append(g.discards[seat], tile)
	g.lastDiscard = &tile
	g.lastDiscardSeat = seat

	log.Printf("[麻将] 座位%d出牌: %s, 剩余手牌:%d", seat, tile, len(g.hands[seat]))

	// 检查其他玩家是否可以响应
	g.checkPendingActions()

	if len(g.pendingActions) > 0 {
		g.phase = PhasePending
		g.pendingResponses = make(map[int]*PendingResponse)
		log.Printf("[麻将] 进入等待响应阶段，有%d个玩家可响应", len(g.pendingActions))
		return false, nil
	} else {
		return true, nil
	}
}

func (g *MahjongGame) handleSelfDrawWin(seat int) (bool, error) {
	if !CanWin(g.hands[seat], g.melds[seat]) {
		return false, fmt.Errorf("当前手牌不满足胡牌条件")
	}

	// 计算番数
	ctx := &WinContext{
		Hand:       g.hands[seat].Copy(),
		Melds:      g.melds[seat],
		WinTile:    g.hands[seat][len(g.hands[seat])-1], // 最后摸的牌
		SelfDrawn:  true,
		DealerSeat: g.dealerSeat,
		WinnerSeat: seat,
		SeatWind:   seat + 1,
		RoundWind:  1, // 简化处理，默认东风圈
	}
	fans, total := CalculateFan(ctx)

	// if total < 8 {
	// 	return false, fmt.Errorf("番数不足8番（当前%d番），不能胡牌", total)
	// }

	g.winResult = &WinResult{
		WinnerSeat: seat,
		WinType:    "self_drawn",
		LoserSeat:  -1,
		FanList:    fans,
		TotalFan:   total,
	}
	g.gameOver = true
	g.winner = fmt.Sprintf("座位%d自摸胡牌（%d番）", seat, total)
	log.Printf("[麻将] %s", g.winner)
	return true, nil
}

func (g *MahjongGame) handleConcealedOrExtendedKong(seat int, data interface{}) (bool, error) {
	tile, ok := ParseTile(data)
	if !ok {
		return false, fmt.Errorf("无效的牌数据")
	}

	// 暗杠
	if g.hands[seat].Count(tile) == 4 {
		g.hands[seat] = g.hands[seat].RemoveN(tile, 4)
		g.melds[seat] = append(g.melds[seat], Meld{
			Type:       MeldKongConcealed,
			Tiles:      Tiles{tile, tile, tile, tile},
			FromPlayer: -1,
		})
		// 摸一张杠牌
		if err := g.drawTile(seat); err != nil {
			return false, err
		}
		log.Printf("[麻将] 座位%d暗杠: %s", seat, tile)
		return false, nil
	}

	// 补杠
	for i, m := range g.melds[seat] {
		if m.Type == MeldPong && m.Tiles[0].Equal(tile) && g.hands[seat].Contains(tile) {
			g.hands[seat] = g.hands[seat].Remove(tile)
			g.melds[seat][i].Type = MeldKongExtended
			g.melds[seat][i].Tiles = append(g.melds[seat][i].Tiles, tile)
			// 摸一张杠牌
			if err := g.drawTile(seat); err != nil {
				return false, err
			}
			log.Printf("[麻将] 座位%d补杠: %s", seat, tile)
			return false, nil
		}
	}

	return false, fmt.Errorf("无法执行杠操作: %s", tile)
}

// === 等待响应阶段 ===

func (g *MahjongGame) handlePendingAction(seat int, act interfaces.Action) (bool, error) {
	actions, hasActions := g.pendingActions[seat]
	if !hasActions {
		return false, fmt.Errorf("你没有可执行的操作")
	}

	// 验证操作合法性
	validAction := false
	if act.Type == ActionPass {
		validAction = true
	} else {
		for _, a := range actions {
			if a == act.Type {
				validAction = true
				break
			}
		}
	}
	if !validAction {
		return false, fmt.Errorf("不合法的操作: %s", act.Type)
	}

	// 记录响应
	g.pendingResponses[seat] = &PendingResponse{Action: act.Type, Data: act.Data}
	if act.Type == ActionChow {
		g.pendingChowData[seat] = act.Data
	}

	log.Printf("[麻将] 座位%d响应: %s", seat, act.Type)

	// 检查是否所有人都已响应
	if g.allPendingResolved() {
		return g.resolvePending()
	}

	return false, nil
}

func (g *MahjongGame) allPendingResolved() bool {
	for seat := range g.pendingActions {
		if _, responded := g.pendingResponses[seat]; !responded {
			return false
		}
	}
	return true
}

func (g *MahjongGame) resolvePending() (bool, error) {
	// 找出最高优先级的响应
	bestSeat := -1
	bestPriority := -1
	bestAction := ""

	for seat, resp := range g.pendingResponses {
		p := actionPriority[resp.Action]
		if p > bestPriority {
			bestPriority = p
			bestSeat = seat
			bestAction = resp.Action
		} else if p == bestPriority && p > 0 {
			// 同优先级按座次（距出牌者最近的优先）
			if g.seatDistance(seat) < g.seatDistance(bestSeat) {
				bestSeat = seat
				bestAction = resp.Action
			}
		}
	}

	// 清理 pending 状态
	g.pendingActions = make(map[int][]string)
	defer func() {
		g.pendingResponses = make(map[int]*PendingResponse)
		g.pendingChowData = make(map[int]interface{})
	}()

	// 所有人都 pass
	if bestAction == ActionPass || bestSeat == -1 {
		// 停止等待响应定时器
		g.stopPendingTimer()
		g.advanceToNextPlayer()
		return false, nil
	}

	tile := *g.lastDiscard

	switch bestAction {
	case ActionWin:
		return g.resolveDiscardWin(bestSeat, tile)
	case ActionKong:
		return g.resolveExposedKong(bestSeat, tile)
	case ActionPong:
		return g.resolvePong(bestSeat, tile)
	case ActionChow:
		return g.resolveChow(bestSeat, tile)
	}

	return false, fmt.Errorf("未知的响应操作: %s", bestAction)
}

func (g *MahjongGame) resolveDiscardWin(seat int, tile Tile) (bool, error) {
	// 将胡的牌加入手牌进行判断
	testHand := append(g.hands[seat].Copy(), tile)

	ctx := &WinContext{
		Hand:       testHand,
		Melds:      g.melds[seat],
		WinTile:    tile,
		SelfDrawn:  false,
		DealerSeat: g.dealerSeat,
		WinnerSeat: seat,
		SeatWind:   seat + 1,
		RoundWind:  1,
	}
	fans, total := CalculateFan(ctx)

	if total < 8 {
		// 番数不足，当作pass处理
		g.advanceToNextPlayer()
		return false, nil
	}

	g.winResult = &WinResult{
		WinnerSeat: seat,
		WinType:    "discard",
		LoserSeat:  g.lastDiscardSeat,
		FanList:    fans,
		TotalFan:   total,
	}
	g.gameOver = true
	g.winner = fmt.Sprintf("座位%d点炮胡牌（%d番），点炮者座位%d", seat, total, g.lastDiscardSeat)
	log.Printf("[麻将] %s", g.winner)

	// 将牌从牌河移除（被吃掉了）
	discarder := g.lastDiscardSeat
	if len(g.discards[discarder]) > 0 {
		g.discards[discarder] = g.discards[discarder][:len(g.discards[discarder])-1]
	}
	g.hands[seat] = append(g.hands[seat], tile)
	SortTiles(g.hands[seat])

	return true, nil
}

func (g *MahjongGame) resolveExposedKong(seat int, tile Tile) (bool, error) {
	g.hands[seat] = g.hands[seat].RemoveN(tile, 3)
	g.melds[seat] = append(g.melds[seat], Meld{
		Type:       MeldKongExposed,
		Tiles:      Tiles{tile, tile, tile, tile},
		FromPlayer: g.lastDiscardSeat,
	})

	// 从牌河移除
	discarder := g.lastDiscardSeat
	if len(g.discards[discarder]) > 0 {
		g.discards[discarder] = g.discards[discarder][:len(g.discards[discarder])-1]
	}

	// 摸杠牌
	if err := g.drawTile(seat); err != nil {
		return false, err
	}
	g.currentSeat = seat
	g.phase = PhasePlay
	log.Printf("[麻将] 座位%d明杠: %s", seat, tile)
	return false, nil
}

func (g *MahjongGame) resolvePong(seat int, tile Tile) (bool, error) {
	g.hands[seat] = g.hands[seat].RemoveN(tile, 2)
	g.melds[seat] = append(g.melds[seat], Meld{
		Type:       MeldPong,
		Tiles:      Tiles{tile, tile, tile},
		FromPlayer: g.lastDiscardSeat,
	})

	// 从牌河移除
	discarder := g.lastDiscardSeat
	if len(g.discards[discarder]) > 0 {
		g.discards[discarder] = g.discards[discarder][:len(g.discards[discarder])-1]
	}

	// 碰牌后不摸牌，直接出牌
	g.currentSeat = seat
	g.phase = PhasePlay
	log.Printf("[麻将] 座位%d碰: %s", seat, tile)
	return false, nil
}

func (g *MahjongGame) resolveChow(seat int, tile Tile) (bool, error) {
	// 解析吃牌数据：客户端传来的2张手牌
	resp := g.pendingChowData[seat]
	chowTiles, err := parseChowTiles(resp)
	if err != nil {
		return false, fmt.Errorf("吃牌数据错误: %v", err)
	}

	// 验证吃牌合法性
	allThree := append(Tiles{tile}, chowTiles...)
	if !isValidChowSet(allThree) {
		return false, fmt.Errorf("不合法的吃牌组合")
	}

	// 从手牌移除
	for _, t := range chowTiles {
		if !g.hands[seat].Contains(t) {
			return false, fmt.Errorf("手牌中没有: %s", t)
		}
		g.hands[seat] = g.hands[seat].Remove(t)
	}

	SortTiles(allThree)
	g.melds[seat] = append(g.melds[seat], Meld{
		Type:       MeldChow,
		Tiles:      allThree,
		FromPlayer: g.lastDiscardSeat,
	})

	// 从牌河移除
	discarder := g.lastDiscardSeat
	if len(g.discards[discarder]) > 0 {
		g.discards[discarder] = g.discards[discarder][:len(g.discards[discarder])-1]
	}

	// 吃牌后不摸牌，直接出牌
	g.currentSeat = seat
	g.phase = PhasePlay
	log.Printf("[麻将] 座位%d吃: %s", seat, tile)
	return false, nil
}

// === 辅助方法 ===

// checkPendingActions 检查出牌后谁可以响应
func (g *MahjongGame) checkPendingActions() {
	g.pendingActions = make(map[int][]string)
	if g.lastDiscard == nil {
		return
	}
	tile := *g.lastDiscard

	for seat := 0; seat < 4; seat++ {
		if seat == g.lastDiscardSeat {
			continue
		}

		var actions []string

		// 胡牌检查
		testHand := append(g.hands[seat].Copy(), tile)
		if CanWin(testHand, g.melds[seat]) {
			actions = append(actions, ActionWin)
		}

		// 杠检查
		if CanKongExposed(g.hands[seat], tile) {
			actions = append(actions, ActionKong)
		}

		// 碰检查
		if CanPong(g.hands[seat], tile) {
			actions = append(actions, ActionPong)
		}

		// 吃检查（只有下家可以吃）
		nextSeat := (g.lastDiscardSeat + 1) % 4
		if seat == nextSeat {
			chowOptions := CanChow(g.hands[seat], tile)
			if len(chowOptions) > 0 {
				actions = append(actions, ActionChow)
			}
		}

		if len(actions) > 0 {
			g.pendingActions[seat] = actions
		}
	}
}

// advanceToNextPlayer 推进到下一个玩家，自动摸牌并启动定时器
func (g *MahjongGame) advanceToNextPlayer() {
	g.currentSeat = (g.lastDiscardSeat + 1) % 4
	g.lastDiscard = nil
	g.phase = PhasePlay

	// 检查牌墙是否还有牌
	if len(g.wall) == 0 {
		g.handleDrawGame()
		return
	}

	// 自动摸牌
	if err := g.drawTile(g.currentSeat); err != nil {
		log.Printf("[麻将] 摸牌失败：%v", err)
		g.handleDrawGame()
		return
	}

	// 启动新回合的出牌定时器
	g.startTurnTimer()
}

// drawTile 摸牌
func (g *MahjongGame) drawTile(seat int) error {
	if len(g.wall) == 0 {
		g.handleDrawGame()
		return fmt.Errorf("牌墙已空")
	}

	tile := g.wall[0]
	g.wall = g.wall[1:]
	g.hands[seat] = append(g.hands[seat], tile)
	// 不排序，最后一张是刚摸的牌（方便标记）

	log.Printf("[麻将] 座位%d摸牌: %s, 牌墙剩余:%d", seat, tile, len(g.wall))
	return nil
}

// handleDrawGame 流局
func (g *MahjongGame) handleDrawGame() {
	g.gameOver = true
	g.winner = "流局"
	log.Printf("[麻将] 流局")
}

// seatDistance 计算座位到出牌者的距离（逆时针）
func (g *MahjongGame) seatDistance(seat int) int {
	return (seat - g.lastDiscardSeat + 4) % 4
}

// getAvailableActions 获取指定座位当前可执行的操作
func (g *MahjongGame) getAvailableActions(seat int) []string {
	if g.phase == PhasePending {
		if actions, ok := g.pendingActions[seat]; ok {
			// 如果还没响应，返回可用操作
			if _, responded := g.pendingResponses[seat]; !responded {
				return append(actions, ActionPass)
			}
		}
		return nil
	}

	if g.phase == PhasePlay && seat == g.currentSeat {
		actions := []string{ActionDiscard}

		// 自摸胡牌检查
		if CanWin(g.hands[seat], g.melds[seat]) {
			actions = append(actions, ActionWin)
		}
		// 暗杠检查
		kongTiles := CanKongConcealed(g.hands[seat])
		if len(kongTiles) > 0 {
			actions = append(actions, ActionKong)
		}
		// 补杠检查
		extKongTiles := CanKongExtended(g.hands[seat], g.melds[seat])
		if len(extKongTiles) > 0 {
			if !containsStr(actions, ActionKong) {
				actions = append(actions, ActionKong)
			}
		}
		return actions
	}

	return nil
}

// === 状态查询 ===

func (g *MahjongGame) GetState() interface{} {
	handCounts := make([]int, 4)
	for i := 0; i < 4; i++ {
		handCounts[i] = len(g.hands[i])
	}

	playersInfo := make([]map[string]interface{}, 4)
	for i := 0; i < 4; i++ {
		playersInfo[i] = map[string]interface{}{
			"player_id":  g.players[i],
			"seat":       i,
			"hand_count": handCounts[i],
			"melds":      MeldsToMaps(g.melds[i]),
			"discards":   g.discards[i].ToMaps(),
			"is_dealer":  i == g.dealerSeat,
		}
	}

	state := map[string]interface{}{
		"phase":            g.phase,
		"current_seat":     g.currentSeat,
		"current":          g.players[g.currentSeat],
		"dealer_seat":      g.dealerSeat,
		"wall_remaining":   len(g.wall),
		"players":          playersInfo,
		"game_over":        g.gameOver,
		"winner":           g.winner,
		"turn_deadline":    g.turnDeadline.UnixMilli(),
		"pending_deadline": g.pendingDeadline.UnixMilli(),
	}

	if g.lastDiscard != nil {
		state["last_discard"] = g.lastDiscard.ToMap()
		state["last_discard_seat"] = g.lastDiscardSeat
	}

	if g.winResult != nil {
		state["win_result"] = g.winResult
	}

	return state
}

func (g *MahjongGame) GetStateForPlayer(playerID string) interface{} {
	seat, ok := g.playerSeats[playerID]
	if !ok {
		return g.GetState()
	}

	baseState := g.GetState().(map[string]interface{})

	// 添加手牌
	baseState["my_seat"] = seat
	baseState["my_hand"] = g.hands[seat].ToMaps()

	// 添加可执行操作
	actions := g.getAvailableActions(seat)
	if len(actions) > 0 {
		baseState["available_actions"] = actions
	}

	// 吃牌时提供可选组合
	if g.phase == PhasePending {
		if acts, ok := g.pendingActions[seat]; ok {
			for _, a := range acts {
				if a == ActionChow && g.lastDiscard != nil {
					chowOptions := CanChow(g.hands[seat], *g.lastDiscard)
					if len(chowOptions) > 0 {
						opts := make([][]map[string]interface{}, len(chowOptions))
						for i, opt := range chowOptions {
							opts[i] = opt.ToMaps()
						}
						baseState["chow_options"] = opts
					}
				}
			}
		}
	}

	// 杠牌时提供可杠选项
	if g.phase == PhasePlay && seat == g.currentSeat {
		kongConcealed := CanKongConcealed(g.hands[seat])
		kongExtended := CanKongExtended(g.hands[seat], g.melds[seat])
		var kongOptions []map[string]interface{}
		for _, t := range kongConcealed {
			kongOptions = append(kongOptions, map[string]interface{}{
				"tile": t.ToMap(),
				"type": "concealed",
			})
		}
		for _, t := range kongExtended {
			kongOptions = append(kongOptions, map[string]interface{}{
				"tile": t.ToMap(),
				"type": "extended",
			})
		}
		if len(kongOptions) > 0 {
			baseState["kong_options"] = kongOptions
		}
	}

	// 别人打出的牌可以明杠时，提供杠选项
	if g.phase == PhasePending && g.lastDiscard != nil {
		if acts, ok := g.pendingActions[seat]; ok {
			for _, a := range acts {
				if a == ActionKong {
					baseState["kong_options"] = []map[string]interface{}{
						{
							"tile": g.lastDiscard.ToMap(),
							"type": "exposed",
						},
					}
					break
				}
			}
		}
	}

	return baseState
}

// === 工具函数 ===

func parseChowTiles(data interface{}) (Tiles, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("吃牌数据必须是数组")
	}
	if len(arr) != 2 {
		return nil, fmt.Errorf("吃牌需要提供2张手牌")
	}
	tiles := make(Tiles, 2)
	for i, item := range arr {
		t, ok := ParseTile(item)
		if !ok {
			return nil, fmt.Errorf("第%d张牌格式错误", i+1)
		}
		tiles[i] = t
	}
	return tiles, nil
}

func isValidChowSet(tiles Tiles) bool {
	if len(tiles) != 3 {
		return false
	}
	// 必须同花色、不能是字牌
	if tiles[0].IsHonor() {
		return false
	}
	for i := 1; i < 3; i++ {
		if tiles[i].Suit != tiles[0].Suit {
			return false
		}
	}
	// 排序后检查是否连续
	sorted := tiles.Copy()
	sort.Sort(sorted)
	return sorted[1].Rank == sorted[0].Rank+1 && sorted[2].Rank == sorted[1].Rank+1
}

func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// === BotPlayer 接口实现 ===

// NeedsBotCheckAfterAction 麻将需要在每次操作后检查所有机器人
// 因为pending阶段任何玩家都可能需要响应（吃碰杠胡）
func (g *MahjongGame) NeedsBotCheckAfterAction() bool {
	return true
}

// GetBotAction 获取指定机器人的操作
// 返回 nil 表示该机器人当前无需操作
func (g *MahjongGame) GetBotAction(botID string) *interfaces.Action {
	seat, ok := g.playerSeats[botID]
	if !ok {
		return nil
	}

	actions := g.getAvailableActions(seat)
	if len(actions) == 0 {
		return nil
	}

	switch g.phase {
	case PhasePlay:
		// 出牌阶段：打出最后一张牌（刚摸的）
		if seat != g.currentSeat {
			return nil
		}
		hand := g.hands[seat]
		if len(hand) == 0 {
			return &interfaces.Action{Type: ActionDiscard, Data: nil}
		}
		lastTile := hand[len(hand)-1]
		return &interfaces.Action{
			Type: ActionDiscard,
			Data: lastTile.ToMap(),
		}

	case PhasePending:
		// 响应阶段：简单AI，总是pass
		return &interfaces.Action{Type: ActionPass}

	default:
		return nil
	}
}
