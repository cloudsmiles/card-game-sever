package mahjong

// FanItem 一个番种
type FanItem struct {
	Name string `json:"name"`
	Fan  int    `json:"fan"`
}

// WinContext 胡牌上下文，用于番种计算
type WinContext struct {
	Hand       Tiles  // 手牌（含胡的那张牌，共14张）
	Melds      []Meld // 副露
	WinTile    Tile   // 胡的那张牌
	SelfDrawn  bool   // 是否自摸
	DealerSeat int    // 庄家座位
	WinnerSeat int    // 胡牌者座位
	SeatWind   int    // 门风 (1=东,2=南,3=西,4=北)
	RoundWind  int    // 圈风 (1=东,2=南,3=西,4=北)
}

// CalculateFan 计算番数，返回番种列表和总番数
func CalculateFan(ctx *WinContext) ([]*FanItem, int) {
	var fans []*FanItem

	// 特殊牌型优先判断
	if IsThirteenOrphans(ctx.Hand) {
		fans = append(fans, &FanItem{Name: "十三幺", Fan: 88})
		return fans, 88
	}

	if IsSevenPairs(ctx.Hand) {
		fans = append(fans, &FanItem{Name: "七对", Fan: 24})
	}

	// 收集所有面子信息用于分析
	allMelds := collectAllMelds(ctx)

	// === 88番 ===
	if isDaSiXi(allMelds) {
		fans = append(fans, &FanItem{Name: "大四喜", Fan: 88})
		return fans, capFan(fans)
	}
	if isDaSanYuan(allMelds) {
		fans = append(fans, &FanItem{Name: "大三元", Fan: 88})
		return fans, capFan(fans)
	}

	// === 64番 ===
	if isXiaoSiXi(allMelds) {
		fans = append(fans, &FanItem{Name: "小四喜", Fan: 64})
	}
	if isXiaoSanYuan(allMelds) {
		fans = append(fans, &FanItem{Name: "小三元", Fan: 64})
	}
	if isZiYiSe(ctx.Hand, ctx.Melds) {
		fans = append(fans, &FanItem{Name: "字一色", Fan: 64})
	}
	if isSiAnKe(ctx, allMelds) {
		fans = append(fans, &FanItem{Name: "四暗刻", Fan: 64})
	}

	// === 24番 ===
	if isQingYiSe(ctx.Hand, ctx.Melds) {
		fans = append(fans, &FanItem{Name: "清一色", Fan: 24})
	}

	// === 16番 ===
	if isSanAnKe(ctx, allMelds) {
		fans = append(fans, &FanItem{Name: "三暗刻", Fan: 16})
	}

	// === 8番 ===
	if isHunYiSe(ctx.Hand, ctx.Melds) {
		fans = append(fans, &FanItem{Name: "混一色", Fan: 8})
	}

	// === 6番 ===
	if isPengPengHu(allMelds) {
		fans = append(fans, &FanItem{Name: "碰碰胡", Fan: 6})
	}

	// === 2番 ===
	if checkJianKe(allMelds) > 0 {
		for i := 0; i < checkJianKe(allMelds); i++ {
			fans = append(fans, &FanItem{Name: "箭刻", Fan: 2})
		}
	}
	if isMenQianQing(ctx) {
		fans = append(fans, &FanItem{Name: "门前清", Fan: 2})
	}
	if hasSeatWindPong(allMelds, ctx.SeatWind) {
		fans = append(fans, &FanItem{Name: "门风刻", Fan: 2})
	}
	if hasRoundWindPong(allMelds, ctx.RoundWind) {
		fans = append(fans, &FanItem{Name: "圈风刻", Fan: 2})
	}

	// === 1番 ===
	if ctx.SelfDrawn {
		fans = append(fans, &FanItem{Name: "自摸", Fan: 1})
	}
	kongCounts := countKongs(ctx.Melds)
	for i := 0; i < kongCounts[0]; i++ {
		fans = append(fans, &FanItem{Name: "明杠", Fan: 1})
	}
	for i := 0; i < kongCounts[1]; i++ {
		fans = append(fans, &FanItem{Name: "暗杠", Fan: 1})
	}
	yaoJiuKe := countYaoJiuKe(allMelds)
	for i := 0; i < yaoJiuKe; i++ {
		fans = append(fans, &FanItem{Name: "幺九刻", Fan: 1})
	}

	total := capFan(fans)
	// 如果没有识别到任何番，给予鸡胡（无番，但满足牌型）
	if total == 0 {
		fans = append(fans, &FanItem{Name: "无番和", Fan: 8})
		total = 8
	}
	return fans, total
}

// MeldInfo 面子分析信息
type MeldInfo struct {
	Type      string // "chow", "pong", "kong", "pair"
	Tile      Tile   // 基准牌
	Concealed bool   // 是否暗的
}

// collectAllMelds 从手牌中分解出所有面子（用于番种判断）
func collectAllMelds(ctx *WinContext) []MeldInfo {
	var result []MeldInfo

	// 已副露的面子
	for _, m := range ctx.Melds {
		info := MeldInfo{Tile: m.BaseTile(), Concealed: m.IsConcealed()}
		switch m.Type {
		case MeldChow:
			info.Type = "chow"
		case MeldPong:
			info.Type = "pong"
		case MeldKongExposed, MeldKongExtended:
			info.Type = "kong"
		case MeldKongConcealed:
			info.Type = "kong"
			info.Concealed = true
		}
		result = append(result, info)
	}

	// 分解手牌中的面子
	handMelds := decomposeHand(ctx.Hand)
	result = append(result, handMelds...)
	return result
}

// decomposeHand 分解手牌为面子+雀头（返回第一种有效分解）
func decomposeHand(hand Tiles) []MeldInfo {
	counts := hand.ToCountArray()
	var result []MeldInfo

	for i := 0; i < 34; i++ {
		if counts[i] >= 2 {
			counts[i] -= 2
			pair := TileFromIndex(i)
			melds := decomposeMeldsCollect(counts)
			if melds != nil {
				result = append(result, MeldInfo{Type: "pair", Tile: pair, Concealed: true})
				result = append(result, melds...)
				return result
			}
			counts[i] += 2
		}
	}
	return result
}

func decomposeMeldsCollect(counts [34]int) []MeldInfo {
	remaining := 0
	for _, c := range counts {
		remaining += c
	}
	if remaining == 0 {
		return []MeldInfo{}
	}

	idx := -1
	for i := 0; i < 34; i++ {
		if counts[i] > 0 {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}

	suit := idx / 9
	rank := idx % 9

	// 尝试刻子
	if counts[idx] >= 3 {
		counts[idx] -= 3
		sub := decomposeMeldsCollect(counts)
		if sub != nil {
			counts[idx] += 3
			return append([]MeldInfo{{Type: "pong", Tile: TileFromIndex(idx), Concealed: true}}, sub...)
		}
		counts[idx] += 3
	}

	// 尝试顺子
	if suit < 3 && rank <= 6 {
		base := suit * 9
		i1, i2, i3 := base+rank, base+rank+1, base+rank+2
		if counts[i1] > 0 && counts[i2] > 0 && counts[i3] > 0 {
			counts[i1]--
			counts[i2]--
			counts[i3]--
			sub := decomposeMeldsCollect(counts)
			if sub != nil {
				counts[i1]++
				counts[i2]++
				counts[i3]++
				return append([]MeldInfo{{Type: "chow", Tile: TileFromIndex(i1), Concealed: true}}, sub...)
			}
			counts[i1]++
			counts[i2]++
			counts[i3]++
		}
	}

	return nil
}

// === 番种判断函数 ===

// 大四喜：4副风刻
func isDaSiXi(melds []MeldInfo) bool {
	windCount := 0
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") && m.Tile.IsWind() {
			windCount++
		}
	}
	return windCount == 4
}

// 大三元：3副箭刻
func isDaSanYuan(melds []MeldInfo) bool {
	dragonCount := 0
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") && m.Tile.IsDragon() {
			dragonCount++
		}
	}
	return dragonCount == 3
}

// 小四喜：3副风刻+1对风将
func isXiaoSiXi(melds []MeldInfo) bool {
	windPongCount := 0
	windPairCount := 0
	for _, m := range melds {
		if m.Tile.IsWind() {
			if m.Type == "pong" || m.Type == "kong" {
				windPongCount++
			} else if m.Type == "pair" {
				windPairCount++
			}
		}
	}
	return windPongCount == 3 && windPairCount == 1
}

// 小三元：2副箭刻+1对箭将
func isXiaoSanYuan(melds []MeldInfo) bool {
	dragonPongCount := 0
	dragonPairCount := 0
	for _, m := range melds {
		if m.Tile.IsDragon() {
			if m.Type == "pong" || m.Type == "kong" {
				dragonPongCount++
			} else if m.Type == "pair" {
				dragonPairCount++
			}
		}
	}
	return dragonPongCount == 2 && dragonPairCount == 1
}

// 字一色：全部字牌
func isZiYiSe(hand Tiles, melds []Meld) bool {
	for _, t := range hand {
		if !t.IsHonor() {
			return false
		}
	}
	for _, m := range melds {
		for _, t := range m.Tiles {
			if !t.IsHonor() {
				return false
			}
		}
	}
	return true
}

// 四暗刻：4组暗刻+自摸
func isSiAnKe(ctx *WinContext, melds []MeldInfo) bool {
	if !ctx.SelfDrawn {
		return false
	}
	concealedPongCount := 0
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") && m.Concealed {
			concealedPongCount++
		}
	}
	return concealedPongCount >= 4
}

// 清一色：单一花色（无字牌）
func isQingYiSe(hand Tiles, melds []Meld) bool {
	suit := Suit(-1)
	for _, t := range hand {
		if t.IsHonor() {
			return false
		}
		if suit == -1 {
			suit = t.Suit
		} else if t.Suit != suit {
			return false
		}
	}
	for _, m := range melds {
		for _, t := range m.Tiles {
			if t.IsHonor() {
				return false
			}
			if suit == -1 {
				suit = t.Suit
			} else if t.Suit != suit {
				return false
			}
		}
	}
	return suit >= 0
}

// 三暗刻
func isSanAnKe(ctx *WinContext, melds []MeldInfo) bool {
	concealedPongCount := 0
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") && m.Concealed {
			concealedPongCount++
		}
	}
	return concealedPongCount >= 3
}

// 混一色：单一花色+字牌
func isHunYiSe(hand Tiles, melds []Meld) bool {
	suit := Suit(-1)
	hasHonor := false
	hasNumbered := false

	check := func(t Tile) bool {
		if t.IsHonor() {
			hasHonor = true
			return true
		}
		hasNumbered = true
		if suit == -1 {
			suit = t.Suit
		} else if t.Suit != suit {
			return false
		}
		return true
	}

	for _, t := range hand {
		if !check(t) {
			return false
		}
	}
	for _, m := range melds {
		for _, t := range m.Tiles {
			if !check(t) {
				return false
			}
		}
	}
	return hasHonor && hasNumbered
}

// 碰碰胡：4组刻子/杠+将
func isPengPengHu(melds []MeldInfo) bool {
	for _, m := range melds {
		if m.Type == "chow" {
			return false
		}
	}
	pongCount := 0
	pairCount := 0
	for _, m := range melds {
		if m.Type == "pong" || m.Type == "kong" {
			pongCount++
		} else if m.Type == "pair" {
			pairCount++
		}
	}
	return pongCount >= 4 && pairCount >= 1
}

// 箭刻：中/发/白的刻子数量
func checkJianKe(melds []MeldInfo) int {
	count := 0
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") && m.Tile.IsDragon() {
			count++
		}
	}
	return count
}

// 门前清：无吃碰明杠
func isMenQianQing(ctx *WinContext) bool {
	if !ctx.SelfDrawn {
		return false
	}
	for _, m := range ctx.Melds {
		if m.Type != MeldKongConcealed {
			return false
		}
	}
	return true
}

// 门风刻
func hasSeatWindPong(melds []MeldInfo, seatWind int) bool {
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") &&
			m.Tile.Suit == SuitZi && m.Tile.Rank == seatWind {
			return true
		}
	}
	return false
}

// 圈风刻
func hasRoundWindPong(melds []MeldInfo, roundWind int) bool {
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") &&
			m.Tile.Suit == SuitZi && m.Tile.Rank == roundWind {
			return true
		}
	}
	return false
}

// countKongs 统计杠的数量 [明杠数, 暗杠数]
func countKongs(melds []Meld) [2]int {
	var counts [2]int
	for _, m := range melds {
		if m.Type == MeldKongExposed || m.Type == MeldKongExtended {
			counts[0]++
		} else if m.Type == MeldKongConcealed {
			counts[1]++
		}
	}
	return counts
}

// countYaoJiuKe 统计幺九刻（1或9的刻子）
func countYaoJiuKe(melds []MeldInfo) int {
	count := 0
	for _, m := range melds {
		if (m.Type == "pong" || m.Type == "kong") && m.Tile.IsTerminal() {
			count++
		}
	}
	return count
}

// capFan 计算总番数（封顶88）
func capFan(fans []*FanItem) int {
	total := 0
	for _, f := range fans {
		total += f.Fan
	}
	if total > 88 {
		total = 88
	}
	return total
}
