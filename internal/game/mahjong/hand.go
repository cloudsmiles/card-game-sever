package mahjong

// CanWin 判断手牌+副露是否满足胡牌条件
// hand: 手牌（14张用于自摸, 或13张+外来牌）
// melds: 已有副露
func CanWin(hand Tiles, melds []Meld) bool {
	if IsSevenPairs(hand) {
		return true
	}
	if IsThirteenOrphans(hand) {
		return true
	}
	return IsStandardWin(hand)
}

// IsStandardWin 判断手牌是否满足 4面子+1雀头（标准胡牌型）
// hand 为自由牌（不含已副露的牌），需要恰好 3*n+2 张
func IsStandardWin(hand Tiles) bool {
	total := len(hand)
	if total%3 != 2 {
		return false
	}
	counts := hand.ToCountArray()
	return canDecompose(counts, total)
}

// canDecompose 递归分解：先选雀头，再分解面子
func canDecompose(counts [34]int, remaining int) bool {
	if remaining == 0 {
		return true
	}
	if remaining%3 != 2 && remaining%3 != 0 {
		return false
	}

	// 需要选雀头时（remaining%3==2）
	if remaining%3 == 2 {
		for i := 0; i < 34; i++ {
			if counts[i] >= 2 {
				counts[i] -= 2
				if canDecomposeMelds(counts, remaining-2) {
					counts[i] += 2
					return true
				}
				counts[i] += 2
			}
		}
		return false
	}
	return canDecomposeMelds(counts, remaining)
}

// canDecomposeMelds 递归分解面子（刻子+顺子），remaining 必须是3的倍数
func canDecomposeMelds(counts [34]int, remaining int) bool {
	if remaining == 0 {
		return true
	}

	// 找到第一张有牌的位置
	idx := -1
	for i := 0; i < 34; i++ {
		if counts[i] > 0 {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}

	suit := idx / 9
	rank := idx % 9

	// 尝试刻子
	if counts[idx] >= 3 {
		counts[idx] -= 3
		if canDecomposeMelds(counts, remaining-3) {
			counts[idx] += 3
			return true
		}
		counts[idx] += 3
	}

	// 尝试顺子（只有万条筒可以，字牌不行）
	if suit < 3 && rank <= 6 {
		base := suit * 9
		i1, i2, i3 := base+rank, base+rank+1, base+rank+2
		if counts[i1] > 0 && counts[i2] > 0 && counts[i3] > 0 {
			counts[i1]--
			counts[i2]--
			counts[i3]--
			if canDecomposeMelds(counts, remaining-3) {
				counts[i1]++
				counts[i2]++
				counts[i3]++
				return true
			}
			counts[i1]++
			counts[i2]++
			counts[i3]++
		}
	}

	return false
}

// IsSevenPairs 判断是否七对子（14张手牌，7个对子）
func IsSevenPairs(hand Tiles) bool {
	if len(hand) != 14 {
		return false
	}
	counts := hand.ToCountArray()
	pairs := 0
	for _, c := range counts {
		if c%2 != 0 {
			return false
		}
		pairs += c / 2
	}
	return pairs == 7
}

// IsThirteenOrphans 判断是否十三幺（1/9万条筒 + 东南西北中发白 各1张，其中1张做对子）
func IsThirteenOrphans(hand Tiles) bool {
	if len(hand) != 14 {
		return false
	}
	required := []Tile{
		{SuitWan, 1}, {SuitWan, 9},
		{SuitTiao, 1}, {SuitTiao, 9},
		{SuitTong, 1}, {SuitTong, 9},
		{SuitZi, 1}, {SuitZi, 2}, {SuitZi, 3}, {SuitZi, 4},
		{SuitZi, 5}, {SuitZi, 6}, {SuitZi, 7},
	}
	counts := hand.ToCountArray()
	hasPair := false
	for _, t := range required {
		idx := t.ToIndex()
		if counts[idx] == 0 {
			return false
		}
		if counts[idx] == 2 {
			hasPair = true
		}
	}
	return hasPair
}

// GetWaitingTiles 获取听牌列表（当前13张手牌，哪些牌可以胡）
func GetWaitingTiles(hand Tiles, melds []Meld) Tiles {
	var waiting Tiles
	// 遍历所有34种牌
	for i := 0; i < 34; i++ {
		t := TileFromIndex(i)
		testHand := append(hand.Copy(), t)
		if CanWin(testHand, melds) {
			waiting = append(waiting, t)
		}
	}
	return waiting
}

// IsListening 判断是否听牌
func IsListening(hand Tiles, melds []Meld) bool {
	return len(GetWaitingTiles(hand, melds)) > 0
}

// CanChow 判断是否可以吃牌（只有下家可以吃上家的牌）
// hand: 当前手牌, tile: 要吃的牌
// 返回所有可能的吃牌组合（每组2张手牌）
func CanChow(hand Tiles, tile Tile) []Tiles {
	if tile.IsHonor() {
		return nil // 字牌不能吃
	}

	var results []Tiles
	suit := tile.Suit
	rank := tile.Rank

	// 检查三种顺子组合
	combos := [][2]int{
		{rank - 2, rank - 1}, // tile是第3张
		{rank - 1, rank + 1}, // tile是第2张
		{rank + 1, rank + 2}, // tile是第1张
	}

	for _, combo := range combos {
		r1, r2 := combo[0], combo[1]
		if r1 < 1 || r1 > 9 || r2 < 1 || r2 > 9 {
			continue
		}
		t1 := Tile{Suit: suit, Rank: r1}
		t2 := Tile{Suit: suit, Rank: r2}
		if hand.Contains(t1) && hand.Contains(t2) {
			// 确保不是同一张牌（如果r1==r2）
			if r1 == r2 {
				if hand.Count(t1) >= 2 {
					results = append(results, Tiles{t1, t2})
				}
			} else {
				results = append(results, Tiles{t1, t2})
			}
		}
	}
	return results
}

// CanPong 判断是否可以碰牌
func CanPong(hand Tiles, tile Tile) bool {
	return hand.Count(tile) >= 2
}

// CanKongExposed 判断是否可以明杠（别人打出的牌，手里有3张）
func CanKongExposed(hand Tiles, tile Tile) bool {
	return hand.Count(tile) >= 3
}

// CanKongConcealed 判断手牌中是否有暗杠（4张相同牌）
// 返回可暗杠的牌列表
func CanKongConcealed(hand Tiles) Tiles {
	counts := hand.ToCountArray()
	var result Tiles
	for i, c := range counts {
		if c == 4 {
			result = append(result, TileFromIndex(i))
		}
	}
	return result
}

// CanKongExtended 判断是否可以补杠（已碰的牌，摸到第4张）
// 返回可补杠的牌列表
func CanKongExtended(hand Tiles, melds []Meld) Tiles {
	var result Tiles
	for _, m := range melds {
		if m.Type == MeldPong {
			t := m.Tiles[0]
			if hand.Contains(t) {
				result = append(result, t)
			}
		}
	}
	return result
}
