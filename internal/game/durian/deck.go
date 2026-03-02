package durian

import "math/rand"

// Deck 牌堆管理
type Deck struct {
	cards       []Card // 当前可用牌堆
	discardPile []Card // 废牌堆（已翻开但非牌架卡的牌）
}

// NewDeck 创建标准牌组（28张水果牌 + 3张猩猩牌，共31张）
func NewDeck() *Deck {
	cards := make([]Card, 0, 31)

	// 28张水果牌（参考PRD 3.2.2节）
	// 注意：每张牌的左右水果必须不同
	fruitCards := []*FruitCard{
		{ID: 1, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitBanana, RightCount: 3},
		{ID: 2, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitBanana, RightCount: 2},
		{ID: 3, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 3, RightFruit: FruitGrape, RightCount: 1},
		{ID: 4, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitGrape, RightCount: 3},
		{ID: 5, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitStrawberry, RightCount: 2},
		{ID: 6, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitStrawberry, RightCount: 4},
		{ID: 7, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 3, RightFruit: FruitGrape, RightCount: 2},
		{ID: 8, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 2, RightFruit: FruitGrape, RightCount: 1},
		{ID: 9, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 1, RightFruit: FruitGrape, RightCount: 4},
		{ID: 10, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 4, RightFruit: FruitStrawberry, RightCount: 1},
		{ID: 11, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 2, RightFruit: FruitStrawberry, RightCount: 3},
		{ID: 12, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 1, RightFruit: FruitStrawberry, RightCount: 2},
		{ID: 13, Type: CardTypeFruit, LeftFruit: FruitGrape, LeftCount: 3, RightFruit: FruitStrawberry, RightCount: 1},
		{ID: 14, Type: CardTypeFruit, LeftFruit: FruitGrape, LeftCount: 2, RightFruit: FruitStrawberry, RightCount: 3},
		{ID: 15, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 4, RightFruit: FruitBanana, RightCount: 1},
		{ID: 16, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 3, RightFruit: FruitGrape, RightCount: 2},
		{ID: 17, Type: CardTypeFruit, LeftFruit: FruitGrape, LeftCount: 2, RightFruit: FruitStrawberry, RightCount: 2},
		{ID: 18, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitStrawberry, RightCount: 4},
		{ID: 19, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitBanana, RightCount: 1},
		{ID: 20, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitStrawberry, RightCount: 1},
		{ID: 21, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 3, RightFruit: FruitStrawberry, RightCount: 2},
		{ID: 22, Type: CardTypeFruit, LeftFruit: FruitGrape, LeftCount: 1, RightFruit: FruitStrawberry, RightCount: 2},
		{ID: 23, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 3, RightFruit: FruitGrape, RightCount: 2},
		{ID: 24, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 2, RightFruit: FruitGrape, RightCount: 3},
		{ID: 25, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 1, RightFruit: FruitBanana, RightCount: 4},
		{ID: 26, Type: CardTypeFruit, LeftFruit: FruitGrape, LeftCount: 4, RightFruit: FruitStrawberry, RightCount: 1},
		{ID: 27, Type: CardTypeFruit, LeftFruit: FruitBanana, LeftCount: 1, RightFruit: FruitGrape, RightCount: 1},
		{ID: 28, Type: CardTypeFruit, LeftFruit: FruitDurian, LeftCount: 2, RightFruit: FruitStrawberry, RightCount: 3},
	}

	for _, fc := range fruitCards {
		cards = append(cards, fc)
	}

	// 3张猩猩兄妹牌
	// 米奇（哥哥）：取消数量为3的水果订单（左或右）
	// 南茜（妹妹）：取消香蕉订单
	// 墨菲（弟弟）：无事发生
	gorillaCards := []*GorillaCard{
		{ID: 101, Type: CardTypeGorilla, Ability: AbilityCancelCount3, Description: "米奇：取消数量为3的水果订单"},
		{ID: 102, Type: CardTypeGorilla, Ability: AbilityCancelBanana, Description: "南茜：取消香蕉订单"},
		{ID: 103, Type: CardTypeGorilla, Ability: AbilityDoNothing, Description: "墨菲：无事发生"},
	}

	for _, gc := range gorillaCards {
		cards = append(cards, gc)
	}

	return &Deck{
		cards:       cards,
		discardPile: make([]Card, 0, 31),
	}
}

// Shuffle 洗牌（Fisher-Yates算法）
func (d *Deck) Shuffle() {
	n := len(d.cards)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
}

// Draw 从牌堆顶部抽取一张牌
// 若牌堆为空，自动将废牌堆重洗后继续
func (d *Deck) Draw() (Card, bool) {
	if len(d.cards) == 0 {
		if len(d.discardPile) == 0 {
			return nil, false
		}
		// 废牌堆重洗入主牌堆
		d.cards = append(d.cards, d.discardPile...)
		d.discardPile = d.discardPile[:0]
		d.Shuffle()
	}

	card := d.cards[0]
	d.cards = d.cards[1:]
	return card, true
}

// Discard 将一张牌放入废牌堆
func (d *Deck) Discard(card Card) {
	d.discardPile = append(d.discardPile, card)
}

// Remaining 返回主牌堆剩余牌数
func (d *Deck) Remaining() int {
	return len(d.cards)
}

// DiscardCount 返回废牌堆牌数
func (d *Deck) DiscardCount() int {
	return len(d.discardPile)
}

// ReshuffleAll 将所有牌（主牌堆+废牌堆）重新洗牌
// 用于新一轮开始时，将上一轮的牌架卡重新混入牌堆
func (d *Deck) ReshuffleAll() {
	// 将所有牌合并到主牌堆
	d.cards = append(d.cards, d.discardPile...)
	d.discardPile = d.discardPile[:0]
	// 重新洗牌
	d.Shuffle()
}
