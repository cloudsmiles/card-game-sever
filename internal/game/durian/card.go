package durian

// FruitType 水果类型
type FruitType string

const (
	FruitDurian     FruitType = "durian"     // 榴莲
	FruitBanana     FruitType = "banana"     // 香蕉
	FruitGrape      FruitType = "grape"      // 葡萄
	FruitStrawberry FruitType = "strawberry" // 草莓
)

// CardType 卡牌类型
type CardType string

const (
	CardTypeFruit   CardType = "fruit"   // 水果牌
	CardTypeGorilla CardType = "gorilla" // 猩猩兄妹牌
)

// GorillaAbility 猩猩兄妹特殊能力类型
type GorillaAbility string

const (
	AbilityCancelBanana GorillaAbility = "cancel_banana" // 南茜：取消香蕉订单
	AbilityCancelCount3 GorillaAbility = "cancel_count3" // 米奇：取消数量为3的水果订单（左或右）
	AbilityDoNothing    GorillaAbility = "do_nothing"    // 墨菲：无事发生
)

// Card 统一卡牌接口
type Card interface {
	GetID() int
	GetCardType() CardType
}

// FruitCard 水果牌（双面水果，每面有1-4个水果）
type FruitCard struct {
	ID         int       `json:"id"`          // 唯一标识（1-28）
	Type       CardType  `json:"card_type"`   // 卡牌类型
	LeftFruit  FruitType `json:"left_fruit"`  // 左侧水果种类
	LeftCount  int       `json:"left_count"`  // 左侧水果数量（1-4）
	RightFruit FruitType `json:"right_fruit"` // 右侧水果种类
	RightCount int       `json:"right_count"` // 右侧水果数量（1-4）
}

func (c *FruitCard) GetID() int            { return c.ID }
func (c *FruitCard) GetCardType() CardType { return CardTypeFruit }

// InventoryContribution 返回该牌贡献的水果库存（左右两面均计入）
func (c *FruitCard) InventoryContribution() map[FruitType]int {
	inv := make(map[FruitType]int)
	inv[c.LeftFruit] += c.LeftCount
	inv[c.RightFruit] += c.RightCount
	return inv
}

// GorillaCard 猩猩兄妹特殊牌
type GorillaCard struct {
	ID          int            `json:"id"`          // 唯一标识（101-103）
	Type        CardType       `json:"card_type"`   // 卡牌类型
	Ability     GorillaAbility `json:"ability"`     // 特殊能力类型
	Description string         `json:"description"` // 能力描述
}

func (c *GorillaCard) GetID() int            { return c.ID }
func (c *GorillaCard) GetCardType() CardType { return CardTypeGorilla }
