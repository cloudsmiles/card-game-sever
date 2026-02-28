package mahjong

import (
	"testing"

	"card-game-server/internal/game/interfaces"
)

// ==================== tile_test ====================

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 136 {
		t.Errorf("牌数错误：期望136，实际%d", len(deck))
	}
	// 统计每种牌应该恰好4张
	counts := deck.ToCountArray()
	for i, c := range counts {
		tile := TileFromIndex(i)
		// 字牌只有7种(rank 1-7)，索引27-33有效，但ToIndex用suit*9+(rank-1)
		// 字牌 suit=3, rank 1-7 → index 27-33，index 34以上无效
		if tile.Suit == SuitZi && tile.Rank > 7 {
			continue // 无效索引
		}
		if tile.Suit == SuitZi && tile.Rank < 1 {
			continue
		}
		maxRank := 9
		if tile.Suit == SuitZi {
			maxRank = 7
		}
		if tile.Rank > maxRank {
			continue
		}
		if c != 4 {
			t.Errorf("牌 %s 数量错误：期望4，实际%d", tile, c)
		}
	}
}

func TestTileParsing(t *testing.T) {
	tests := []struct {
		data interface{}
		ok   bool
		suit Suit
		rank int
	}{
		{map[string]interface{}{"suit": "wan", "rank": float64(1)}, true, SuitWan, 1},
		{map[string]interface{}{"suit": "tiao", "rank": float64(9)}, true, SuitTiao, 9},
		{map[string]interface{}{"suit": "tong", "rank": float64(5)}, true, SuitTong, 5},
		{map[string]interface{}{"suit": "zi", "rank": float64(7)}, true, SuitZi, 7},
		{map[string]interface{}{"suit": "zi", "rank": float64(8)}, false, 0, 0},       // 字牌rank不能>7
		{map[string]interface{}{"suit": "invalid", "rank": float64(1)}, false, 0, 0},   // 无效花色
		{map[string]interface{}{"suit": "wan", "rank": float64(0)}, false, 0, 0},        // rank不能<1
	}

	for _, tt := range tests {
		tile, ok := ParseTile(tt.data)
		if ok != tt.ok {
			t.Errorf("ParseTile(%v): ok=%v, 期望%v", tt.data, ok, tt.ok)
			continue
		}
		if ok && (tile.Suit != tt.suit || tile.Rank != tt.rank) {
			t.Errorf("ParseTile(%v): got %v, 期望 suit=%d rank=%d", tt.data, tile, tt.suit, tt.rank)
		}
	}
}

func TestTilesOperations(t *testing.T) {
	tiles := Tiles{
		{SuitWan, 1}, {SuitWan, 1}, {SuitWan, 2}, {SuitTiao, 3},
	}

	if !tiles.Contains(Tile{SuitWan, 1}) {
		t.Error("应包含1万")
	}
	if tiles.Contains(Tile{SuitWan, 9}) {
		t.Error("不应包含9万")
	}
	if tiles.Count(Tile{SuitWan, 1}) != 2 {
		t.Error("1万应有2张")
	}

	removed := tiles.Remove(Tile{SuitWan, 1})
	if len(removed) != 3 {
		t.Errorf("移除后应有3张，实际%d", len(removed))
	}
	if removed.Count(Tile{SuitWan, 1}) != 1 {
		t.Error("移除后1万应有1张")
	}
}

// ==================== hand_test ====================

func TestIsStandardWin(t *testing.T) {
	// 标准胡牌：4面子+1雀头
	// 123万 456万 789万 111条 + 55筒
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 2}, {SuitWan, 3},
		{SuitWan, 4}, {SuitWan, 5}, {SuitWan, 6},
		{SuitWan, 7}, {SuitWan, 8}, {SuitWan, 9},
		{SuitTiao, 1}, {SuitTiao, 1}, {SuitTiao, 1},
		{SuitTong, 5}, {SuitTong, 5},
	}
	if !IsStandardWin(hand) {
		t.Error("应判定为标准胡牌（4面子+1雀头）")
	}
}

func TestIsStandardWin_PongOnly(t *testing.T) {
	// 碰碰胡：111万 222条 333筒 东东东 + 55万
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 1}, {SuitWan, 1},
		{SuitTiao, 2}, {SuitTiao, 2}, {SuitTiao, 2},
		{SuitTong, 3}, {SuitTong, 3}, {SuitTong, 3},
		{SuitZi, 1}, {SuitZi, 1}, {SuitZi, 1},
		{SuitWan, 5}, {SuitWan, 5},
	}
	if !IsStandardWin(hand) {
		t.Error("应判定为胡牌（碰碰胡）")
	}
}

func TestIsStandardWin_Fail(t *testing.T) {
	// 非胡牌
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 2}, {SuitWan, 3},
		{SuitWan, 4}, {SuitWan, 5}, {SuitWan, 6},
		{SuitWan, 7}, {SuitWan, 8}, {SuitWan, 9},
		{SuitTiao, 1}, {SuitTiao, 2}, {SuitTiao, 3},
		{SuitTong, 5}, {SuitTong, 6}, // 14张但没有雀头
	}
	if IsStandardWin(hand) {
		t.Error("不应判定为胡牌")
	}
}

func TestIsSevenPairs(t *testing.T) {
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 1},
		{SuitWan, 3}, {SuitWan, 3},
		{SuitTiao, 5}, {SuitTiao, 5},
		{SuitTong, 7}, {SuitTong, 7},
		{SuitZi, 1}, {SuitZi, 1},
		{SuitZi, 5}, {SuitZi, 5},
		{SuitZi, 6}, {SuitZi, 6},
	}
	if !IsSevenPairs(hand) {
		t.Error("应判定为七对子")
	}
}

func TestIsThirteenOrphans(t *testing.T) {
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 9},
		{SuitTiao, 1}, {SuitTiao, 9},
		{SuitTong, 1}, {SuitTong, 9},
		{SuitZi, 1}, {SuitZi, 2}, {SuitZi, 3}, {SuitZi, 4},
		{SuitZi, 5}, {SuitZi, 6}, {SuitZi, 7},
		{SuitZi, 1}, // 对子
	}
	if !IsThirteenOrphans(hand) {
		t.Error("应判定为十三幺")
	}
}

func TestCanChow(t *testing.T) {
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 2}, {SuitWan, 5}, {SuitWan, 6},
	}
	// 可以用1万2万吃3万
	results := CanChow(hand, Tile{SuitWan, 3})
	if len(results) == 0 {
		t.Error("应该可以吃3万")
	}

	// 字牌不能吃
	results = CanChow(hand, Tile{SuitZi, 1})
	if len(results) != 0 {
		t.Error("字牌不应该能吃")
	}
}

func TestCanPong(t *testing.T) {
	hand := Tiles{
		{SuitWan, 5}, {SuitWan, 5}, {SuitWan, 3},
	}
	if !CanPong(hand, Tile{SuitWan, 5}) {
		t.Error("应该可以碰5万")
	}
	if CanPong(hand, Tile{SuitWan, 3}) {
		t.Error("不应该可以碰3万（只有1张）")
	}
}

func TestGetWaitingTiles(t *testing.T) {
	// 只差5万即可胡：123万 456万 789万 111条 + 5?筒
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 2}, {SuitWan, 3},
		{SuitWan, 4}, {SuitWan, 5}, {SuitWan, 6},
		{SuitWan, 7}, {SuitWan, 8}, {SuitWan, 9},
		{SuitTiao, 1}, {SuitTiao, 1}, {SuitTiao, 1},
		{SuitTong, 5},
	}
	waiting := GetWaitingTiles(hand, nil)
	if len(waiting) == 0 {
		t.Error("应该有听牌")
	}
	found := false
	for _, w := range waiting {
		if w.Equal(Tile{SuitTong, 5}) {
			found = true
		}
	}
	if !found {
		t.Error("应该听5筒")
	}
}

// ==================== fan_test ====================

func TestCalculateFan_QingYiSe(t *testing.T) {
	// 清一色：123万 456万 789万 111万 + 55万
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 2}, {SuitWan, 3},
		{SuitWan, 4}, {SuitWan, 5}, {SuitWan, 6},
		{SuitWan, 7}, {SuitWan, 8}, {SuitWan, 9},
		{SuitWan, 1}, {SuitWan, 1}, {SuitWan, 1},
		{SuitWan, 5}, {SuitWan, 5},
	}
	ctx := &WinContext{
		Hand:      hand,
		Melds:     nil,
		WinTile:   Tile{SuitWan, 5},
		SelfDrawn: true,
		SeatWind:  1,
		RoundWind: 1,
	}
	fans, total := CalculateFan(ctx)
	if total < 24 {
		t.Errorf("清一色应至少24番，实际%d番, fans=%v", total, fans)
	}
}

func TestCalculateFan_SevenPairs(t *testing.T) {
	hand := Tiles{
		{SuitWan, 1}, {SuitWan, 1},
		{SuitWan, 3}, {SuitWan, 3},
		{SuitTiao, 5}, {SuitTiao, 5},
		{SuitTong, 7}, {SuitTong, 7},
		{SuitZi, 1}, {SuitZi, 1},
		{SuitZi, 5}, {SuitZi, 5},
		{SuitZi, 6}, {SuitZi, 6},
	}
	ctx := &WinContext{
		Hand:      hand,
		Melds:     nil,
		WinTile:   Tile{SuitZi, 6},
		SelfDrawn: true,
		SeatWind:  1,
		RoundWind: 1,
	}
	fans, total := CalculateFan(ctx)
	found := false
	for _, f := range fans {
		if f.Name == "七对" {
			found = true
		}
	}
	if !found {
		t.Errorf("应识别七对，fans=%v, total=%d", fans, total)
	}
}

// ==================== game_test ====================

func TestGameInit(t *testing.T) {
	g := New()

	players := []string{"p1", "p2", "p3", "p4"}
	if err := g.Init(players); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	if g.MaxPlayers() != 4 {
		t.Error("MaxPlayers应为4")
	}
	if g.MinPlayers() != 4 {
		t.Error("MinPlayers应为4")
	}
	if g.IsGameOver() {
		t.Error("游戏不应结束")
	}
	if g.ID() != "mahjong" {
		t.Error("ID应为mahjong")
	}

	mg := g.(*MahjongGame)
	// 庄家14张，闲家各13张
	if len(mg.hands[0]) != 14 {
		t.Errorf("庄家应有14张牌，实际%d", len(mg.hands[0]))
	}
	for i := 1; i < 4; i++ {
		if len(mg.hands[i]) != 13 {
			t.Errorf("座位%d应有13张牌，实际%d", i, len(mg.hands[i]))
		}
	}
	// 牌墙应剩余 136 - 14 - 13*3 = 83 张
	if len(mg.wall) != 83 {
		t.Errorf("牌墙应剩余83张，实际%d", len(mg.wall))
	}
}

func TestGameInit_WrongPlayers(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2", "p3"}); err == nil {
		t.Error("3人应返回错误")
	}
}

func TestGameDiscard(t *testing.T) {
	g := New()
	players := []string{"p1", "p2", "p3", "p4"}
	g.Init(players)

	mg := g.(*MahjongGame)

	// 庄家（座位0）出第一张手牌
	tile := mg.hands[0][0]
	action := interfaces.Action{
		Type: "discard",
		Data: tile.ToMap(),
	}

	_, err := g.ProcessAction("p1", action)
	if err != nil {
		t.Fatalf("出牌失败: %v", err)
	}

	// 出牌后庄家应该少一张牌
	if mg.phase == PhasePlay {
		// 如果无人响应直接推进，下家已摸牌
		// 座位1现在应该有14张（13+摸1张）
		if len(mg.hands[1]) != 14 {
			t.Errorf("下家摸牌后应有14张，实际%d", len(mg.hands[1]))
		}
	}
}

func TestGameDiscard_NotYourTurn(t *testing.T) {
	g := New()
	players := []string{"p1", "p2", "p3", "p4"}
	g.Init(players)

	mg := g.(*MahjongGame)
	tile := mg.hands[1][0] // 座位1的牌

	action := interfaces.Action{
		Type: "discard",
		Data: tile.ToMap(),
	}

	// 座位1不是当前回合（庄家座位0先出）
	_, err := g.ProcessAction("p2", action)
	if err == nil {
		t.Error("非当前回合应返回错误")
	}
}

func TestGetStateForPlayer(t *testing.T) {
	g := New()
	players := []string{"p1", "p2", "p3", "p4"}
	g.Init(players)

	state := g.GetStateForPlayer("p1")
	stateMap, ok := state.(map[string]interface{})
	if !ok {
		t.Fatal("状态应为map")
	}

	// 应包含手牌
	if _, ok := stateMap["my_hand"]; !ok {
		t.Error("应包含my_hand")
	}
	// 应包含可执行操作
	if _, ok := stateMap["available_actions"]; !ok {
		t.Error("庄家应有可执行操作")
	}
	// 应包含座位信息
	if stateMap["my_seat"] != 0 {
		t.Error("p1应在座位0")
	}
}

func TestFullGameFlow_BasicDiscard(t *testing.T) {
	g := New()
	players := []string{"p1", "p2", "p3", "p4"}
	g.Init(players)
	mg := g.(*MahjongGame)

	// 模拟连续几轮出牌
	for round := 0; round < 4; round++ {
		if mg.gameOver {
			break
		}

		currentPlayer := mg.players[mg.currentSeat]
		hand := mg.hands[mg.currentSeat]
		if len(hand) == 0 {
			break
		}

		// 出最后一张牌
		tile := hand[len(hand)-1]
		action := interfaces.Action{
			Type: "discard",
			Data: tile.ToMap(),
		}

		_, err := g.ProcessAction(currentPlayer, action)
		if err != nil {
			t.Fatalf("第%d轮出牌失败(%s): %v", round, currentPlayer, err)
		}

		// 如果进入pending阶段，所有有动作的人都pass
		if mg.phase == PhasePending {
			for seat := range mg.pendingActions {
				pid := mg.players[seat]
				passAction := interfaces.Action{Type: "pass"}
				_, err := g.ProcessAction(pid, passAction)
				if err != nil {
					t.Fatalf("pass失败(%s): %v", pid, err)
				}
			}
		}
	}
}
