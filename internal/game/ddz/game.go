package ddz

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"card-game-server/internal/game/interfaces"
)

type Card struct {
	Value int    // 3~15 (3~A), 16=小王, 17=大王
	Suit  string // "1"~ "4" 或 "小王"/"大王"
}

type Hand []Card

type DdzGame struct {
	players     []string
	hands       map[string]Hand
	bottomCards Hand
	landlord    string
	calls       map[string]int // 叫分记录
	callTurn    int
	outTurn     int
	lastPlay    Play
	phase       string // "call" 或 "play"
	gameOver    bool
	winner      string
}

type Play struct {
	Type  string
	Cards Hand
	Value int // 用于比较的主牌值（例如炸弹/火箭用最小牌值或特殊标记）
}

func New() interfaces.Game {
	return &DdzGame{}
}

func (g *DdzGame) ID() string { return "ddz" }

func (g *DdzGame) Init(players []string) error {
	if len(players) != 3 {
		return fmt.Errorf("斗地主必须正好3名玩家")
	}
	g.players = players
	g.hands = make(map[string]Hand)
	g.calls = make(map[string]int)
	g.phase = "call"
	g.callTurn = 0
	g.outTurn = 0
	g.lastPlay = Play{}
	rand.Seed(time.Now().UnixNano())

	g.dealCards()
	return nil
}

func (g *DdzGame) dealCards() {
	var deck Hand
	for suit := 1; suit <= 4; suit++ {
		for v := 3; v <= 15; v++ {
			deck = append(deck, Card{Value: v, Suit: strconv.Itoa(suit)})
		}
	}
	deck = append(deck, Card{Value: 16, Suit: "小王"}, Card{Value: 17, Suit: "大王"})

	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	for i, p := range g.players {
		g.hands[p] = deck[i*17 : (i+1)*17]
		sort.Sort(g.hands[p])
	}
	g.bottomCards = deck[51:54]
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
		g.callTurn = (g.callTurn + 1) % 3

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
			return true, nil // 阶段切换
		}
		return false, nil // 继续叫地主

	case "play":
		if act.Type == "pass" {
			g.outTurn = (g.outTurn + 1) % 3
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

		playType, valid := parsePlayType(played)
		if !valid {
			return false, fmt.Errorf("无效的牌型")
		}

		if !g.beatsLast(playType, played) {
			return false, fmt.Errorf("无法压制上家")
		}

		// 移除手牌
		g.hands[playerID] = removeCards(g.hands[playerID], played)

		g.lastPlay = Play{Type: playType, Cards: played}
		turnEnd := len(g.hands[playerID]) == 0

		if turnEnd {
			g.gameOver = true
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

		g.outTurn = (g.outTurn + 1) % 3
		return false, nil
	}

	return false, fmt.Errorf("未知游戏阶段")
}

func (g *DdzGame) AdvanceTurn() {
	// 本游戏中 AdvanceTurn 由 ProcessAction 内部控制，不额外调用
}

func (g *DdzGame) GetState() interface{} {
	handsPublic := make(map[string][]int)
	for p, hand := range g.hands {
		vals := make([]int, len(hand))
		for i, c := range hand {
			vals[i] = c.Value
		}
		handsPublic[p+"_hand"] = vals
	}

	return map[string]interface{}{
		"phase":     g.phase,
		"players":   g.players,
		"landlord":  g.landlord,
		"current":   g.CurrentTurn(),
		"last_play": g.lastPlay,
		"game_over": g.gameOver,
		"winner":    g.winner,
		"hands":     handsPublic, // 只暴露 value 列表
		"bottom":    g.bottomCardsValues(),
	}
}

func (g *DdzGame) bottomCardsValues() []int {
	vals := make([]int, len(g.bottomCards))
	for i, c := range g.bottomCards {
		vals[i] = c.Value
	}
	return vals
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

	// 王炸
	if len(cards) == 2 && cards[0].Value == 16 && cards[1].Value == 17 {
		return "rocket", true
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
		// 四带二 等可后续扩展
	}

	// 顺子、连对、飞机等暂未实现，可后续补充
	return "", false
}

func (g *DdzGame) beatsLast(playType string, cards Hand) bool {
	if g.lastPlay.Type == "" {
		return true // 第一手任意合法牌型
	}

	if playType == "rocket" {
		return true
	}
	if playType == "bomb" && g.lastPlay.Type != "rocket" {
		return true
	}

	// 同类型比牌（简化，只比长度和首张大小）
	if playType == g.lastPlay.Type && len(cards) == len(g.lastPlay.Cards) {
		return cards[0].Value > g.lastPlay.Cards[0].Value
	}

	return false
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
