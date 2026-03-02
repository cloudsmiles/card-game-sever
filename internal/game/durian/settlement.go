package durian

// SettlementResult 结算结果（保存本轮结算详情）
type SettlementResult struct {
	BellRinger        string             `json:"bell_ringer"`          // 摇铃玩家
	AllHolderCards    map[string]Card    `json:"all_holder_cards"`     // 所有玩家牌架卡（公开）
	GorillaEffects    []GorillaEffect    `json:"gorilla_effects"`      // 猩猩牌特效列表
	OrdersBeforeCancel map[FruitType]int `json:"orders_before_cancel"` // 应用猩猩特效前的订单
	OrdersAfterCancel  map[FruitType]int `json:"orders_after_cancel"`  // 应用猩猩特效后的订单
	Inventory          map[FruitType]int `json:"inventory"`            // 库存总量
	IsShortage         bool              `json:"is_shortage"`          // 是否缺货
	ShortageFruits     []FruitType       `json:"shortage_fruits"`      // 缺货的水果种类
	PunishedPlayer     string            `json:"punished_player"`      // 受罚玩家
	PunishReason       string            `json:"punish_reason"`        // 受罚原因（"last_order" 或 "bell_ringer"）
	AngerTokenGiven    int               `json:"anger_token_given"`    // 分配的愤怒标记分值
}

// GorillaEffect 猩猩牌特效记录
type GorillaEffect struct {
	PlayerID        string         `json:"player_id"`        // 持有猩猩牌的玩家
	Ability         GorillaAbility `json:"ability"`          // 特殊能力类型
	CancelledOrders int            `json:"cancelled_orders"` // 被取消的订单数量
}

// CalculateInventory 计算所有玩家牌架卡的水果库存总量
// 猩猩牌不贡献库存
func CalculateInventory(holderCards map[string]Card) map[FruitType]int {
	inventory := map[FruitType]int{
		FruitDurian:     0,
		FruitBanana:     0,
		FruitGrape:      0,
		FruitStrawberry: 0,
	}
	for _, card := range holderCards {
		switch c := card.(type) {
		case *FruitCard:
			inventory[c.LeftFruit] += c.LeftCount
			inventory[c.RightFruit] += c.RightCount
		case *GorillaCard:
			// 猩猩牌不贡献库存
		}
	}
	return inventory
}

// ApplyGorillaEffects 应用猩猩兄妹牌特效，返回调整后的订单区
// 持有猩猩牌的玩家会取消对应水果的所有订单
func ApplyGorillaEffects(orders map[FruitType]int, holderCards map[string]Card) (map[FruitType]int, []GorillaEffect) {
	result := copyOrders(orders)
	effects := make([]GorillaEffect, 0)

	for playerID, card := range holderCards {
		if g, ok := card.(*GorillaCard); ok {
			var cancelledFruit FruitType
			switch g.Ability {
			case AbilityCancelDurian:
				cancelledFruit = FruitDurian
			case AbilityCancelBanana:
				cancelledFruit = FruitBanana
			case AbilityCancelGrape:
				cancelledFruit = FruitGrape
			default:
				continue
			}
			cancelledAmount := result[cancelledFruit]
			result[cancelledFruit] = 0
			effects = append(effects, GorillaEffect{
				PlayerID:        playerID,
				Ability:         g.Ability,
				CancelledOrders: cancelledAmount,
			})
		}
	}
	return result, effects
}

// CheckShortage 判断是否缺货，返回缺货标志和缺货水果列表
func CheckShortage(orders, inventory map[FruitType]int) (bool, []FruitType) {
	shortFruits := make([]FruitType, 0)
	for fruit, demand := range orders {
		if demand > inventory[fruit] {
			shortFruits = append(shortFruits, fruit)
		}
	}
	return len(shortFruits) > 0, shortFruits
}

// DoSettlement 执行完整结算流程，返回结算结果
func DoSettlement(
	bellRinger string,
	lastOrderPlayerIdx int,
	players []string,
	holderCards map[string]Card,
	orders map[FruitType]int,
	tokenPool *AngerTokenPool,
	playerAnger map[string]*PlayerAnger,
) (*SettlementResult, error) {
	result := &SettlementResult{
		BellRinger:         bellRinger,
		AllHolderCards:     holderCards,
		OrdersBeforeCancel: copyOrders(orders),
	}

	// 1. 计算库存总量
	result.Inventory = CalculateInventory(holderCards)

	// 2. 应用猩猩兄妹牌特效
	adjustedOrders, gorillaEffects := ApplyGorillaEffects(orders, holderCards)
	result.GorillaEffects = gorillaEffects
	result.OrdersAfterCancel = adjustedOrders

	// 3. 判断是否缺货
	isShortage, shortFruits := CheckShortage(adjustedOrders, result.Inventory)
	result.IsShortage = isShortage
	result.ShortageFruits = shortFruits

	// 4. 确定受罚玩家
	var punishedPlayer string
	var punishReason string

	if isShortage {
		// 缺货：上一个接订单的玩家受罚
		if lastOrderPlayerIdx >= 0 && lastOrderPlayerIdx < len(players) {
			punishedPlayer = players[lastOrderPlayerIdx]
			punishReason = "last_order"
		} else {
			// 没有人接过订单（此情况理论上在缺货时不会出现，但做防御处理）
			punishedPlayer = bellRinger
			punishReason = "bell_ringer"
		}
	} else {
		// 库存充足（包含订单区为空的情况）：摇铃者受罚
		punishedPlayer = bellRinger
		punishReason = "bell_ringer"
	}

	result.PunishedPlayer = punishedPlayer
	result.PunishReason = punishReason

	// 5. 分配愤怒标记
	if tokenPool.IsEmpty() {
		// 标记池已空（理论上不应发生），直接返回，不分配标记
		result.AngerTokenGiven = 0
		return result, nil
	}

	token, err := tokenPool.TakeSmallest()
	if err != nil {
		return result, err
	}

	result.AngerTokenGiven = token.Value

	if anger, exists := playerAnger[punishedPlayer]; exists {
		anger.AddToken(token)
	}

	return result, nil
}

// copyOrders 复制订单 map
func copyOrders(orders map[FruitType]int) map[FruitType]int {
	result := make(map[FruitType]int, len(orders))
	for k, v := range orders {
		result[k] = v
	}
	return result
}
