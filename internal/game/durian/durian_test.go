package durian

import (
	"testing"

	"card-game-server/internal/game/interfaces"
)

// ─────────────────────────────────────────────
// 卡牌基础测试
// ─────────────────────────────────────────────

func TestFruitCard_GetCardType(t *testing.T) {
	card := &FruitCard{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitBanana, RightCount: 3}
	if card.GetCardType() != CardTypeFruit {
		t.Errorf("期望 CardTypeFruit，得到 %s", card.GetCardType())
	}
	if card.GetID() != 1 {
		t.Errorf("期望 ID=1，得到 %d", card.GetID())
	}
}

func TestFruitCard_InventoryContribution(t *testing.T) {
	card := &FruitCard{
		ID:         1,
		Type:       CardTypeFruit,
		LeftFruit:  FruitDurian,
		LeftCount:  2,
		RightFruit: FruitBanana,
		RightCount: 3,
	}
	inv := card.InventoryContribution()
	if inv[FruitDurian] != 2 {
		t.Errorf("榴莲库存期望 2，得到 %d", inv[FruitDurian])
	}
	if inv[FruitBanana] != 3 {
		t.Errorf("香蕉库存期望 3，得到 %d", inv[FruitBanana])
	}
}

func TestFruitCard_SameFruitBothSides(t *testing.T) {
	// 左右两面相同水果（如牌ID=15: 左4榴莲 右1榴莲）
	card := &FruitCard{
		ID:         15,
		Type:       CardTypeFruit,
		LeftFruit:  FruitDurian,
		LeftCount:  4,
		RightFruit: FruitDurian,
		RightCount: 1,
	}
	inv := card.InventoryContribution()
	if inv[FruitDurian] != 5 {
		t.Errorf("同面水果库存期望 5，得到 %d", inv[FruitDurian])
	}
}

func TestGorillaCard_GetCardType(t *testing.T) {
	card := &GorillaCard{ID: 101, Type: CardTypeGorilla, Ability: AbilityCancelDurian}
	if card.GetCardType() != CardTypeGorilla {
		t.Errorf("期望 CardTypeGorilla，得到 %s", card.GetCardType())
	}
	if card.GetID() != 101 {
		t.Errorf("期望 ID=101，得到 %d", card.GetID())
	}
}

// ─────────────────────────────────────────────
// 牌堆测试
// ─────────────────────────────────────────────

func TestNewDeck_CardCount(t *testing.T) {
	deck := NewDeck()
	if deck.Remaining() != 31 {
		t.Errorf("期望31张牌，得到 %d 张", deck.Remaining())
	}
}

func TestNewDeck_UniqueIDs(t *testing.T) {
	deck := NewDeck()
	seen := make(map[int]bool)
	for deck.Remaining() > 0 {
		card, ok := deck.Draw()
		if !ok {
			t.Fatal("Draw 返回 false，但牌堆不应该空")
		}
		id := card.GetID()
		if seen[id] {
			t.Errorf("重复的牌 ID: %d", id)
		}
		seen[id] = true
	}
	if len(seen) != 31 {
		t.Errorf("期望31张不重复的牌，得到 %d 张", len(seen))
	}
}

func TestDeck_Shuffle(t *testing.T) {
	deck1 := NewDeck()
	deck2 := NewDeck()
	deck1.Shuffle()
	deck2.Shuffle()
	if deck1.Remaining() != 31 {
		t.Errorf("洗牌后牌数应仍为31，得到 %d", deck1.Remaining())
	}
	if deck2.Remaining() != 31 {
		t.Errorf("洗牌后牌数应仍为31，得到 %d", deck2.Remaining())
	}
}

func TestDeck_DrawAll(t *testing.T) {
	deck := NewDeck()
	deck.Shuffle()
	for i := 0; i < 31; i++ {
		_, ok := deck.Draw()
		if !ok {
			t.Fatalf("第 %d 次 Draw 失败，牌堆不应该空", i+1)
		}
	}
	if deck.Remaining() != 0 {
		t.Errorf("抽完后剩余应为0，得到 %d", deck.Remaining())
	}
}

func TestDeck_ReshuffleFromDiscard(t *testing.T) {
	deck := NewDeck()
	// 抽出所有牌并丢入废牌堆
	cards := make([]Card, 0, 31)
	for deck.Remaining() > 0 {
		c, _ := deck.Draw()
		cards = append(cards, c)
	}
	for _, c := range cards {
		deck.Discard(c)
	}
	// 废牌堆重洗后应能继续抽牌
	c, ok := deck.Draw()
	if !ok {
		t.Fatal("废牌堆重洗后 Draw 应该成功")
	}
	if c == nil {
		t.Fatal("重洗后抽到的牌不应为 nil")
	}
}

// ─────────────────────────────────────────────
// 愤怒标记测试
// ─────────────────────────────────────────────

func TestNewAngerTokenPool(t *testing.T) {
	pool := NewAngerTokenPool()
	if len(pool.Remaining) != 7 {
		t.Errorf("期望7枚标记，得到 %d", len(pool.Remaining))
	}
	for i, token := range pool.Remaining {
		expected := i + 1
		if token.Value != expected {
			t.Errorf("标记[%d] 期望分值 %d，得到 %d", i, expected, token.Value)
		}
	}
}

func TestAngerTokenPool_TakeSmallest(t *testing.T) {
	pool := NewAngerTokenPool()

	token, err := pool.TakeSmallest()
	if err != nil {
		t.Fatalf("TakeSmallest 应成功，得到错误: %v", err)
	}
	if token.Value != 1 {
		t.Errorf("第1次取标记期望分值1，得到 %d", token.Value)
	}
	if len(pool.Remaining) != 6 {
		t.Errorf("取走1枚后期望剩余6枚，得到 %d", len(pool.Remaining))
	}

	token2, err := pool.TakeSmallest()
	if err != nil {
		t.Fatalf("第2次 TakeSmallest 应成功")
	}
	if token2.Value != 2 {
		t.Errorf("第2次取标记期望分值2，得到 %d", token2.Value)
	}
}

func TestAngerTokenPool_IsEmpty(t *testing.T) {
	pool := NewAngerTokenPool()
	if pool.IsEmpty() {
		t.Error("新建标记池不应为空")
	}
	for !pool.IsEmpty() {
		pool.TakeSmallest() //nolint:errcheck
	}
	if !pool.IsEmpty() {
		t.Error("取完所有标记后应为空")
	}
	_, err := pool.TakeSmallest()
	if err == nil {
		t.Error("空标记池 TakeSmallest 应返回错误")
	}
}

func TestPlayerAnger_TotalScore(t *testing.T) {
	pa := &PlayerAnger{PlayerID: "p1"}
	if pa.TotalScore() != 0 {
		t.Errorf("初始总分应为0，得到 %d", pa.TotalScore())
	}
	pa.AddToken(&AngerToken{Value: 1})
	pa.AddToken(&AngerToken{Value: 3})
	if pa.TotalScore() != 4 {
		t.Errorf("总分期望4，得到 %d", pa.TotalScore())
	}
}

func TestPlayerAnger_IsEliminated(t *testing.T) {
	pa := &PlayerAnger{PlayerID: "p1"}
	pa.AddToken(&AngerToken{Value: 3})
	pa.AddToken(&AngerToken{Value: 3})
	if pa.IsEliminated() {
		t.Error("总分6不应触发淘汰")
	}
	pa.AddToken(&AngerToken{Value: 1})
	if !pa.IsEliminated() {
		t.Error("总分7应触发淘汰")
	}
}

// ─────────────────────────────────────────────
// 结算逻辑测试
// ─────────────────────────────────────────────

func TestCalculateInventory_OnlyFruitCards(t *testing.T) {
	holderCards := map[string]Card{
		"p1": &FruitCard{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitBanana, RightCount: 3},
		"p2": &FruitCard{ID: 2, Type: CardTypeFruit, LeftFruit: FruitGrape, LeftCount: 1, RightFruit: FruitStrawberry, RightCount: 4},
	}
	inv := CalculateInventory(holderCards)
	if inv[FruitDurian] != 2 {
		t.Errorf("榴莲库存期望2，得到 %d", inv[FruitDurian])
	}
	if inv[FruitBanana] != 3 {
		t.Errorf("香蕉库存期望3，得到 %d", inv[FruitBanana])
	}
	if inv[FruitGrape] != 1 {
		t.Errorf("葡萄库存期望1，得到 %d", inv[FruitGrape])
	}
	if inv[FruitStrawberry] != 4 {
		t.Errorf("草莓库存期望4，得到 %d", inv[FruitStrawberry])
	}
}

func TestCalculateInventory_GorillaCardNoContribution(t *testing.T) {
	holderCards := map[string]Card{
		"p1": &FruitCard{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitBanana, RightCount: 3},
		"p2": &GorillaCard{ID: 101, Type: CardTypeGorilla, Ability: AbilityCancelDurian},
	}
	inv := CalculateInventory(holderCards)
	if inv[FruitDurian] != 2 {
		t.Errorf("猩猩牌不贡献库存，榴莲期望2，得到 %d", inv[FruitDurian])
	}
	if inv[FruitBanana] != 3 {
		t.Errorf("猩猩牌不贡献库存，香蕉期望3，得到 %d", inv[FruitBanana])
	}
}

func TestApplyGorillaEffects_CancelDurian(t *testing.T) {
	orders := map[FruitType]int{
		FruitDurian:     3,
		FruitBanana:     2,
		FruitGrape:      1,
		FruitStrawberry: 0,
	}
	holderCards := map[string]Card{
		"p1": &GorillaCard{ID: 101, Type: CardTypeGorilla, Ability: AbilityCancelDurian},
	}
	result, effects := ApplyGorillaEffects(orders, holderCards)
	if result[FruitDurian] != 0 {
		t.Errorf("猩猩大哥应取消榴莲订单，期望0，得到 %d", result[FruitDurian])
	}
	if result[FruitBanana] != 2 {
		t.Errorf("香蕉订单不应受影响，期望2，得到 %d", result[FruitBanana])
	}
	if len(effects) != 1 {
		t.Errorf("期望1个猩猩特效，得到 %d", len(effects))
	}
	if effects[0].CancelledOrders != 3 {
		t.Errorf("取消的订单量期望3，得到 %d", effects[0].CancelledOrders)
	}
}

func TestCheckShortage_IsShortage(t *testing.T) {
	orders := map[FruitType]int{FruitDurian: 5, FruitBanana: 2}
	inventory := map[FruitType]int{FruitDurian: 3, FruitBanana: 4}
	isShort, fruits := CheckShortage(orders, inventory)
	if !isShort {
		t.Error("榴莲缺货，应返回 true")
	}
	if len(fruits) != 1 || fruits[0] != FruitDurian {
		t.Errorf("缺货水果应为榴莲，得到 %v", fruits)
	}
}

func TestCheckShortage_NotShortage(t *testing.T) {
	orders := map[FruitType]int{FruitDurian: 2, FruitBanana: 2}
	inventory := map[FruitType]int{FruitDurian: 3, FruitBanana: 4, FruitGrape: 0, FruitStrawberry: 0}
	isShort, fruits := CheckShortage(orders, inventory)
	if isShort {
		t.Error("库存充足，应返回 false")
	}
	if len(fruits) != 0 {
		t.Errorf("不缺货时缺货列表应为空，得到 %v", fruits)
	}
}

func TestDoSettlement_ShortageLastOrderPunished(t *testing.T) {
	players := []string{"p1", "p2", "p3"}
	holderCards := map[string]Card{
		"p1": &FruitCard{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitBanana, RightCount: 1},
		"p2": &FruitCard{ID: 2, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitBanana, RightCount: 1},
		"p3": &FruitCard{ID: 3, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitBanana, RightCount: 1},
	}
	// 订单：香蕉需求5 > 库存3，缺货
	orders := map[FruitType]int{
		FruitDurian: 1, FruitBanana: 5, FruitGrape: 0, FruitStrawberry: 0,
	}
	tokenPool := NewAngerTokenPool()
	playerAnger := map[string]*PlayerAnger{
		"p1": {PlayerID: "p1"},
		"p2": {PlayerID: "p2"},
		"p3": {PlayerID: "p3"},
	}

	// p2(索引1) 是最后接订单的人，p3 摇铃
	result, err := DoSettlement("p3", 1, players, holderCards, orders, tokenPool, playerAnger)
	if err != nil {
		t.Fatalf("结算失败: %v", err)
	}
	if !result.IsShortage {
		t.Error("应判定为缺货")
	}
	if result.PunishedPlayer != "p2" {
		t.Errorf("缺货时上一个接订单者(p2)应受罚，得到 %s", result.PunishedPlayer)
	}
	if result.PunishReason != "last_order" {
		t.Errorf("受罚原因期望 last_order，得到 %s", result.PunishReason)
	}
	if result.AngerTokenGiven != 1 {
		t.Errorf("第一次受罚应得到1分标记，得到 %d", result.AngerTokenGiven)
	}
	if playerAnger["p2"].TotalScore() != 1 {
		t.Errorf("p2 愤怒分应为1，得到 %d", playerAnger["p2"].TotalScore())
	}
}

func TestDoSettlement_NotShortageBellRingerPunished(t *testing.T) {
	players := []string{"p1", "p2"}
	holderCards := map[string]Card{
		"p1": &FruitCard{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 3, RightFruit: FruitBanana, RightCount: 3},
		"p2": &FruitCard{ID: 2, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 3, RightFruit: FruitBanana, RightCount: 3},
	}
	orders := map[FruitType]int{
		FruitDurian: 2, FruitBanana: 1, FruitGrape: 0, FruitStrawberry: 0,
	}
	tokenPool := NewAngerTokenPool()
	playerAnger := map[string]*PlayerAnger{
		"p1": {PlayerID: "p1"},
		"p2": {PlayerID: "p2"},
	}

	result, err := DoSettlement("p1", 0, players, holderCards, orders, tokenPool, playerAnger)
	if err != nil {
		t.Fatalf("结算失败: %v", err)
	}
	if result.IsShortage {
		t.Error("库存充足，不应判定为缺货")
	}
	if result.PunishedPlayer != "p1" {
		t.Errorf("库存充足时摇铃者(p1)应受罚，得到 %s", result.PunishedPlayer)
	}
	if result.PunishReason != "bell_ringer" {
		t.Errorf("受罚原因期望 bell_ringer，得到 %s", result.PunishReason)
	}
}

func TestDoSettlement_EmptyOrdersBellRingerPunished(t *testing.T) {
	players := []string{"p1", "p2"}
	holderCards := map[string]Card{
		"p1": &FruitCard{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitBanana, RightCount: 2},
		"p2": &FruitCard{ID: 2, Type: CardTypeFruit, LeftFruit: FruitGrape, LeftCount: 2, RightFruit: FruitStrawberry, RightCount: 2},
	}
	// 订单区为空（没有人接过订单就摇铃）
	orders := map[FruitType]int{
		FruitDurian: 0, FruitBanana: 0, FruitGrape: 0, FruitStrawberry: 0,
	}
	tokenPool := NewAngerTokenPool()
	playerAnger := map[string]*PlayerAnger{
		"p1": {PlayerID: "p1"},
		"p2": {PlayerID: "p2"},
	}

	// lastOrderPlayerIdx = -1 表示无人接过订单
	result, err := DoSettlement("p1", -1, players, holderCards, orders, tokenPool, playerAnger)
	if err != nil {
		t.Fatalf("结算失败: %v", err)
	}
	if result.IsShortage {
		t.Error("订单为空，不缺货，不应判定为缺货")
	}
	if result.PunishedPlayer != "p1" {
		t.Errorf("订单为空时摇铃者(p1)应受罚，得到 %s", result.PunishedPlayer)
	}
}

func TestDoSettlement_GorillaCardCancelsOrder(t *testing.T) {
	players := []string{"p1", "p2", "p3"}
	holderCards := map[string]Card{
		"p1": &FruitCard{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitBanana, RightCount: 1},
		"p2": &FruitCard{ID: 2, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitBanana, RightCount: 1},
		// p3 持有猩猩大哥，取消榴莲订单
		"p3": &GorillaCard{ID: 101, Type: CardTypeGorilla, Ability: AbilityCancelDurian},
	}
	// 榴莲订单5 > 库存2，本来缺货；猩猩牌取消后变为0，不再缺货
	orders := map[FruitType]int{
		FruitDurian: 5, FruitBanana: 1, FruitGrape: 0, FruitStrawberry: 0,
	}
	tokenPool := NewAngerTokenPool()
	playerAnger := map[string]*PlayerAnger{
		"p1": {PlayerID: "p1"},
		"p2": {PlayerID: "p2"},
		"p3": {PlayerID: "p3"},
	}

	// p2(索引1) 最后接订单，p1 摇铃
	result, err := DoSettlement("p1", 1, players, holderCards, orders, tokenPool, playerAnger)
	if err != nil {
		t.Fatalf("结算失败: %v", err)
	}
	if result.IsShortage {
		t.Error("猩猩牌取消榴莲后不应缺货")
	}
	if result.PunishedPlayer != "p1" {
		t.Errorf("不缺货时摇铃者(p1)应受罚，得到 %s", result.PunishedPlayer)
	}
	if len(result.GorillaEffects) != 1 {
		t.Errorf("期望1个猩猩特效，得到 %d", len(result.GorillaEffects))
	}
}

// ─────────────────────────────────────────────
// 游戏流程测试
// ─────────────────────────────────────────────

func TestDurianGame_Init_ValidPlayers(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2", "p3"}); err != nil {
		t.Fatalf("初始化3人游戏失败: %v", err)
	}
	if g.IsGameOver() {
		t.Error("初始化后游戏不应结束")
	}
}

func TestDurianGame_Init_TooFewPlayers(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1"}); err == nil {
		t.Error("1人游戏应返回错误")
	}
}

func TestDurianGame_Init_TooManyPlayers(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}); err == nil {
		t.Error("8人游戏应返回错误")
	}
}

func TestDurianGame_MaxMinPlayers(t *testing.T) {
	g := New()
	if g.MinPlayers() != 2 {
		t.Errorf("最少玩家数应为2，得到 %d", g.MinPlayers())
	}
	if g.MaxPlayers() != 7 {
		t.Errorf("最多玩家数应为7，得到 %d", g.MaxPlayers())
	}
}

func TestDurianGame_GetStateForPlayer_NoSelfCard(t *testing.T) {
	g := New()
	players := []string{"p1", "p2", "p3"}
	if err := g.Init(players); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// 安全性测试：GetStateForPlayer 不应包含自己的牌架卡
	for _, pid := range players {
		state := g.GetStateForPlayer(pid)
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			t.Fatalf("状态应为 map[string]interface{}")
		}

		// your_card 必须永远为 nil
		yourCard, exists := stateMap["your_card"]
		if !exists {
			t.Errorf("玩家 %s 的状态缺少 your_card 字段", pid)
		}
		if yourCard != nil {
			t.Errorf("玩家 %s 的 your_card 应为 nil，信息隔离失败！", pid)
		}

		// others_cards 不应包含自己
		othersCards, ok := stateMap["others_cards"].([]map[string]interface{})
		if ok {
			for _, oc := range othersCards {
				if oc["player_id"] == pid {
					t.Errorf("玩家 %s 的 others_cards 中不应包含自己的牌架卡", pid)
				}
			}
		}
	}
}

func TestDurianGame_TakeOrder_WrongTurn(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	current := g.CurrentTurn()
	wrongPlayer := "p1"
	if current == "p1" {
		wrongPlayer = "p2"
	}

	act := interfaces.Action{Type: "take_order", Data: map[string]interface{}{"chosen_side": "left"}}
	_, err := g.ProcessAction(wrongPlayer, act)
	if err == nil {
		t.Error("非当前回合玩家操作应返回错误")
	}
}

func TestDurianGame_TakeOrder_InvalidSide(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	current := g.CurrentTurn()
	act := interfaces.Action{Type: "take_order", Data: map[string]interface{}{"chosen_side": "top"}}
	_, err := g.ProcessAction(current, act)
	if err == nil {
		t.Error("非法 chosen_side 应返回错误")
	}
}

func TestDurianGame_TakeOrder_Success(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	current := g.CurrentTurn()
	act := interfaces.Action{Type: "take_order", Data: map[string]interface{}{"chosen_side": "left"}}
	turnEnded, err := g.ProcessAction(current, act)
	if err != nil {
		t.Fatalf("take_order 应成功: %v", err)
	}
	if !turnEnded {
		t.Error("take_order 后应返回 turnEnded=true")
	}
}

func TestDurianGame_RingBell_TriggerSettlement(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// 先接一个订单
	current := g.CurrentTurn()
	takeAct := interfaces.Action{Type: "take_order", Data: map[string]interface{}{"chosen_side": "left"}}
	turnEnded, err := g.ProcessAction(current, takeAct)
	if err != nil {
		t.Fatalf("take_order 失败: %v", err)
	}
	if turnEnded {
		g.AdvanceTurn()
	}

	// 再摇铃
	current = g.CurrentTurn()
	bellAct := interfaces.Action{Type: "ring_bell", Data: nil}
	turnEnded, err = g.ProcessAction(current, bellAct)
	if err != nil {
		t.Fatalf("ring_bell 失败: %v", err)
	}
	if turnEnded {
		t.Error("ring_bell 后应返回 turnEnded=false")
	}

	state := g.GetState().(map[string]interface{})
	phase := state["phase"].(string)
	if phase == "settlement" {
		t.Error("ring_bell 后不应停留在 settlement 阶段")
	}
}

func TestDurianGame_GameOver_WhenAngerScoreReaches7(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	dg := g.(*DurianGame)
	// 手动给 p1 累积 6 分（1+2+3）
	dg.playerAnger["p1"].AddToken(&AngerToken{Value: 1})
	dg.playerAnger["p1"].AddToken(&AngerToken{Value: 2})
	dg.playerAnger["p1"].AddToken(&AngerToken{Value: 3})
	// 模拟前3枚标记已取走
	dg.tokenPool.Remaining = dg.tokenPool.Remaining[3:]

	// 设置缺货场景：p1 是最后接订单者，摇铃后 p1 再受罚（取4分标记，总分=10>=7）
	dg.lastOrderPlayerIdx = 0
	dg.orders[FruitBanana] = 100

	current := g.CurrentTurn()
	bellAct := interfaces.Action{Type: "ring_bell", Data: nil}
	g.ProcessAction(current, bellAct) //nolint:errcheck

	if !g.IsGameOver() {
		t.Error("p1 愤怒分应达到>=7，游戏应结束")
	}
	if g.Winner() == "" {
		t.Error("游戏结束后应有获胜者")
	}
}

func TestDurianGame_Winner_LowestAngerScore(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2", "p3"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	dg := g.(*DurianGame)
	dg.playerAnger["p1"].AddToken(&AngerToken{Value: 3})
	dg.playerAnger["p2"].AddToken(&AngerToken{Value: 1})
	dg.playerAnger["p3"].AddToken(&AngerToken{Value: 2})
	dg.gameOver = true
	dg.winner = dg.determineWinner()

	if g.Winner() != "p2" {
		t.Errorf("p2 分数最低(1)，应获胜，得到 %s", g.Winner())
	}
}

func TestDurianGame_Winner_TiedLowestAngerScore(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2", "p3"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	dg := g.(*DurianGame)
	dg.playerAnger["p1"].AddToken(&AngerToken{Value: 1})
	dg.playerAnger["p2"].AddToken(&AngerToken{Value: 1})
	dg.playerAnger["p3"].AddToken(&AngerToken{Value: 3})
	dg.gameOver = true
	dg.winner = dg.determineWinner()

	winner := g.Winner()
	if winner != "p1,p2" && winner != "p2,p1" {
		t.Errorf("p1和p2并列最低，应共同获胜，得到 %s", winner)
	}
}

func TestDurianGame_FullFlow_2Players(t *testing.T) {
	g := New()
	if err := g.Init([]string{"p1", "p2"}); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// p1 接订单
	current := g.CurrentTurn()
	takeAct := interfaces.Action{Type: "take_order", Data: map[string]interface{}{"chosen_side": "right"}}
	turnEnded, err := g.ProcessAction(current, takeAct)
	if err != nil {
		t.Fatalf("take_order 失败: %v", err)
	}
	if turnEnded {
		g.AdvanceTurn()
	}

	// 另一个玩家摇铃
	current = g.CurrentTurn()
	bellAct := interfaces.Action{Type: "ring_bell", Data: nil}
	_, err = g.ProcessAction(current, bellAct)
	if err != nil {
		t.Fatalf("ring_bell 失败: %v", err)
	}

	if g.IsGameOver() {
		t.Log("游戏已结束，winner:", g.Winner())
		return
	}

	state := g.GetState().(map[string]interface{})
	phase := state["phase"].(string)
	if phase != "playing" {
		t.Errorf("新一轮应为 playing 阶段，得到 %s", phase)
	}
	roundNum := state["round_number"].(int)
	if roundNum < 2 {
		t.Errorf("轮次应至少为2，得到 %d", roundNum)
	}
}
