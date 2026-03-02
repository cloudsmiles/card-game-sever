# 《榴莲忘返》(Durian) 游戏产品需求文档（PRD）

版本：v1.0  
创建日期：2026-03-02  
产品经理：Card Game PM  
原著：Oink Games（日本）

---

## 1. 产品概述

### 1.1 游戏名称与定位

- **游戏名称**：榴莲忘返（Durian）
- **原版出品**：Oink Games（日本），轻量级桌游
- **游戏定位**：面向在线玩家的 2-7 人推理博弈游戏，核心机制为「信息不对称」——玩家看不到自己的牌，却能看到所有人的牌，通过推理与心理博弈判断店铺库存是否足以满足订单
- **技术架构**：复用现有 card-game-server 的三层架构（Core 核心逻辑层 / Engine 引擎层 / Adapter 适配器层），确保与斗地主、麻将游戏共享房间管理、玩家连接、WebSocket 通信等基础设施

### 1.2 目标用户群体

- **核心用户**：18-35 岁，喜欢轻量派对游戏、逻辑推理的玩家
- **次要用户**：桌游爱好者、喜欢「狼人杀」类信息不对称游戏的用户
- **使用场景**：朋友联机派对、碎片化休闲娱乐、线上快速对局

### 1.3 核心体验关键词

- **信息不对称**：看不到自己的牌，只能靠推理和观察他人
- **快节奏**：单局 5-15 分钟，轻量上手
- **心理博弈**：摇铃的时机是核心决策，考验对公共信息的综合判断
- **易上手难精通**：规则简单，但准确估算库存需要数学推理能力

### 1.4 差异化亮点

1. **独特的「印度扑克」式推理机制**：看不到自己牌的设定创造天然信息不对称，区别于所有常规卡牌游戏
2. **双面水果牌设计**：每张牌两个水果，选哪种水果作为订单本身是策略决策
3. **猩猩兄妹特殊牌**：随机触发特殊规则，增加局面变数
4. **分值递增的愤怒标记**：受罚时被迫取最小分值标记，最终累计最少愤怒分即可获胜，形成独特压力曲线
5. **架构复用**：继承现有服务器架构，开发周期短，维护成本低

---

## 2. 核心机制

### 2.1 游戏人数与角色

- **游戏人数**：2-7 人
- **角色定位**：所有玩家平等，无特殊角色区分
- **主题背景**：玩家扮演丛林水果店的猩猩店员，需确保店铺库存（所有人牌架上的水果）足以满足顾客订单（公共订单区的需求）

### 2.2「看不到自己牌」机制（核心设定）

这是游戏最核心的设计：

```
每个玩家面前放置一个"牌架"（card holder）
↓
每人从牌堆抽取一张水果牌，将其"背对自己"插入牌架
↓
结果：
  - 自己：看不到自己牌架上的水果
  - 其他所有人：能看到你牌架上的水果（包括左右两面的水果）
↓
所有玩家牌架上的水果总量 = 店铺总库存
玩家需要根据"能看到别人的牌、看不到自己的牌"来推理总库存
```

**信息可见性规则摘要**：

| 信息类型 | 自己可见 | 其他玩家可见 |
|---------|---------|------------|
| 自己牌架上的牌 | ❌ 不可见 | ✅ 可见 |
| 他人牌架上的牌 | ✅ 可见 | ✅ 可见 |
| 公共订单区 | ✅ 可见 | ✅ 可见 |
| 牌堆剩余数量 | ✅ 可见 | ✅ 可见 |
| 愤怒标记情况 | ✅ 可见 | ✅ 可见 |

### 2.3 库存计算规则

**总库存 = 所有玩家牌架上水果牌的所有水果数量之和**

> ⚠️ 关键：每张水果牌左右两面各有一种水果，**每面有 1~4 个该水果**，两面水果**都算入库存**。
> 例如：某玩家牌架上是「左：3个香蕉 / 右：2个葡萄」，则该玩家贡献库存：香蕉 3 个 + 葡萄 2 个。

**示例（4人游戏）**：

| 玩家 | 牌架水果牌 | 贡献库存 |
|-----|---------|---------|
| 玩家A | 左：2个榴莲 / 右：3个香蕉 | 榴莲+2, 香蕉+3 |
| 玩家B | 左：1个葡萄 / 右：4个葡萄 | 葡萄+5 |
| 玩家C | 左：3个草莓 / 右：1个榴莲 | 草莓+3, 榴莲+1 |
| 玩家D | 左：2个香蕉 / 右：2个草莓 | 香蕉+2, 草莓+2 |
| **总库存** | | **榴莲:3, 香蕉:5, 葡萄:5, 草莓:5** |

若公共订单区需求为：榴莲2、香蕉3、草莓6，则库存中草莓只有5，缺货（草莓 6 > 5）。

### 2.4 回合流程

每轮游戏中，按顺序轮到玩家行动，每次**二选一**：

#### 行动A：接新订单（Take Order）

1. 从牌堆顶部翻开一张新牌（所有人可见）
2. 该牌有**左右两种水果**，每面各有 **1~4 个**水果
3. 玩家选择其中**一面**的水果作为新订单
4. 将该面的水果种类和数量加入公共订单区（订单区记录每种水果的需求数量，增加该面的数量）
5. 该牌翻开后废弃（不进入任何人的手牌）
6. 行动结束，轮到下一个玩家

> **示例**：翻开的牌为「左：3个香蕉 / 右：2个榴莲」，玩家选择左面 → 订单区香蕉 +3。
>
> **策略点**：选哪面水果作为订单是关键决策。数量大的一面对订单冲击更强，但你需要判断库存是否扛得住。

#### 行动B：摇铃呼叫店长（Ring the Bell）

1. 玩家宣告「我认为订单已经超出库存（缺货）！」
2. 立即触发结算
3. 所有玩家翻开各自牌架上的水果牌，公开库存
4. 按**结算规则**判定谁受罚

### 2.5 结算规则

摇铃后公开所有牌架上的牌，计算总库存与订单区需求：

#### 情况一：订单 > 库存（缺货）

- 某种水果的订单需求数量 > 库存中该种水果的数量
- **判定**：摇铃正确！缺货是真实的
- **受罚**：**上一个接了订单的玩家**获得一个愤怒标记（是他的接单行为导致了缺货）

> 边界情况：若摇铃是第一个行动（没有人接过订单），则规则视为「订单区为空，库存充足」，摇铃者受罚。

#### 情况二：订单 ≤ 库存（货够卖）

- 所有水果种类的订单需求 ≤ 库存中对应水果数量
- **判定**：摇铃错误！库存仍然充足，浪费了店长时间
- **受罚**：**摇铃的玩家**获得一个愤怒标记

#### 结算后处理

1. 清空公共订单区（所有订单清零）
2. 所有玩家从牌架取下旧牌，重新各抽一张新水果牌，背对自己插入牌架
3. 从上一个受罚玩家的**下一位**开始，进入新一轮

---

## 3. 卡牌设计

### 3.1 水果种类

游戏中共 4 种水果：

| 水果 | 英文 | 标识符 |
|-----|------|-------|
| 榴莲 | Durian | `durian` |
| 香蕉 | Banana | `banana` |
| 葡萄 | Grape | `grape` |
| 草莓 | Strawberry | `strawberry` |

### 3.2 水果牌（Fruit Cards）

**总计 28 张水果牌**，每张牌左右两面各有一种水果，且每面的**水果数量为 1~4 个**（两面水果种类可以相同也可以不同）。

> 例如：一张牌可以是「左：3个香蕉 / 右：2个榴莲」，意味着该牌贡献 3 个香蕉 + 2 个榴莲到库存。

#### 3.2.1 数据结构

```go
// internal/game/core/durian/card.go
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

// FruitCard 水果牌（双面水果，每面有1-4个水果）
type FruitCard struct {
    ID          int       `json:"id"`           // 唯一标识（1-28）
    CardType    CardType  `json:"card_type"`    // 卡牌类型
    LeftFruit   FruitType `json:"left_fruit"`   // 左侧水果种类
    LeftCount   int       `json:"left_count"`   // 左侧水果数量（1-4）
    RightFruit  FruitType `json:"right_fruit"`  // 右侧水果种类
    RightCount  int       `json:"right_count"`  // 右侧水果数量（1-4）
}

// InventoryContribution 返回该牌贡献的水果库存
func (c *FruitCard) InventoryContribution() map[FruitType]int {
    inv := make(map[FruitType]int)
    inv[c.LeftFruit] += c.LeftCount
    inv[c.RightFruit] += c.RightCount
    return inv
}

// GorillaCard 猩猩兄妹特殊牌
type GorillaCard struct {
    ID          int      `json:"id"`           // 唯一标识（101-103）
    CardType    CardType `json:"card_type"`    // 卡牌类型
    AbilityType string   `json:"ability_type"` // 特殊能力类型
    Description string   `json:"description"`  // 能力描述
}

// Card 统一卡牌接口
type Card interface {
    GetID() int
    GetCardType() CardType
}

func (c *FruitCard) GetID() int            { return c.ID }
func (c *FruitCard) GetCardType() CardType { return CardTypeFruit }

func (c *GorillaCard) GetID() int            { return c.ID }
func (c *GorillaCard) GetCardType() CardType { return CardTypeGorilla }
```

#### 3.2.2 28 张水果牌完整分布

水果牌设计原则：
- 左右两面水果组合覆盖所有 4 种水果的排列
- 每面水果数量为 **1~4 个**
- 确保每种水果总量相对均衡，支撑多人游戏的推理空间

| 牌ID | 左侧水果 | 左侧数量 | 右侧水果 | 右侧数量 |
|-----|---------|---------|---------|---------|
| 1 | 榴莲 | 2 | 香蕉 | 3 |
| 2 | 榴莲 | 1 | 香蕉 | 2 |
| 3 | 榴莲 | 3 | 葡萄 | 1 |
| 4 | 榴莲 | 1 | 葡萄 | 3 |
| 5 | 榴莲 | 2 | 草莓 | 2 |
| 6 | 榴莲 | 1 | 草莓 | 4 |
| 7 | 香蕉 | 3 | 葡萄 | 2 |
| 8 | 香蕉 | 2 | 葡萄 | 1 |
| 9 | 香蕉 | 1 | 葡萄 | 4 |
| 10 | 香蕉 | 4 | 草莓 | 1 |
| 11 | 香蕉 | 2 | 草莓 | 3 |
| 12 | 香蕉 | 1 | 草莓 | 2 |
| 13 | 葡萄 | 3 | 草莓 | 1 |
| 14 | 葡萄 | 2 | 草莓 | 3 |
| 15 | 榴莲 | 4 | 榴莲 | 1 |
| 16 | 香蕉 | 3 | 香蕉 | 2 |
| 17 | 葡萄 | 2 | 葡萄 | 2 |
| 18 | 草莓 | 1 | 草莓 | 4 |
| 19 | 榴莲 | 2 | 香蕉 | 1 |
| 20 | 榴莲 | 1 | 草莓 | 1 |
| 21 | 香蕉 | 3 | 草莓 | 2 |
| 22 | 葡萄 | 1 | 草莓 | 2 |
| 23 | 榴莲 | 3 | 葡萄 | 2 |
| 24 | 香蕉 | 2 | 葡萄 | 3 |
| 25 | 榴莲 | 1 | 香蕉 | 4 |
| 26 | 葡萄 | 4 | 草莓 | 1 |
| 27 | 香蕉 | 1 | 葡萄 | 1 |
| 28 | 榴莲 | 2 | 草莓 | 3 |

**水果库存总量统计**（28张牌左右两面汇总）：

| 水果 | 总库存量 | 出现在多少张牌上 |
|-----|---------|--------------|
| 榴莲 | 24 | 13 张 |
| 香蕉 | 31 | 14 张 |
| 葡萄 | 28 | 13 张 |
| 草莓 | 29 | 14 张 |
| **合计** | **112** | — |

> 注：以上分布为参考设计值。实际以原版桌游物料为准，如不可获取可按上表实现，后续通过平衡性测试调整。
> 
> **设计说明**：总库存 112 水果分布在 28 张牌上，平均每张牌 4 个水果（左右各约 2 个）。在 4 人游戏中，玩家牌架上平均总库存约 16 个水果，配合订单累积，为推理和虚张声势提供充足的数值空间。

### 3.3 猩猩兄妹牌（Gorilla Cards）

**总计 3 张特殊牌**，与 28 张水果牌混在同一牌堆中（牌堆总计 31 张）。

#### 3.3.1 触发时机

当某玩家抽到这张牌（「接新订单」动作从牌堆翻出），并将其放置到自己牌架上时，立即触发能力。

> 猩猩兄妹牌可以作为玩家的「牌架卡」使用（背对自己放置），此时它的特殊能力在该回合结束时激活。

#### 3.3.2 三张特殊牌说明

| 牌ID | 名称 | 能力描述 | 机制说明 |
|-----|-----|---------|---------|
| 101 | 猩猩大哥 | 取消本轮中「榴莲」类型的所有订单 | 结算时，订单区中所有榴莲订单清零，不计入缺货判断 |
| 102 | 猩猩二姐 | 取消本轮中「香蕉」类型的所有订单 | 结算时，订单区中所有香蕉订单清零，不计入缺货判断 |
| 103 | 猩猩三妹 | 取消本轮中「葡萄」类型的所有订单 | 结算时，订单区中所有葡萄订单清零，不计入缺货判断 |

> **注**：猩猩兄妹牌在牌架上时，不向左右两个水果贡献库存（该牌没有水果面，仅有特殊能力）。持有猩猩兄妹牌的玩家，贡献的库存为 0。

#### 3.3.3 特殊牌数据结构扩展

```go
// GorillaAbility 猩猩兄妹特殊能力类型
type GorillaAbility string

const (
    AbilityCancelDurian     GorillaAbility = "cancel_durian"     // 取消榴莲订单
    AbilityCancelBanana     GorillaAbility = "cancel_banana"     // 取消香蕉订单
    AbilityCancelGrape      GorillaAbility = "cancel_grape"      // 取消葡萄订单
)

// GorillaCard（完整定义）
type GorillaCard struct {
    ID      int            `json:"id"`
    CardType CardType      `json:"card_type"`
    Ability GorillaAbility `json:"ability"`
    Description string     `json:"description"`
}

// HasFruits 猩猩牌不贡献水果库存
func (c *GorillaCard) HasFruits() bool { return false }
```

### 3.4 牌堆构成总结

| 类型 | 数量 | 说明 |
|-----|-----|-----|
| 水果牌 | 28 张 | 每张左右各一种水果，双面均计入库存 |
| 猩猩兄妹牌 | 3 张 | 特殊能力牌，触发时取消某类订单 |
| **合计** | **31 张** | 游戏开始时洗牌，玩家每轮各抽 1 张 |

---

## 4. 愤怒标记系统

### 4.1 标记分值设计

愤怒标记（Anger Token）共 7 枚，分值分别为：

| 标记编号 | 分值 | 备注 |
|---------|-----|------|
| 标记①  | 1 分 | 最低分值，首次受罚取此 |
| 标记②  | 2 分 | |
| 标记③  | 3 分 | |
| 标记④  | 4 分 | |
| 标记⑤  | 5 分 | |
| 标记⑥  | 6 分 | |
| 标记⑦  | 7 分 | 最高单枚标记 |

### 4.2 取标记规则

- 受罚时，从**公共标记池**中取出**当前剩余分值最小**的那枚标记
- 标记一旦被取走，不归还
- 一次结算只取 **1 枚**标记

**示例**：

```
当前标记池剩余：①②③④⑤⑥⑦（全部7枚）
→ 第1次受罚：取①（1分），标记池剩余：②③④⑤⑥⑦
→ 第2次受罚（同人或他人）：取②（2分），标记池剩余：③④⑤⑥⑦
→ 第3次受罚：取③（3分），标记池剩余：④⑤⑥⑦
```

### 4.3 游戏结束条件

当任意玩家的**愤怒标记总分 ≥ 7 分**时，游戏立即结束。

> 注意：「总分」是该玩家持有的所有愤怒标记的分值之和，不是枚数。
>
> **示例**：玩家A持有①②③③，总分 = 1+2+3+3 = 9 ≥ 7 → 游戏结束。

### 4.4 胜负判定

游戏结束时：
- **愤怒标记总分最少**的玩家获胜
- **并列最少**时，并列玩家共同获胜

### 4.5 标记系统数据结构

```go
// internal/game/core/durian/anger_token.go
package durian

// AngerToken 愤怒标记
type AngerToken struct {
    Value int `json:"value"` // 分值（1-7）
}

// AngerTokenPool 公共标记池
type AngerTokenPool struct {
    Remaining []*AngerToken `json:"remaining"` // 剩余标记（按分值升序）
}

// TakeSmallest 取出最小分值标记
func (p *AngerTokenPool) TakeSmallest() (*AngerToken, error) {
    if len(p.Remaining) == 0 {
        return nil, fmt.Errorf("标记池已空")
    }
    // 已按升序排列，取第一枚
    token := p.Remaining[0]
    p.Remaining = p.Remaining[1:]
    return token, nil
}

// IsEmpty 标记池是否为空
func (p *AngerTokenPool) IsEmpty() bool {
    return len(p.Remaining) == 0
}

// NewAngerTokenPool 初始化标记池（7枚，分值1-7）
func NewAngerTokenPool() *AngerTokenPool {
    tokens := make([]*AngerToken, 0, 7)
    for i := 1; i <= 7; i++ {
        tokens = append(tokens, &AngerToken{Value: i})
    }
    return &AngerTokenPool{Remaining: tokens}
}

// PlayerAnger 玩家持有的愤怒标记
type PlayerAnger struct {
    PlayerID string        `json:"player_id"`
    Tokens   []*AngerToken `json:"tokens"` // 持有的标记列表
}

// TotalScore 计算愤怒总分
func (pa *PlayerAnger) TotalScore() int {
    total := 0
    for _, t := range pa.Tokens {
        total += t.Value
    }
    return total
}

// IsEliminated 是否达到淘汰条件（总分>=7）
func (pa *PlayerAnger) IsEliminated() bool {
    return pa.TotalScore() >= 7
}
```

---

## 5. 平衡性设计

### 5.1 各玩家胜率预期

| 人数 | 理论均等胜率 | 说明 |
|-----|-----------|------|
| 2人 | 50% / 人 | 完全对称 |
| 3人 | 33.3% / 人 | — |
| 4人 | 25% / 人 | — |
| 5人 | 20% / 人 | — |
| 6人 | 16.7% / 人 | — |
| 7人 | 14.3% / 人 | — |

> **注意**：先手玩家拥有轻微优势（可观察更多他人的牌后做决策），但游戏轮次多，长期影响可忽略。

### 5.2 运气与技巧占比

- **运气占比**：35%（牌堆顺序、抽到哪张牌）
- **技巧占比**：65%（推理总库存、摇铃时机判断、订单选择策略）

**技巧体现维度**：

| 技巧维度 | 说明 |
|---------|------|
| 库存推理 | 看到所有他人的牌，结合自己的历史信息，推理自己牌架上可能是哪类水果 |
| 订单策略 | 接订单时选哪种水果，引导局面对自己有利 |
| 摇铃时机 | 判断「当前订单是否已超出库存」，太早（浪费机会）或太晚（被他人摇铃）都会受罚 |
| 读心术 | 观察其他玩家的表情/犹豫程度，判断他们对库存的判断 |

### 5.3 猩猩兄妹牌的平衡影响

- 猩猩兄妹牌能力使某类订单归零，可能逆转局面（本以为缺货 → 特殊牌取消后库存够了）
- **正向**：增加博弈层次，避免游戏变成纯数学计算
- **风险**：可能使接近结束的局面被强行延长

**平衡措施**：
- 猩猩兄妹牌的能力在结算时才生效，摇铃前所有人均不知道对方是否持有特殊牌（背对自己放置）
- 若猩猩兄妹牌的存在导致频繁「翻盘」，可在迭代版本中调整其出现概率（减少为 2 张或降低特殊能力效果）

### 5.4 反制机制设计

| 策略 | 反制手段 |
|-----|---------|
| 频繁接订单拉高需求 | 其他人观察到订单过多，会提早摇铃；接单者若是导致缺货的人，会受罚 |
| 迟迟不摇铃等待他人出错 | 有人会在缺货临界点接一个订单，触发缺货，导致接单者受罚 |
| 过早摇铃 | 若库存仍充裕，摇铃者受罚 |
| 依赖猩猩牌救场 | 猩猩牌是随机的，不能保证持有目标水果的取消能力 |

### 5.5 数值调节方案

- **愤怒标记阈值**（当前：7分）：若游戏过短可调低，过长可调高
- **标记分值曲线**（当前：1-7 线性）：可改为 1-2-3-5-8 等非线性，增加后期压力感
- **特殊牌数量**（当前：3张）：调整为 2 张可降低随机性，调整为 4 张可增加变数

---

## 6. 游戏流程与状态机

### 6.1 游戏状态定义

```go
// internal/game/core/durian/state.go
package durian

// GamePhase 游戏阶段
type GamePhase string

const (
    PhaseWaiting    GamePhase = "waiting"     // 等待玩家加入
    PhaseReady      GamePhase = "ready"       // 玩家就绪，准备开始
    PhaseDealing    GamePhase = "dealing"     // 发牌阶段（每轮开始，玩家抽取牌架卡）
    PhasePlaying    GamePhase = "playing"     // 游戏进行中（等待当前玩家行动）
    PhaseSettlement GamePhase = "settlement"  // 结算阶段（摇铃后公开牌架）
    PhaseRoundEnd   GamePhase = "round_end"   // 单轮结束（展示结果、分配愤怒标记）
    PhaseGameOver   GamePhase = "game_over"   // 游戏结束（有玩家愤怒分>=7）
)
```

### 6.2 状态流转图

```
PhaseWaiting ──【玩家就绪(2-7人)】──> PhaseReady
                    │
                    ▼
               PhaseDealing ──【所有人抽牌插入牌架】──> PhasePlaying
                    ▲                                       │
                    │                               ┌───────┴────────┐
                    │                               │  当前玩家行动   │
                    │                               │  ① 接新订单    │
                    │                               │  ② 摇铃       │
                    │                               └───────┬────────┘
                    │                                       │
                    │                         ┌─────────────┴─────────────┐
                    │                         │                           │
                    │                    【接订单】                   【摇铃】
                    │                  更新订单区                         │
                    │                  轮到下一人                         ▼
                    │                  (回到PhasePlaying)        PhaseSettlement
                    │                                              （公开所有牌架）
                    │                                                     │
                    │                                              判断缺货/不缺货
                    │                                                     │
                    │                                            PhaseRoundEnd
                    │                                         （分配愤怒标记）
                    │                                                     │
                    │                               ┌─────────────────────┤
                    │                               │                     │
                    │                       【无人愤怒分>=7】      【有人愤怒分>=7】
                    └───────────────────────────────┘                     │
                          （重置订单区，重新发牌）                           ▼
                                                                    PhaseGameOver
```

### 6.3 关键流程详细说明

#### 6.3.1 开局流程

```
1. 房间创建 → PhaseWaiting
2. 玩家加入（2-7人），所有人准备 → PhaseReady
3. 倒计时 3 秒 → PhaseDealing
4. 执行发牌：
   - 洗牌（Fisher-Yates 算法，31张牌混洗）
   - 每位玩家从牌堆顶部各取 1 张，背对自己插入牌架
   - 服务端记录每位玩家的牌架卡（不发给该玩家，只发给其他玩家）
5. 随机决定首位行动玩家（或按加入顺序）
6. 进入 PhasePlaying，等待当前玩家行动
```

#### 6.3.2 单轮行动流程

```
A. 当前玩家行动阶段（PhasePlaying）
   - 服务端通知当前玩家可执行的操作：接订单 or 摇铃
   - 超时（30秒）未操作 → 自动执行「接新订单」（取牌堆顶，随机选一种水果）

B. 行动A：接新订单
   - 翻开牌堆顶部一张牌（广播给所有人）
   - 当前玩家选择左侧或右侧水果作为订单
   - 订单区对应水果数量 +1
   - 翻开的牌废弃
   - 轮到下一位玩家（AdvanceTurn）
   - 回到 PhasePlaying

C. 行动B：摇铃
   - 进入 PhaseSettlement
   - 所有玩家牌架上的牌公开
   - 计算总库存（汇总所有水果牌的左右水果）
   - 应用猩猩兄妹牌特殊能力（若有人持有，取消对应水果订单）
   - 对比订单区 vs 总库存
   - 进入 PhaseRoundEnd，分配愤怒标记
```

#### 6.3.3 结算流程

```
1. 公开所有牌架卡
2. 计算各水果库存总量：
   for each player:
     if player.HolderCard is FruitCard:
       inventory[LeftFruit]  += LeftCount    // 左侧水果数量（1-4）
       inventory[RightFruit] += RightCount   // 右侧水果数量（1-4）
     else if player.HolderCard is GorillaCard:
       inventory 不变（猩猩牌不贡献水果）
       激活特殊能力（取消对应水果的订单）

3. 对比订单区 vs 实际库存：
   isShortage = false
   for each fruitType in orders:
     if orders[fruitType] > inventory[fruitType]:
       isShortage = true
       break

4. 分配愤怒标记：
   if isShortage:
     punished = lastOrderPlayer  // 上一个接订单的玩家
   else:
     punished = bellRingerPlayer // 摇铃的玩家

5. punished 玩家从标记池取最小分值标记

6. 检查游戏是否结束：
   if punished.TotalAngerScore() >= 7:
     → PhaseGameOver
   else:
     → 清空订单区，重新发牌，进入下一轮
     → 从 punished 玩家的下一位开始行动
```

### 6.4 异常处理路径

| 异常场景 | 处理逻辑 |
|---------|---------|
| 玩家断线（游戏中）| 该玩家变为「托管」状态，超时自动接订单（随机选择水果），3分钟内可重连恢复控制 |
| 玩家断线（等待行动阶段）| 60秒未重连，踢出房间；若剩余人数 < 2 人，游戏终止 |
| 操作超时（30秒）| 自动执行接订单动作（从牌堆翻牌，随机选择左/右水果） |
| 牌堆耗尽（无牌可翻）| 重新洗入所有已废弃的牌（不含当前牌架上的牌），继续游戏 |
| 标记池耗尽 | 理论上不应出现（7枚标记对应7分阈值，至多被取完时所有标记都分配出去游戏就结束了）；若出现，以当前持有最少分的玩家获胜，游戏结束 |
| 猩猩牌取消订单后订单区清零 | 此时订单为0，等价于「库存够」，摇铃者受罚 |
| 摇铃是第一个行动（订单区为空）| 视为「库存充足」，摇铃者受罚 |
| 非法操作（选择牌上不存在的水果面）| 拒绝操作，返回错误码 4001，要求重新操作 |

---

## 7. WebSocket 协议设计

### 7.1 协议格式（与现有项目一致）

```json
{
  "type": "消息类型",
  "data": {
    // 具体数据
  },
  "timestamp": 1709107200000
}
```

### 7.2 客户端 → 服务器消息

#### 7.2.1 加入房间

```json
{
  "type": "join_room",
  "data": {
    "room_id": "room_durian_001",
    "player_id": "player_67890",
    "token": "auth_token_here"
  }
}
```

#### 7.2.2 准备游戏

```json
{
  "type": "ready",
  "data": {
    "player_id": "player_67890"
  }
}
```

#### 7.2.3 行动A：接新订单（选择水果面）

```json
{
  "type": "take_order",
  "data": {
    "player_id": "player_67890",
    "chosen_side": "left"
  }
}
```

> `chosen_side` 必须为 `"left"` 或 `"right"`，表示选择当前翻开牌的哪一面水果作为订单。服务端会自动将该面的水果种类和数量加入订单区。

#### 7.2.4 行动B：摇铃

```json
{
  "type": "ring_bell",
  "data": {
    "player_id": "player_67890"
  }
}
```

### 7.3 服务器 → 客户端消息

#### 7.3.1 游戏状态广播

```json
{
  "type": "game_state",
  "data": {
    "room_id": "room_durian_001",
    "phase": "playing",
    "current_turn_player": "player_67890",
    "round_number": 3,
    "deck_remaining": 18,
    "orders": {
      "durian": 2,
      "banana": 1,
      "grape": 0,
      "strawberry": 3
    },
    "anger_token_pool": [3, 4, 5, 6, 7],
    "players": [
      {
        "player_id": "player_67890",
        "nickname": "阿明",
        "anger_score": 3,
        "anger_tokens": [1, 2],
        "is_online": true
      },
      {
        "player_id": "player_11111",
        "nickname": "小花",
        "anger_score": 1,
        "anger_tokens": [1],
        "is_online": true
      }
    ]
  }
}
```

#### 7.3.2 发牌通知（每轮开始）

> **⚠️ 关键安全设计**：此消息为个人专属消息，服务器**分别发送给每个玩家**。玩家A的 `your_card_visible` 为 `false`（不透露自己的牌），`others_cards` 包含其他所有人的牌架信息。

**发给玩家A（player_67890）的消息**：

```json
{
  "type": "round_deal",
  "data": {
    "round_number": 4,
    "your_card": null,
    "your_card_visible": false,
    "others_cards": [
      {
        "player_id": "player_11111",
        "nickname": "小花",
        "card": {
          "id": 12,
          "card_type": "fruit",
          "left_fruit": "banana",
          "left_count": 1,
          "right_fruit": "strawberry",
          "right_count": 2
        }
      },
      {
        "player_id": "player_22222",
        "nickname": "老王",
        "card": {
          "id": 101,
          "card_type": "gorilla",
          "ability": "cancel_durian",
          "description": "取消本轮所有榴莲订单"
        }
      }
    ]
  }
}
```

**发给玩家B（player_11111）的消息（包含对player_11111自己隐藏、其他人可见的信息）**：

```json
{
  "type": "round_deal",
  "data": {
    "round_number": 4,
    "your_card": null,
    "your_card_visible": false,
    "others_cards": [
      {
        "player_id": "player_67890",
        "nickname": "阿明",
        "card": {
          "id": 8,
          "card_type": "fruit",
          "left_fruit": "durian",
          "left_count": 2,
          "right_fruit": "banana",
          "right_count": 3
        }
      },
      {
        "player_id": "player_22222",
        "nickname": "老王",
        "card": {
          "id": 101,
          "card_type": "gorilla",
          "ability": "cancel_durian",
          "description": "取消本轮所有榴莲订单"
        }
      }
    ]
  }
}
```

> 规则总结：**服务端绝对不将玩家自己的牌架卡信息发给该玩家。** `your_card` 字段始终为 `null`，由客户端根据 `your_card_visible: false` 渲染「背面朝向自己的牌架」动效。

#### 7.3.3 行动请求通知（轮到某玩家行动）

```json
{
  "type": "action_request",
  "data": {
    "player_id": "player_67890",
    "available_actions": ["take_order", "ring_bell"],
    "timeout_ms": 30000
  }
}
```

#### 7.3.4 翻开订单牌广播（翻牌时立即广播给所有人）

> 行动玩家执行「接新订单」时，先广播翻开的牌内容，再等待玩家选择哪面水果。

```json
{
  "type": "order_card_flipped",
  "data": {
    "player_id": "player_67890",
    "flipped_card": {
      "id": 15,
      "card_type": "fruit",
      "left_fruit": "banana",
      "left_count": 3,
      "right_fruit": "grape",
      "right_count": 2
    },
    "deck_remaining": 17
  }
}
```

#### 7.3.5 订单更新广播（玩家选择水果面后）

```json
{
  "type": "order_updated",
  "data": {
    "player_id": "player_67890",
    "chosen_side": "left",
    "chosen_fruit": "banana",
    "chosen_count": 3,
    "orders": {
      "durian": 2,
      "banana": 5,
      "grape": 0,
      "strawberry": 3
    },
    "next_turn_player": "player_11111"
  }
}
```

#### 7.3.6 结算广播（摇铃后）

> 此消息发给所有人，完整公开所有牌架信息（包括每个玩家自己的牌架卡）。

```json
{
  "type": "settlement",
  "data": {
    "bell_ringer": "player_67890",
    "all_holder_cards": [
      {
        "player_id": "player_67890",
        "card": {
          "id": 8,
          "card_type": "fruit",
          "left_fruit": "durian",
          "left_count": 2,
          "right_fruit": "banana",
          "right_count": 3
        }
      },
      {
        "player_id": "player_11111",
        "card": {
          "id": 12,
          "card_type": "fruit",
          "left_fruit": "banana",
          "left_count": 1,
          "right_fruit": "grape",
          "right_count": 2
        }
      },
      {
        "player_id": "player_22222",
        "card": {
          "id": 101,
          "card_type": "gorilla",
          "ability": "cancel_durian",
          "description": "取消本轮所有榴莲订单"
        }
      }
    ],
    "gorilla_effects": [
      {
        "player_id": "player_22222",
        "ability": "cancel_durian",
        "cancelled_orders": 2
      }
    ],
    "orders_before_cancel": {
      "durian": 2,
      "banana": 2,
      "grape": 0,
      "strawberry": 3
    },
    "orders_after_cancel": {
      "durian": 0,
      "banana": 2,
      "grape": 0,
      "strawberry": 3
    },
    "inventory": {
      "durian": 1,
      "banana": 3,
      "grape": 1,
      "strawberry": 0
    },
    "is_shortage": true,
    "shortage_fruits": ["strawberry"],
    "punished_player": "player_11111",
    "punish_reason": "last_order",
    "anger_token_given": 3
  }
}
```

#### 7.3.7 轮次结束广播

```json
{
  "type": "round_end",
  "data": {
    "round_number": 4,
    "punished_player": "player_11111",
    "anger_token_given": 3,
    "player_anger_scores": [
      {"player_id": "player_67890", "score": 3, "tokens": [1, 2]},
      {"player_id": "player_11111", "score": 6, "tokens": [1, 5]},
      {"player_id": "player_22222", "score": 0, "tokens": []}
    ],
    "next_first_player": "player_22222",
    "game_over": false
  }
}
```

#### 7.3.8 游戏结束广播

```json
{
  "type": "game_over",
  "data": {
    "trigger_player": "player_11111",
    "final_scores": [
      {"player_id": "player_67890", "score": 3, "rank": 2},
      {"player_id": "player_11111", "score": 9, "rank": 3},
      {"player_id": "player_22222", "score": 0, "rank": 1}
    ],
    "winners": ["player_22222"],
    "is_tie": false
  }
}
```

#### 7.3.9 错误通知

```json
{
  "type": "error",
  "data": {
    "code": 4001,
    "message": "非法操作：chosen_side 必须为 left 或 right",
    "details": {
      "chosen_side": "top",
      "available_sides": ["left", "right"]
    }
  }
}
```

### 7.4 错误码定义

| 错误码 | 说明 |
|-------|------|
| 4001 | 非法操作（所选面不是 left/right、不在行动回合等）|
| 4002 | 操作超时 |
| 4003 | 不在当前玩家回合 |
| 4004 | 游戏状态错误（非 playing 阶段）|
| 4005 | 权限不足（未认证）|
| 4006 | 牌堆已空（极端情况，自动重洗后重试）|
| 5001 | 服务器内部错误 |

---

## 8. 架构适配说明

### 8.1 实现 Game 接口

榴莲忘返游戏需实现 `internal/game/interfaces/game.go` 中定义的 `Game` 接口：

```go
// internal/game/core/durian/durian_game.go
package durian

import "card-game-server/internal/game/interfaces"

// DurianGame 实现 Game 接口
type DurianGame struct {
    id          string
    players     []string            // 玩家ID列表（顺序即行动顺序）
    phase       GamePhase
    deck        []Card              // 牌堆（洗牌后）
    discardPile []Card              // 废牌堆
    holderCards map[string]Card     // 每位玩家的牌架卡（playerID -> Card）
    orders      map[FruitType]int   // 当前订单区（每种水果的需求数量）
    tokenPool   *AngerTokenPool     // 公共愤怒标记池
    playerAnger map[string]*PlayerAnger // 每位玩家的愤怒标记
    currentTurn int                 // 当前行动玩家在players中的索引
    lastOrderPlayerIdx int          // 上一个接订单的玩家索引（-1表示无人接过订单）
    roundNumber int                 // 当前轮次编号
}

// 确保实现接口
var _ interfaces.Game = (*DurianGame)(nil)

func (g *DurianGame) ID() string { return g.id }

func (g *DurianGame) MinPlayers() int { return 2 }
func (g *DurianGame) MaxPlayers() int { return 7 }

func (g *DurianGame) Init(players []string) error {
    if len(players) < g.MinPlayers() || len(players) > g.MaxPlayers() {
        return fmt.Errorf("玩家数量不符合要求：需要 2-7 人，当前 %d 人", len(players))
    }
    g.players = players
    g.phase = PhaseDealing
    g.tokenPool = NewAngerTokenPool()
    g.playerAnger = make(map[string]*PlayerAnger)
    for _, pid := range players {
        g.playerAnger[pid] = &PlayerAnger{PlayerID: pid}
    }
    g.orders = make(map[FruitType]int)
    g.lastOrderPlayerIdx = -1
    g.roundNumber = 0
    return g.startNewRound()
}

func (g *DurianGame) CurrentTurn() string {
    return g.players[g.currentTurn]
}

func (g *DurianGame) AdvanceTurn() {
    g.currentTurn = (g.currentTurn + 1) % len(g.players)
}

// ProcessAction 处理玩家行动
// action 的实际类型为 *DurianAction
func (g *DurianGame) ProcessAction(playerID string, action interface{}) (bool, error) {
    if g.CurrentTurn() != playerID {
        return false, fmt.Errorf("错误：不在 %s 的回合", playerID)
    }
    a, ok := action.(*DurianAction)
    if !ok {
        return false, fmt.Errorf("错误：无效的行动类型")
    }
    switch a.Type {
    case ActionTakeOrder:
        return g.processTakeOrder(playerID, a)
    case ActionRingBell:
        return g.processRingBell(playerID)
    default:
        return false, fmt.Errorf("错误：未知行动类型 %s", a.Type)
    }
}

// GetState 返回公共游戏状态（不含任何玩家的私密牌架信息）
func (g *DurianGame) GetState() interface{} {
    return &DurianPublicState{
        Phase:              g.phase,
        RoundNumber:        g.roundNumber,
        CurrentTurnPlayer:  g.CurrentTurn(),
        DeckRemaining:      len(g.deck),
        Orders:             g.orders,
        AngerTokenPool:     g.tokenPool.Remaining,
        PlayerAngerScores:  g.getAngerSummary(),
    }
}

// GetStateForPlayer 返回指定玩家视角的游戏状态
// ⚠️ 关键：该玩家自己的牌架卡不包含在返回数据中！
func (g *DurianGame) GetStateForPlayer(playerID string) interface{} {
    othersCards := make(map[string]Card)
    for pid, card := range g.holderCards {
        if pid != playerID { // 只包含其他玩家的牌架卡，绝不包含自己的
            othersCards[pid] = card
        }
    }
    return &DurianPlayerState{
        PublicState:    g.GetState().(*DurianPublicState),
        YourCard:       nil,   // 永远为 nil，自己看不到自己的牌
        OthersCards:    othersCards,
    }
}

func (g *DurianGame) IsGameOver() bool {
    return g.phase == PhaseGameOver
}

func (g *DurianGame) Winner() string {
    if !g.IsGameOver() {
        return ""
    }
    minScore := -1
    winners := []string{}
    for pid, anger := range g.playerAnger {
        score := anger.TotalScore()
        if minScore == -1 || score < minScore {
            minScore = score
            winners = []string{pid}
        } else if score == minScore {
            winners = append(winners, pid)
        }
    }
    if len(winners) == 1 {
        return winners[0]
    }
    // 并列时返回逗号拼接（适配现有接口，可返回第一名）
    return strings.Join(winners, ",")
}
```

### 8.2 行动数据结构

```go
// internal/game/core/durian/action.go
package durian

// ActionType 行动类型
type ActionType string

const (
    ActionTakeOrder ActionType = "take_order" // 接新订单
    ActionRingBell  ActionType = "ring_bell"  // 摇铃呼叫店长
)

// DurianAction 榴莲忘返行动
type DurianAction struct {
    Type       ActionType `json:"type"`
    ChosenSide string     `json:"chosen_side,omitempty"` // 仅 take_order 时有效："left" 或 "right"
}
```

### 8.3 复用现有架构

#### 8.3.1 Core 层（核心逻辑）

路径：`internal/game/core/durian/`

| 文件 | 职责 |
|-----|------|
| `durian_game.go` | 游戏主逻辑，实现 Game 接口 |
| `card.go` | 卡牌数据结构（FruitCard、GorillaCard）|
| `deck.go` | 牌堆管理（洗牌、发牌、废牌回收）|
| `anger_token.go` | 愤怒标记系统 |
| `state.go` | 状态枚举与状态数据结构 |
| `action.go` | 行动类型与数据结构 |
| `settlement.go` | 结算逻辑（库存计算、缺货判断、猩猩特效应用）|

#### 8.3.2 Engine 层

路径：`internal/game/engine/`

复用现有 `game_engine.go`，扩展游戏类型注册：

```go
// 注册榴莲忘返游戏工厂
engine.RegisterGameFactory("durian", func() interfaces.Game {
    return durian.NewDurianGame()
})
```

#### 8.3.3 Adapter 层

路径：`internal/game/adapter/`

复用现有 WebSocket Adapter，扩展消息类型路由：

```go
// internal/game/adapter/durian_handler.go
package adapter

func (a *WebSocketAdapter) handleDurianMessage(msg *Message, player *Player) {
    switch msg.Type {
    case "take_order":
        a.handleTakeOrder(msg, player)
    case "ring_bell":
        a.handleRingBell(msg, player)
    case "ready":
        a.handleReady(msg, player)
    }
}

func (a *WebSocketAdapter) handleTakeOrder(msg *Message, player *Player) {
    var data struct {
        ChosenSide string `json:"chosen_side"` // "left" 或 "right"
    }
    if err := json.Unmarshal(msg.Data, &data); err != nil {
        a.sendError(player, 4001, "请求格式错误")
        return
    }

    game := a.getGame(player.RoomID)
    action := &durian.DurianAction{
        Type:       durian.ActionTakeOrder,
        ChosenSide: data.ChosenSide,
    }
    _, err := game.ProcessAction(player.ID, action)
    if err != nil {
        a.sendError(player, 4001, err.Error())
        return
    }

    // 广播订单更新给所有人
    a.broadcastOrderUpdate(player.RoomID, game)
}

func (a *WebSocketAdapter) handleRingBell(msg *Message, player *Player) {
    game := a.getGame(player.RoomID)
    action := &durian.DurianAction{Type: durian.ActionRingBell}
    _, err := game.ProcessAction(player.ID, action)
    if err != nil {
        a.sendError(player, 4001, err.Error())
        return
    }

    // 结算后，广播结算消息（此时包含所有人的牌架卡，全部公开）
    a.broadcastSettlement(player.RoomID, game)
}
```

### 8.4 共享基础设施

| 模块 | 路径 | 复用说明 |
|-----|------|---------|
| 房间管理 | `internal/room/manager.go` | 直接复用，支持 durian 房间类型 |
| 玩家管理 | `internal/player/manager.go` | 复用玩家连接、认证逻辑 |
| 消息路由 | `internal/server/router.go` | 扩展路由规则，增加 durian 消息处理器 |
| 状态持久化 | `internal/storage/redis.go` | 复用 Redis 存储，持久化游戏状态 |
| 日志系统 | `pkg/logger/` | 直接复用 |
| 心跳检测 | Adapter 层 | 复用现有心跳与断线检测 |

### 8.5 数据库扩展

```sql
-- 榴莲忘返游戏记录表
CREATE TABLE durian_game_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id VARCHAR(32) NOT NULL,
    round_num INT NOT NULL,
    bell_ringer_id VARCHAR(32) NOT NULL,
    is_shortage BOOLEAN NOT NULL,
    punished_player_id VARCHAR(32) NOT NULL,
    anger_token_value INT NOT NULL,
    orders_snapshot JSON,    -- 结算时订单区快照
    inventory_snapshot JSON, -- 结算时库存快照
    holder_cards_snapshot JSON, -- 所有玩家牌架卡快照
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_punished (punished_player_id)
);

-- 玩家统计表扩展
ALTER TABLE player_stats ADD COLUMN durian_games INT DEFAULT 0;
ALTER TABLE player_stats ADD COLUMN durian_wins INT DEFAULT 0;
ALTER TABLE player_stats ADD COLUMN durian_avg_anger_score FLOAT DEFAULT 0;
```

---

## 9. 风险与应对

### 9.1 信息安全风险

| 风险 | 说明 | 防范措施 |
|-----|-----|---------|
| 客户端作弊（获取自己的牌）| 黑客修改客户端拦截他人消息，推断自己的牌 | 服务端严格执行 `GetStateForPlayer` 信息隔离，`your_card` 字段永远为 null |
| 抓包查看他人消息 | 玩家抓包分析 WebSocket 流量 | 全程 WSS（TLS加密），避免明文传输 |
| 作弊工具辅助计算 | 外挂自动计算库存总量 | 接受（库存计算是游戏核心乐趣，属于技巧范畴，无需防御）|
| 服务端日志泄露 | 日志记录了玩家的牌架卡内容 | 游戏日志脱敏，不记录牌面明文到可访问日志中 |

### 9.2 极端情况处理

| 情况 | 处理方案 |
|-----|---------|
| 猩猩牌取消所有订单，订单归零后摇铃 | 订单为0视为「不缺货」，摇铃者受罚 |
| 所有7张愤怒标记分配完毕 | 此时必有玩家总分≥7，游戏应已结束；若出现边界bug，以当前最低分玩家获胜 |
| 多人同时满足愤怒分≥7 | 以最后触发的那次结算为准，当局结算完后判定，最低愤怒分的玩家获胜（并列则共同胜利）|
| 仅剩2人，一人断线 | 断线玩家进入托管，自动接订单（随机选择），若60秒仍无法重连，游戏终止，在线玩家胜出 |
| 牌堆耗尽无牌可翻 | 将废牌堆重新洗入牌堆（不含当前牌架上的牌），继续游戏 |
| 猩猩牌在牌堆耗尽后被翻出 | 猩猩牌也可以被翻出作为「订单翻牌」，但此时无水果可选，视为「特殊事件」：该玩家必须摇铃（强制结算）|

### 9.3 玩家体验保障

1. **新手引导**：
   - 首局开启「教学模式」，显示库存计算提示（可见到自己的牌，仅教学用途）
   - 引导玩家理解「看不到自己的牌」的规则
   
2. **信息可视化**：
   - 订单区水果数量大字体显示，支持颜色区分不同水果
   - 他人牌架卡正面展示（水果图案清晰可辨），自己牌架显示背面
   
3. **操作提示**：
   - 轮到自己行动时，高亮「接订单」和「摇铃」按钮
   - 剩余操作时间倒计时显示（30秒）
   
4. **结算动效**：
   - 摇铃后，所有牌架翻转动画（背面 → 正面）
   - 猩猩牌触发时，对应水果订单格动态消除动效
   - 愤怒标记分配动画

5. **回放系统**：
   - 每局保存完整对局记录（各轮牌架卡、订单选择、摇铃时机）
   - 玩家可复盘查看（包括自己当时未知的牌）

---

## 10. 验收标准

### 10.1 功能验收

| 功能模块 | 验收标准 |
|---------|---------|
| 牌堆构成 | 31张牌无重复，洗牌后随机分布 |
| 发牌隔离 | `GetStateForPlayer(A)` 中不包含玩家A自己的牌架卡，只有其他玩家的牌 |
| 接订单 | 只能选择当前翻开牌的左面或右面，选非法值返回4001错误 |
| 订单更新 | 每次接订单后，订单区正确累加对应水果的数量（1-4个），广播给所有人 |
| 摇铃结算 | 正确计算所有玩家牌架上的水果库存总量（两面水果均计入）|
| 猩猩牌特效 | 猩猩牌持有者不贡献库存，且取消对应类型的所有订单 |
| 缺货判断 | 至少一种水果订单数量 > 库存数量 → 缺货；否则 → 库存充足 |
| 受罚逻辑 | 缺货时上一个接订单者受罚；库存充足时摇铃者受罚 |
| 愤怒标记 | 受罚时取当前标记池中分值最小的一枚 |
| 游戏结束 | 任意玩家愤怒总分 ≥ 7 时，游戏立即结束 |
| 胜者判定 | 愤怒总分最少的玩家获胜，并列则共同获胜 |
| 断线重连 | 3分钟内重连，恢复正确的游戏状态（含其他人牌架信息，不含自己的牌）|
| 操作超时 | 30秒未操作，自动接订单（随机选择左/右水果）|

### 10.2 安全验收

| 安全项 | 验收标准 |
|-------|---------|
| 信息隔离 | 通过抓包验证：任意玩家收到的消息中不含自己的牌架卡数据 |
| 非法选择 | 选择牌上不存在的水果，服务端返回 4001，游戏状态不变 |
| 越权操作 | 非当前行动玩家发送行动，服务端返回 4003，游戏状态不变 |
| 状态一致性 | 结算后所有玩家收到的结算数据中，牌架信息与服务端记录完全一致 |

### 10.3 性能验收

| 指标 | 标准 |
|-----|-----|
| 单服务器承载 | 支持 500 个并发榴莲忘返房间（最多3500名玩家）|
| 消息延迟 | 行动广播延迟 < 100ms（P95）|
| 结算计算耗时 | 库存与订单对比计算 < 10ms（含数量汇总） |
| 内存占用 | 单房间 < 1MB |
| 断线重连 | < 2 秒恢复游戏状态 |

### 10.4 测试用例

**核心测试场景**：
1. ✅ 2人游戏完整流程（发牌→接订单→摇铃→结算→新一轮）
2. ✅ 7人游戏完整流程
3. ✅ 缺货触发（摇铃正确，上一接单者受罚）
4. ✅ 库存充足触发（摇铃错误，摇铃者受罚）
5. ✅ 猩猩大哥触发（榴莲订单清零后重新判断缺货）
6. ✅ 猩猩牌持有者不贡献库存
7. ✅ 愤怒标记累积达到7分触发游戏结束
8. ✅ 并列最低分共同获胜
9. ✅ `GetStateForPlayer` 不泄露自己的牌（安全测试）
10. ✅ 非法操作拒绝（选择不在牌上的水果）
11. ✅ 操作超时自动接订单
12. ✅ 订单区为空时摇铃，摇铃者受罚
13. ✅ 牌堆耗尽后废牌重洗继续游戏
14. ✅ 断线重连后状态恢复（含他人牌架，不含自己牌架）

---

## 11. 后续迭代方向

### 11.1 第一阶段（MVP）

- ✅ 核心玩法：2-7人对局，完整的接订单/摇铃/结算流程
- ✅ 信息隔离：`GetStateForPlayer` 严格隔离自己牌架信息
- ✅ 愤怒标记系统：分值制，游戏结束判定
- ✅ 猩猩兄妹牌：三种特殊能力
- ✅ WebSocket 通信：与现有项目一致的协议风格
- ✅ 断线重连：3分钟内恢复

### 11.2 第二阶段（体验优化）

- 🔲 新手教学模式（临时透视自己的牌，教学局不计分）
- 🔲 水果图案 UI 优化（生动的表情包式水果图案）
- 🔲 操作动效（翻牌动画、摇铃音效、愤怒标记飞入动画）
- 🔲 对局回放系统（复盘功能）
- 🔲 排行榜（按愤怒平均分排名）

### 11.3 第三阶段（玩法扩展）

- 🔲 **进阶版猩猩牌**：增加更多复杂能力（如「互换两人牌架」「查看自己的牌1秒」）
- 🔲 **自定义规则**：房主可调整愤怒标记阈值（5分/7分/10分模式）
- 🔲 **水果种类扩展**：增加第5种水果（如芒果），调整牌组构成
- 🔲 **AI 对战模式**：单人与 AI 练习（AI 可配置难度，体现不同推理水平）
- 🔲 **快速模式**：牌堆减半，游戏时间压缩到 5 分钟内

### 11.4 第四阶段（商业化）

- 🔲 主题皮肤系统（牌背、牌架、桌面主题）
- 🔲 段位系统（根据愤怒分均值和胜率评级）
- 🔲 赛季排行与奖励
- 🔲 好友私房（邀请码加入）
- 🔲 观战模式（观战者可看到所有人的牌架，包括每人自己看不到的牌）

---

## 附录

### 附录A：术语表

| 术语 | 解释 |
|-----|------|
| 牌架卡 | 每轮玩家背对自己插入牌架的那张牌，自己不可见 |
| 总库存 | 所有玩家牌架卡上的水果总量（左右两面均计入）|
| 订单区 | 公共区域，记录本轮每种水果的需求数量 |
| 摇铃 | 宣告「订单已超出库存」并触发结算的行动 |
| 愤怒标记 | 受罚时获得的分值标记，累计7分导致游戏结束 |
| 猩猩兄妹牌 | 特殊牌，不贡献库存，能取消某类水果的所有订单 |
| 缺货 | 某种水果的订单需求 > 库存数量 |
| 印度扑克 | 指「持牌人看不到自己的牌」这一类游戏机制 |

### 附录B：库存计算伪代码

```go
// settlement.go
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
            inventory[c.LeftFruit] += c.LeftCount   // 累加左侧水果数量（1-4）
            inventory[c.RightFruit] += c.RightCount // 累加右侧水果数量（1-4）
        case *GorillaCard:
            // 猩猩牌不贡献库存
        }
    }
    return inventory
}

func ApplyGorillaEffects(orders map[FruitType]int, holderCards map[string]Card) map[FruitType]int {
    result := copyMap(orders)
    for _, card := range holderCards {
        if g, ok := card.(*GorillaCard); ok {
            switch g.Ability {
            case AbilityCancelDurian:
                result[FruitDurian] = 0
            case AbilityCancelBanana:
                result[FruitBanana] = 0
            case AbilityCancelGrape:
                result[FruitGrape] = 0
            }
        }
    }
    return result
}

func IsShortage(orders, inventory map[FruitType]int) (bool, []FruitType) {
    shortFruits := []FruitType{}
    for fruit, demand := range orders {
        if demand > inventory[fruit] {
            shortFruits = append(shortFruits, fruit)
        }
    }
    return len(shortFruits) > 0, shortFruits
}
```

### 附录C：开发优先级

**P0（必须，MVP 核心）**：
- 牌堆构成与洗牌
- 发牌阶段（每人抽 1 张，服务端记录，信息隔离）
- 接订单行动（翻牌 → 选水果 → 更新订单区）
- 摇铃行动 → 库存计算 → 缺货判断 → 受罚逻辑
- 愤怒标记系统（分值制、游戏结束判定）
- `GetStateForPlayer` 信息隔离（玩家看不到自己的牌）
- WebSocket 消息协议

**P1（重要）**：
- 猩猩兄妹牌特殊能力
- 断线重连
- 操作超时自动处理
- 牌堆耗尽后废牌重洗

**P2（可选）**：
- 新手教学模式
- 对局回放
- 排行榜
- 结算动效

---

**文档版本历史**：
- v1.0 (2026-03-02)：初始版本，完整 PRD

**评审状态**：待评审

**预计开发周期**：
- 核心玩法：2 周
- UI/UX：1.5 周
- 测试优化：0.5 周
- **总计：4 周**
