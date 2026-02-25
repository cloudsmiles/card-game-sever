package simple

import (
	"card-game-server/internal/game/interfaces"
	"fmt"
)

type SimpleGame struct {
	players   []string
	turnIndex int
	lastCard  int
	hands     map[string][]int
	gameOver  bool
	winner    string
}

func New() interfaces.Game {
	return &SimpleGame{}
}

func (g *SimpleGame) MaxPlayers() int { return 2 }
func (g *SimpleGame) MinPlayers() int { return 2 }

func (g *SimpleGame) ID() string { return "simple" }

func (g *SimpleGame) Init(players []string) error {
	g.players = players
	g.turnIndex = 0
	g.lastCard = 0
	g.hands = make(map[string][]int)
	for i, p := range players {
		g.hands[p] = []int{1 + i*3, 2 + i*3, 3 + i*3}
	}
	return nil
}

func (g *SimpleGame) ProcessAction(playerID string, action interface{}) (bool, error) {
	if g.CurrentTurn() != playerID {
		return false, fmt.Errorf("不是你的回合")
	}
	act, ok := action.(interfaces.Action)
	if !ok || act.Type != "play_card" {
		return false, fmt.Errorf("无效的出牌指令")
	}
	card, ok := act.Data.(float64)
	if !ok {
		return false, fmt.Errorf("卡牌必须是数字")
	}
	c := int(card)

	// 【临时放宽】允许出任意牌（测试用，后面改回 c > g.lastCard）
	// if c <= g.lastCard {
	//     return false, fmt.Errorf("必须出比上家大的牌 %d > %d", c, g.lastCard)
	// }

	g.lastCard = c
	g.hands[playerID] = removeCard(g.hands[playerID], c)

	if len(g.hands[playerID]) == 0 {
		g.gameOver = true
		g.winner = playerID
		return true, nil
	}
	return true, nil
}

func (g *SimpleGame) CurrentTurn() string {
	return g.players[g.turnIndex]
}

func (g *SimpleGame) AdvanceTurn() {
	g.turnIndex = (g.turnIndex + 1) % len(g.players)
}

func (g *SimpleGame) GetState() interface{} {
	return map[string]interface{}{
		"players":   g.players,
		"current":   g.CurrentTurn(),
		"last_card": g.lastCard,
		"hands":     g.hands,
		"game_over": g.gameOver,
		"winner":    g.winner,
	}
}

func (g *SimpleGame) IsGameOver() bool { return g.gameOver }
func (g *SimpleGame) Winner() string   { return g.winner }

func removeCard(hand []int, c int) []int {
	for i, v := range hand {
		if v == c {
			return append(hand[:i], hand[i+1:]...)
		}
	}
	return hand
}
