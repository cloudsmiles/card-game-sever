# 麻将游戏产品需求文档（PRD）

版本：v1.0  
创建日期：2026-02-28  
产品经理：Card Game PM

---

## 1. 产品概述

### 1.1 游戏名称与定位
- **游戏名称**：国标麻将（简化版）
- **游戏定位**：面向在线玩家的4人对局麻将游戏，采用简化的国标规则，兼顾竞技性与易上手性
- **技术架构**：复用现有card-game-server的三层架构（Core核心逻辑层 / Engine引擎层 / Adapter适配器层），确保与斗地主游戏共享房间管理、玩家连接、WebSocket通信等基础设施

### 1.2 目标用户群体
- **核心用户**：18-45岁，熟悉麻将规则的休闲玩家
- **次要用户**：对传统麻将感兴趣的新手玩家，希望体验标准化规则
- **使用场景**：碎片化时间娱乐、朋友联机对战、竞技排位

### 1.3 核心体验关键词
- **标准化**：采用国标规则简化版，减少地域规则差异
- **快节奏**：单局15-25分钟，适合移动端短时娱乐
- **易懂难精**：基础规则简单，但番种组合需要策略思考
- **公平竞技**：无付费增益，纯技术对抗

### 1.4 差异化亮点
1. **简化番种系统**：从国标111种番种简化到30种常见番种，降低新手门槛
2. **智能提示系统**：自动检测可胡牌型，标记危险牌，提升用户体验
3. **架构复用**：继承现有服务器架构，开发周期短，维护成本低
4. **断线重连**：完整的断线保护机制，玩家可在3分钟内重连继续游戏

---

## 2. 核心机制

### 2.1 游戏人数与角色分配
- **固定4人对局**：东、南、西、北四个方位
- **庄家机制**：
  - 首局随机选定庄家（东家）
  - 庄家胡牌或流局听牌 → 连庄
  - 闲家胡牌 → 庄家按逆时针轮转（东→北→西→南→东）
- **座次顺序**：逆时针轮流出牌（东→南→西→北→东）

### 2.2 牌组构成与发牌规则
**牌组总数**：136张牌
- **万子牌**：1-9万，各4张，共36张
- **条子牌**：1-9条，各4张，共36张
- **筒子牌**：1-9筒，各4张，共36张
- **字牌**：东南西北中发白，各4张，共28张

**发牌规则**：
1. 洗牌并砌牌（虚拟牌墙）
2. 庄家掷骰子（2个骰子，点数2-12），确定开牌位置
3. 按逆时针顺序发牌：
   - 庄家14张（含起手自摸1张）
   - 闲家各13张
4. 牌墙末尾14张为杠牌（宝牌指示牌区域，本版本暂不实现宝牌）

### 2.3 回合流程与行动选择

#### 2.3.1 回合结构
```
循环 {
  当前玩家摸牌（从牌墙顶部）
  → 判断：是否自摸胡牌？
     - 是 → 进入结算
     - 否 → 玩家选择行动
  → 玩家出牌
  → 其他玩家判断：是否可吃/碰/杠/胡？
     - 有胡 → 优先级最高，直接结算
     - 有杠 → 执行杠牌，摸杠牌，继续出牌
     - 有碰 → 执行碰牌，继续出牌
     - 有吃 → 只有下家可吃，执行后出牌
     - 无响应 → 进入下一家回合
}
```

#### 2.3.2 行动选项
| 行动 | 条件 | 说明 |
|-----|------|-----|
| **出牌** | 必选 | 从手牌中打出1张牌 |
| **吃** | 上家刚打出的牌 | 用手牌中的2张牌组成顺子（仅限下家，只能吃上家的牌）|
| **碰** | 任意家刚打出的牌 | 用手牌中的2张相同牌组成刻子 |
| **明杠** | 任意家刚打出的牌 | 用手牌中的3张相同牌组成杠（点杠）|
| **暗杠** | 自己摸牌后 | 手牌有4张相同牌（暗杠）|
| **补杠** | 已碰过的牌 | 摸到第4张相同牌，补杠 |
| **胡牌** | 符合胡牌条件 | 自摸或点炮胡牌 |
| **过** | 可选 | 放弃吃/碰/杠的机会 |

#### 2.3.3 行动优先级
```
胡牌（点炮） > 杠（碰杠/明杠） > 碰 > 吃
```
- 多人同时胡牌：按座次逆时针，距离打牌者最近的人优先（"截胡"）
- 一人喊胡后，其他人的杠/碰/吃操作失效

### 2.4 胜负判定条件

#### 2.4.1 胡牌基本型
标准胡牌牌型：**3n+2结构**（n≥4）
- **4组面子 + 1对将**
- **面子**：顺子（123万）、刻子（333条）、杠（4444筒）
- **将牌**：任意1对相同的牌（对子）

特殊牌型（满足即可胡）：
- **七对子**：7个对子（2*7=14张）
- **十三幺**：1/9万条筒 + 东南西北中发白各1张，其中1张做对子（共14张）

#### 2.4.2 胡牌必备条件
1. **满足基本牌型**（3n+2 或 特殊牌型）
2. **至少8番**（简化规则，降低胡牌门槛）
3. **无诈胡**（系统自动校验）

#### 2.4.3 结算方式
- **自摸**：三家各支付胡牌者番数对应的底分
- **点炮**：点炮者支付全部底分
- **流局**：牌墙摸完仍无人胡牌，听牌者瓜分未听牌者底分

---

## 3. 麻将牌设计

### 3.1 牌面定义

#### 3.1.1 数据结构
```go
type MahjongTile struct {
    ID       int    // 唯一标识（1-136）
    Suit     string // 花色：wan(万), tiao(条), tong(筒), zi(字)
    Rank     int    // 点数：1-9（序数牌），1-7（字牌：东南西北中发白）
    IsHonor  bool   // 是否为字牌
}
```

#### 3.1.2 牌面编码规则
| 花色 | Suit代码 | Rank范围 | 示例 |
|-----|---------|---------|------|
| 万子 | wan | 1-9 | 1万=wan1, 9万=wan9 |
| 条子 | tiao | 1-9 | 1条=tiao1, 9条=tiao9 |
| 筒子 | tong | 1-9 | 1筒=tong1, 9筒=tong9 |
| 字牌 | zi | 1-7 | 东=zi1, 南=zi2, 西=zi3, 北=zi4, 中=zi5, 发=zi6, 白=zi7 |

### 3.2 牌型分类

#### 3.2.1 面子类型
| 类型 | 定义 | 示例 | 计数 |
|-----|------|------|------|
| **顺子** | 同花色连续3张 | 1万2万3万, 5条6条7条 | 普通面子 |
| **刻子** | 3张相同牌 | 3筒3筒3筒, 东东东 | 普通面子 |
| **杠** | 4张相同牌 | 8万8万8万8万 | 特殊面子（需补牌）|

#### 3.2.2 特殊牌组
| 类型 | 组成 | 说明 |
|-----|------|------|
| **老头牌** | 1、9万条筒 | 用于"全带幺"等番种 |
| **幺九牌** | 1、9万条筒 + 字牌 | 用于"混幺九"等番种 |
| **绿一色** | 2、3、4、6、8条 + 发 | 特殊番种（本版本可选实现）|

### 3.3 番种设计（简化版30种）

#### 3.3.1 番种分级
| 番数 | 番种名称 | 说明 | 示例 |
|-----|---------|------|------|
| **88番** | 大四喜 | 4副风刻（杠）| 东东东 南南南 西西西 北北北 + 将 |
| **88番** | 大三元 | 3副箭刻（杠）| 中中中 发发发 白白白 + 2组面子 + 将 |
| **64番** | 小四喜 | 3副风刻 + 1对风将 | 东东东 南南南 西西西 + 北北（将）|
| **64番** | 小三元 | 2副箭刻 + 1对箭将 | 中中中 发发发 + 白白（将）|
| **64番** | 字一色 | 全部字牌 | 东东东 南南 中中中 发发发 白白白白 |
| **64番** | 四暗刻 | 4组暗刻 + 自摸胡 | （手牌）222万 555条 888筒 东东东 + 66（将）|
| **32番** | 一色双龙会 | 同花色2组123、789 + 5做将 | 123万 789万 123万 789万 + 55万 |
| **24番** | 七对 | 7个对子 | 11万 33条 55筒 77筒 东东 中中 发发 |
| **24番** | 清一色 | 单一花色（无字牌）| 11223344556677万 |
| **24番** | 一色四同顺 | 同花色4组相同顺子 | 123万 123万 123万 123万 + 其他面子 + 将 |
| **16番** | 三暗刻 | 3组暗刻 | （手牌）222万 555条 888筒 + 其他面子 + 将 |
| **12番** | 全大 | 全部789牌 | 789万 777条 999筒 888万 + 77条 |
| **12番** | 全中 | 全部456牌 | 456万 444条 666筒 555条 + 44万 |
| **12番** | 全小 | 全部123牌 | 123万 111条 333筒 222条 + 11万 |
| **8番** | 混一色 | 单一花色 + 字牌 | 123万 456万 789万 东东东 + 11万 |
| **6番** | 碰碰胡 | 4组刻子（杠）+ 将 | 333万 555条 东东东 中中中 + 77筒 |
| **6番** | 全带幺 | 每组面子/将都含1或9 | 123万 789条 111筒 999万 + 99条 |
| **4番** | 全求人 | 4组面子全靠吃/碰 + 单钓将胡牌 | （明面）123万 456条 东东东 + 单钓5筒 |
| **2番** | 箭刻 | 1组中/发/白的刻子 | 中中中 + 其他面子 |
| **2番** | 圈风刻 | 与圈风相同的风刻（东风圈=东刻）| 东东东（东风圈时）|
| **2番** | 门风刻 | 与门风相同的风刻（东家=东刻）| 南南南（南家时）|
| **2番** | 门前清 | 无吃/碰/明杠，自摸胡 | （手牌）未吃碰过 + 自摸 |
| **2番** | 平胡 | 4组顺子 + 非字牌将 + 边/坎/单钓 | 123万 456条 789筒 234万 + 55条（边坎单）|
| **2番** | 四归一 | 4张相同牌分散在刻子/将/顺子中 | 123万（1万1张） + 111万（3张）= 4张1万 |
| **1番** | 一般高 | 2组相同花色相同点数的顺子 | 123万 123万 + 其他面子 |
| **1番** | 连六 | 同花色6张连续序数牌 | 123456万 + 其他面子 |
| **1番** | 老少副 | 同花色123 + 789的2组顺子 | 123万 789万 + 其他面子 |
| **1番** | 幺九刻 | 1组1或9的刻子 | 111万 或 999条 |
| **1番** | 明杠 | 有明杠（点杠）| 任意明杠 |
| **1番** | 暗杠 | 有暗杠 | 任意暗杠 |

#### 3.3.2 番种计算规则
1. **不重复原则**：高级番种包含低级特征时，不重复计分
   - 示例：清一色（24番）已包含"无字牌"，不再计"无字"的分
2. **可叠加番种**：不同维度的番种可累加
   - 示例：碰碰胡（6番）+ 混一色（8番）= 14番
3. **封顶规则**：单局最高**88番**
4. **最低胡牌**：至少**8番**才可胡牌（简化规则）

---

## 4. 平衡性设计

### 4.1 各角色胜率预期
| 角色 | 理论胜率 | 说明 |
|-----|---------|------|
| 庄家（东家）| 26-28% | 多摸1张牌，优先出牌，连庄优势 |
| 闲家（南西北）| 24-25%/人 | 均等机会 |

**平衡措施**：
- 庄家胡牌，闲家支付双倍底分（自摸*2，点炮*1.5）
- 连庄3次后强制轮庄（防止庄家优势过大）

### 4.2 运气与技巧占比
- **运气占比**：40%（起手牌、摸牌顺序）
- **技巧占比**：60%（舍牌策略、听牌选择、防守意识）

**技巧体现维度**：
1. **进攻**：如何快速听牌、选择高番牌型
2. **防守**：识别他人听牌，避免点炮
3. **决策**：是否吃碰（暴露手牌vs加速听牌）

### 4.3 反制机制设计
| 策略 | 反制手段 |
|-----|---------|
| 快速胡牌（低番）| 对手可选择防守，拖入流局 |
| 追求高番 | 听牌慢，易被他人抢先胡牌 |
| 过度吃碰 | 手牌暴露，易被针对防守 |
| 保守防守 | 可能错失胡牌机会 |

### 4.4 数值调节方案
**底分-番数对照表**（建议采用二次曲线）
| 番数 | 底分（游戏币）| 说明 |
|-----|------------|------|
| 8番 | 10 | 最低胡牌 |
| 16番 | 30 | 三暗刻 |
| 24番 | 60 | 清一色、七对 |
| 32番 | 120 | 一色双龙会 |
| 64番 | 300 | 小四喜、字一色 |
| 88番 | 500 | 封顶大牌 |

**调节参数**：
- `BaseCoin`：基础底分（默认10）
- `FanMultiplier`：番数倍率系数（可动态调整，控制经济产出）
- `DealerMultiplier`：庄家倍率（默认1.5）

---

## 5. 游戏流程与状态机

### 5.1 游戏状态定义
```go
type GameState int

const (
    StateWaiting      GameState = 0  // 等待玩家
    StateReady        GameState = 1  // 准备阶段（4人已满）
    StateDealing      GameState = 2  // 发牌中
    StatePlaying      GameState = 3  // 游戏中
    StatePending      GameState = 4  // 等待玩家响应（吃/碰/杠/胡）
    StateRoundEnd     GameState = 5  // 单局结束（展示结算）
    StateGameEnd      GameState = 6  // 整场结束（4圈结束或手动解散）
)
```

### 5.2 状态流转图
```
StateWaiting (0) ──【4人准备】──> StateReady (1)
                        │
                        ▼
                   StateDealing (2) ──【发牌完成】──> StatePlaying (3)
                        │                               │
                        │                               ▼
                        │                         ┌─────────────┐
                        │                         │  正常出牌   │
                        │                         │  回合流转   │
                        │                         └──────┬──────┘
                        │                                │
                        │                                ▼
                        │                        StatePending (4) ──【无人响应】──> StatePlaying (3)
                        │                                │                               ▲
                        │                                │【有人响应】                    │
                        │                                ▼                               │
                        │                          ┌───────────┐                         │
                        │                          │  吃/碰/杠  │ ────────────────────────┘
                        │                          └───────────┘
                        │                                │【胡牌】
                        │                                ▼
                        └────────────────────> StateRoundEnd (5)
                                                           │
                                                           │【圈数未满】
                                                           ├──────────> StateDealing (2)
                                                           │
                                                           │【圈数结束】
                                                           └──────────> StateGameEnd (6)
```

### 5.3 关键流程详细说明

#### 5.3.1 开局流程
```
1. 房间创建 → StateWaiting
2. 玩家加入（4人） → StateReady
3. 倒计时3秒 → StateDealing
4. 执行发牌：
   - 洗牌（Fisher-Yates算法）
   - 庄家掷骰子（随机2-12）
   - 分配手牌（庄家14张，闲家13张）
   - 广播手牌给各玩家（仅发送自己的牌）
5. 庄家开始出牌 → StatePlaying
```

#### 5.3.2 回合流程
```
A. 玩家摸牌阶段
   - 从牌墙顶部摸1张
   - 检查是否自摸胡牌
     - 是 → StateRoundEnd
     - 否 → 等待玩家操作

B. 玩家出牌阶段
   - 玩家选择1张手牌打出
   - 广播出牌事件给所有人
   - 进入StatePending（等待响应）

C. 响应等待阶段（StatePending）
   - 3秒倒计时，允许其他玩家：
     * 胡牌（点炮）
     * 杠（明杠/碰杠）
     * 碰
     * 吃（仅下家）
   - 优先级判断：
     * 胡 > 杠 > 碰 > 吃
     * 同优先级按座次顺序
   - 倒计时结束或所有人响应 → 执行最高优先级操作

D. 操作执行
   - 胡牌 → StateRoundEnd
   - 杠 → 玩家摸杠牌，继续出牌（回到B）
   - 碰 → 玩家展示碰牌，继续出牌（回到B）
   - 吃 → 玩家展示吃牌，继续出牌（回到B）
   - 无响应 → 下一家摸牌（回到A）
```

#### 5.3.3 结算流程
```
1. 触发结算条件：
   - 有人胡牌（自摸/点炮）
   - 流局（牌墙摸完）

2. 计算分数：
   - 识别胡牌玩家的所有番种
   - 累加番数（不重复原则）
   - 计算底分：BaseCoin * FanMultiplier^(番数/8)
   - 庄家倍率加成

3. 结算规则：
   - 自摸：三家各支付底分
   - 点炮：点炮者支付3倍底分（替三家支付）
   - 流局：听牌者瓜分未听牌者各1000分

4. 广播结算信息：
   - 胡牌玩家手牌
   - 番种列表
   - 分数变化
   - 庄家轮转信息

5. 进入StateRoundEnd（展示5秒）

6. 判断游戏是否继续：
   - 未满4圈（东南西北圈）→ StateDealing（下一局）
   - 已满4圈 或 玩家主动解散 → StateGameEnd
```

### 5.4 异常处理路径
| 异常场景 | 处理逻辑 |
|---------|---------|
| 玩家断线（游戏中）| 托管AI代打，3分钟内可重连恢复控制 |
| 玩家断线（等待阶段）| 60秒未重连，踢出房间，其他玩家返回大厅 |
| 操作超时（出牌）| 15秒未操作，自动打出第一张牌 |
| 操作超时（响应）| 3秒未响应，自动"过" |
| 非法操作（出牌不在手牌中）| 拒绝操作，返回错误码，要求重新操作 |
| 作弊检测（同IP多开）| 标记风险账号，限制匹配，记录日志 |
| 服务器崩溃 | 持久化房间状态到Redis，重启后恢复游戏 |

---

## 6. WebSocket协议设计

### 6.1 协议格式
```json
{
  "type": "消息类型",
  "data": {
    // 具体数据
  },
  "timestamp": 1709107200000
}
```

### 6.2 客户端 → 服务器消息

#### 6.2.1 加入房间
```json
{
  "type": "join_room",
  "data": {
    "room_id": "room_12345",
    "player_id": "player_67890",
    "token": "auth_token_here"
  }
}
```

#### 6.2.2 准备游戏
```json
{
  "type": "ready",
  "data": {
    "player_id": "player_67890"
  }
}
```

#### 6.2.3 出牌
```json
{
  "type": "discard",
  "data": {
    "player_id": "player_67890",
    "tile": {
      "suit": "wan",
      "rank": 5
    }
  }
}
```

#### 6.2.4 吃牌
```json
{
  "type": "chow",
  "data": {
    "player_id": "player_67890",
    "target_tile": {"suit": "wan", "rank": 3},
    "hand_tiles": [
      {"suit": "wan", "rank": 1},
      {"suit": "wan", "rank": 2}
    ]
  }
}
```

#### 6.2.5 碰牌
```json
{
  "type": "pong",
  "data": {
    "player_id": "player_67890",
    "target_tile": {"suit": "tiao", "rank": 5}
  }
}
```

#### 6.2.6 杠牌
```json
{
  "type": "kong",
  "data": {
    "player_id": "player_67890",
    "kong_type": "concealed", // concealed(暗杠), exposed(明杠), extended(补杠)
    "tile": {"suit": "tong", "rank": 8}
  }
}
```

#### 6.2.7 胡牌
```json
{
  "type": "win",
  "data": {
    "player_id": "player_67890",
    "win_type": "self_drawn" // self_drawn(自摸), discard(点炮)
  }
}
```

#### 6.2.8 过（放弃操作）
```json
{
  "type": "pass",
  "data": {
    "player_id": "player_67890"
  }
}
```

### 6.3 服务器 → 客户端消息

#### 6.3.1 房间状态更新
```json
{
  "type": "room_state",
  "data": {
    "room_id": "room_12345",
    "state": 3, // StatePlaying
    "players": [
      {
        "player_id": "player_67890",
        "nickname": "张三",
        "position": "east", // east, south, west, north
        "is_dealer": true,
        "score": 25000,
        "is_ready": true
      }
      // ... 其他3个玩家
    ],
    "current_round": 1, // 当前局数（1-16，4圈*4局）
    "dealer_position": "east"
  }
}
```

#### 6.3.2 发牌通知
```json
{
  "type": "deal_tiles",
  "data": {
    "player_id": "player_67890",
    "hand_tiles": [
      {"suit": "wan", "rank": 1},
      {"suit": "wan", "rank": 3},
      // ... 共13或14张
    ],
    "wall_remaining": 70 // 剩余牌墙数量
  }
}
```

#### 6.3.3 摸牌通知
```json
{
  "type": "draw_tile",
  "data": {
    "player_id": "player_67890",
    "tile": {"suit": "tiao", "rank": 5},
    "wall_remaining": 69
  }
}
```

#### 6.3.4 出牌广播
```json
{
  "type": "tile_discarded",
  "data": {
    "player_id": "player_11111",
    "tile": {"suit": "tong", "rank": 9},
    "discarded_tiles": [
      {"suit": "wan", "rank": 1},
      {"suit": "zi", "rank": 5},
      // ... 该玩家所有已打出的牌
    ]
  }
}
```

#### 6.3.5 操作请求通知
```json
{
  "type": "action_request",
  "data": {
    "player_id": "player_67890",
    "available_actions": ["chow", "pong", "win"], // 可执行的操作
    "target_tile": {"suit": "wan", "rank": 7},
    "timeout": 3000 // 毫秒
  }
}
```

#### 6.3.6 吃碰杠广播
```json
{
  "type": "meld_exposed",
  "data": {
    "player_id": "player_67890",
    "meld_type": "pong", // chow, pong, kong
    "tiles": [
      {"suit": "tiao", "rank": 5},
      {"suit": "tiao", "rank": 5},
      {"suit": "tiao", "rank": 5}
    ],
    "from_player_id": "player_11111" // 被吃碰杠的牌来自哪个玩家
  }
}
```

#### 6.3.7 胡牌结算
```json
{
  "type": "round_result",
  "data": {
    "winner_id": "player_67890",
    "win_type": "self_drawn",
    "loser_id": null, // 点炮者ID（自摸时为null）
    "hand_tiles": [
      {"suit": "wan", "rank": 1},
      // ... 胡牌者的完整手牌
    ],
    "fan_list": [
      {"name": "清一色", "fan": 24},
      {"name": "门前清", "fan": 2},
      {"name": "一般高", "fan": 1}
    ],
    "total_fan": 27,
    "base_score": 80,
    "score_changes": [
      {"player_id": "player_67890", "change": +240},
      {"player_id": "player_11111", "change": -80},
      {"player_id": "player_22222", "change": -80},
      {"player_id": "player_33333", "change": -80}
    ],
    "next_dealer": "east" // 下一局庄家位置
  }
}
```

#### 6.3.8 流局通知
```json
{
  "type": "draw_game",
  "data": {
    "reason": "wall_empty", // 牌墙摸完
    "listening_players": ["player_67890", "player_11111"],
    "score_changes": [
      {"player_id": "player_67890", "change": +1000},
      {"player_id": "player_11111", "change": +1000},
      {"player_id": "player_22222", "change": -1000},
      {"player_id": "player_33333", "change": -1000}
    ]
  }
}
```

#### 6.3.9 错误通知
```json
{
  "type": "error",
  "data": {
    "code": 4001,
    "message": "非法操作：该牌不在手牌中",
    "details": {
      "requested_tile": {"suit": "wan", "rank": 5}
    }
  }
}
```

### 6.4 错误码定义
| 错误码 | 说明 |
|-------|------|
| 4001 | 非法操作（牌不在手牌中、不符合规则）|
| 4002 | 操作超时 |
| 4003 | 不在玩家回合 |
| 4004 | 房间状态错误（不在游戏中）|
| 4005 | 权限不足（未认证）|
| 5001 | 服务器内部错误 |

---

## 7. 架构适配说明

### 7.1 复用现有架构
基于项目现有的三层架构，麻将游戏模块应遵循以下分层：

#### 7.1.1 Core层（核心逻辑）
路径：`internal/game/core/mahjong/`
- `mahjong_game.go`：麻将游戏核心逻辑（实现GameCore接口）
- `tile.go`：麻将牌数据结构与工具函数
- `hand.go`：手牌管理（听牌判断、胡牌校验）
- `fan_calculator.go`：番种识别与计算
- `meld.go`：面子（吃碰杠）管理
- `rule_validator.go`：规则校验器

**核心接口**：
```go
type MahjongGame interface {
    // 继承通用游戏接口
    game.GameCore
    
    // 麻将特有方法
    DrawTile(playerID string) (*Tile, error)
    DiscardTile(playerID string, tile *Tile) error
    Chow(playerID string, tiles []*Tile) error
    Pong(playerID string) error
    Kong(playerID string, kongType KongType) error
    DeclareWin(playerID string) (*WinResult, error)
    CalculateFan(hand *Hand) ([]*FanItem, int)
}
```

#### 7.1.2 Engine层（游戏引擎）
路径：`internal/game/engine/`
- 复用现有的`game_engine.go`
- 扩展房间类型支持：`RoomTypeMahjong = 2`
- 注册麻将游戏工厂：
```go
engine.RegisterGameFactory(RoomTypeMahjong, func() GameCore {
    return mahjong.NewMahjongGame()
})
```

#### 7.1.3 Adapter层（协议适配）
路径：`internal/game/adapter/`
- 复用现有的`websocket_adapter.go`
- 扩展消息类型，增加麻将专用消息（如`chow`, `pong`, `kong`）
- 复用心跳检测、断线重连逻辑

### 7.2 共享基础设施
| 模块 | 路径 | 复用说明 |
|-----|------|---------|
| 房间管理 | `internal/room/manager.go` | 直接复用，支持麻将房间类型 |
| 玩家管理 | `internal/player/manager.go` | 复用玩家连接、认证逻辑 |
| 消息路由 | `internal/server/router.go` | 扩展路由规则，增加麻将消息处理器 |
| 状态持久化 | `internal/storage/redis.go` | 复用Redis存储，保存麻将房间状态 |
| 日志系统 | `pkg/logger/` | 直接复用 |

### 7.3 数据库扩展
**新增表结构**（参考斗地主设计）：
```sql
-- 麻将游戏记录表
CREATE TABLE mahjong_game_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id VARCHAR(32) NOT NULL,
    round_num INT NOT NULL,
    winner_id VARCHAR(32),
    win_type ENUM('self_drawn', 'discard', 'draw') NOT NULL,
    loser_id VARCHAR(32),
    total_fan INT,
    base_score INT,
    hand_tiles JSON, -- 胡牌者手牌
    fan_list JSON,   -- 番种列表
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_winner_id (winner_id)
);

-- 玩家成绩表（复用现有表，增加麻将字段）
ALTER TABLE player_stats ADD COLUMN mahjong_games INT DEFAULT 0;
ALTER TABLE player_stats ADD COLUMN mahjong_wins INT DEFAULT 0;
ALTER TABLE player_stats ADD COLUMN mahjong_avg_fan FLOAT DEFAULT 0;
```

---

## 8. 风险与应对

### 8.1 可能的作弊手段
| 作弊方式 | 检测手段 | 防范措施 |
|---------|---------|---------|
| 同IP多开 | 记录IP，同IP限制同房间数 | 同IP最多1个账号在同一房间 |
| 协同作弊（通讯工具）| 检测异常胜率、分数流向 | AI分析，疑似作弊账号隔离匹配 |
| 抓包修改出牌 | 服务器强校验，不信任客户端 | 所有操作服务器验证手牌合法性 |
| 刷分（故意输分）| 检测连续异常对局 | 触发风控，冻结账号 |
| 断线逃跑 | 记录逃跑次数 | 逃跑3次/日封禁24小时 |

### 8.2 极端情况处理
| 情况 | 处理方案 |
|-----|---------|
| 四人同时听牌流局 | 按听牌张数多少分配流局分，相同则平分 |
| 多人同时胡同一张牌 | 按座次距离打牌者最近的人优先（截胡）|
| 抢杠胡 | 有人补杠时，其他人可用该牌胡（抢杠胡，额外+1番）|
| 网络波动导致消息乱序 | 消息携带序列号，服务器按序处理，丢弃过期消息 |
| 玩家长时间不操作 | 15秒自动出牌，3次触发后转托管AI |
| 天胡（庄家起手14张直接胡）| 作为特殊番种，88番封顶 |
| 四杠流局 | 4次杠出现后，牌墙剩余牌<10张时流局 |

### 8.3 玩家体验保障
1. **新手引导**：
   - 前3局开启"新手模式"，自动标注可胡牌型
   - 危险牌高亮提示（可关闭）
   
2. **智能提示**：
   - 听牌自动检测并提示
   - 可胡牌时自动弹窗（防止漏胡）
   
3. **回放系统**：
   - 每局游戏保存完整牌谱
   - 玩家可复盘查看操作记录
   
4. **公平性保证**：
   - 洗牌算法采用密码学安全的随机数生成器
   - 发牌过程写入日志，可追溯审计

---

## 9. 验收标准

### 9.1 功能验收
| 功能模块 | 验收标准 |
|---------|---------|
| 洗牌发牌 | 136张牌无重复，随机分布，庄家14张闲家13张 |
| 出牌逻辑 | 只能打出手牌中的牌，出牌后正确广播给其他玩家 |
| 吃牌 | 仅下家可吃，组成顺子后正确移除手牌并展示 |
| 碰牌 | 任意玩家可碰，组成刻子后正确移除手牌并展示 |
| 杠牌 | 暗杠/明杠/补杠三种类型正确处理，摸杠牌逻辑正确 |
| 听牌判断 | 准确识别听牌状态，提示可胡的牌 |
| 胡牌校验 | 严格验证牌型合法性，至少8番才可胡 |
| 番种计算 | 30种番种识别准确率100%，不重复计分规则正确 |
| 自摸结算 | 三家各扣分，胡牌者得分 = 3倍底分 |
| 点炮结算 | 点炮者扣3倍底分，其他两家不扣分 |
| 流局处理 | 听牌者瓜分未听牌者分数，庄家轮转正确 |
| 断线重连 | 3分钟内重连恢复完整游戏状态（手牌、出牌记录、当前回合）|
| 托管AI | 15秒未操作自动出牌，AI决策合理（不点炮为优先）|

### 9.2 性能验收
| 指标 | 标准 |
|-----|-----|
| 单服务器承载 | 支持1000个并发房间（4000玩家在线）|
| 消息延迟 | 出牌到广播延迟 < 100ms（P95）|
| 番种计算耗时 | < 50ms（复杂牌型）|
| 内存占用 | 单房间 < 5MB |
| 断线重连 | < 2秒恢复游戏状态 |

### 9.3 测试用例
**核心测试场景**：
1. ✅ 正常对局完整流程（发牌→出牌→结算）
2. ✅ 七对子胡牌（24番）
3. ✅ 清一色 + 碰碰胡（30番）
4. ✅ 抢杠胡（特殊情况）
5. ✅ 四人流局（牌墙摸完）
6. ✅ 多人同时喊胡（截胡逻辑）
7. ✅ 连庄3次后强制轮庄
8. ✅ 玩家中途断线重连
9. ✅ 非法操作拒绝（打出不在手牌中的牌）
10. ✅ 番种计算边界测试（88番封顶）

---

## 10. 后续迭代方向

### 10.1 第一阶段（MVP）
- ✅ 核心玩法：4人对局、30种基础番种
- ✅ 基础UI：手牌展示、出牌动画、结算面板
- ✅ 房间系统：创建/加入/解散
- ✅ 断线重连

### 10.2 第二阶段（体验优化）
- 🔲 新手引导与教学模式
- 🔲 智能提示系统（听牌提示、危险牌标记）
- 🔲 牌谱回放系统
- 🔲 排行榜与成就系统
- 🔲 好友房间（邀请制）

### 10.3 第三阶段（玩法扩展）
- 🔲 竞技场模式（匹配系统）
- 🔲 番种扩展（增加到50种）
- 🔲 宝牌系统（翻宝牌，增加随机性）
- 🔲 血战到底模式（一人胡牌，其他人继续）
- 🔲 AI对战模式（单人练习）

### 10.4 第四阶段（商业化）
- 🔲 段位系统（青铜→王者）
- 🔲 赛季结算与奖励
- 🔲 皮肤系统（牌面、牌桌）
- 🔲 观战系统
- 🔲 战队/公会功能

---

## 11. 附录

### 11.1 术语表
| 术语 | 解释 |
|-----|------|
| 番 | 麻将计分单位，类似"倍数" |
| 面子 | 顺子、刻子、杠的统称 |
| 将牌 | 胡牌时的对子（雀头）|
| 听牌 | 只差1张牌即可胡牌的状态 |
| 门前清 | 未吃碰杠，手牌全暗 |
| 暗刻 | 手牌中的3张相同牌（未碰出）|
| 明刻 | 碰出的3张相同牌 |
| 宝牌 | 翻出的指示牌，持有则加番（本版本不实现）|
| 振听 | 打过的牌自己不能胡（本版本不实现）|

### 11.2 参考资料
- 《中国麻将竞赛规则（2006版）》
- 国标麻将番种标准表
- 雀魂麻将、天凤麻将规则参考

### 11.3 开发优先级
**P0（必须）**：
- 核心玩法循环（摸牌→出牌→吃碰杠胡）
- 30种基础番种识别
- 自摸/点炮结算
- WebSocket通信

**P1（重要）**：
- 断线重连
- 流局处理
- 托管AI
- 操作超时处理

**P2（可选）**：
- 新手引导
- 听牌提示
- 回放系统
- 排行榜

---

## 附录：代码实现示例

### 示例1：核心数据结构
```go
// internal/game/core/mahjong/tile.go
package mahjong

type Suit string

const (
    SuitWan  Suit = "wan"  // 万
    SuitTiao Suit = "tiao" // 条
    SuitTong Suit = "tong" // 筒
    SuitZi   Suit = "zi"   // 字
)

type Tile struct {
    Suit    Suit   `json:"suit"`
    Rank    int    `json:"rank"` // 1-9 或 1-7(字牌)
    IsHonor bool   `json:"is_honor"`
}

func (t *Tile) IsTerminal() bool {
    return !t.IsHonor && (t.Rank == 1 || t.Rank == 9)
}

func (t *Tile) IsSimple() bool {
    return !t.IsHonor && t.Rank >= 2 && t.Rank <= 8
}

func (t *Tile) Equal(other *Tile) bool {
    return t.Suit == other.Suit && t.Rank == other.Rank
}
```

### 示例2：番种计算接口
```go
// internal/game/core/mahjong/fan_calculator.go
package mahjong

type FanItem struct {
    Name string `json:"name"`
    Fan  int    `json:"fan"`
}

type FanCalculator interface {
    Calculate(hand *Hand) ([]*FanItem, int)
}

type StandardFanCalculator struct{}

func (c *StandardFanCalculator) Calculate(hand *Hand) ([]*FanItem, int) {
    fans := make([]*FanItem, 0)
    
    // 检测清一色
    if c.isPureOneSuit(hand) {
        fans = append(fans, &FanItem{Name: "清一色", Fan: 24})
    }
    
    // 检测碰碰胡
    if c.isAllPongs(hand) {
        fans = append(fans, &FanItem{Name: "碰碰胡", Fan: 6})
    }
    
    // ... 其他番种检测
    
    total := 0
    for _, f := range fans {
        total += f.Fan
    }
    
    return fans, total
}
```

### 示例3：WebSocket消息处理
```go
// internal/game/adapter/mahjong_handler.go
package adapter

func (a *WebSocketAdapter) handleMahjongMessage(msg *Message, player *Player) {
    switch msg.Type {
    case "discard":
        a.handleDiscard(msg, player)
    case "chow":
        a.handleChow(msg, player)
    case "pong":
        a.handlePong(msg, player)
    case "kong":
        a.handleKong(msg, player)
    case "win":
        a.handleWin(msg, player)
    case "pass":
        a.handlePass(msg, player)
    }
}

func (a *WebSocketAdapter) handleDiscard(msg *Message, player *Player) {
    // 1. 解析出牌数据
    var data struct {
        Tile *mahjong.Tile `json:"tile"`
    }
    json.Unmarshal(msg.Data, &data)
    
    // 2. 调用核心逻辑
    game := a.getGame(player.RoomID)
    err := game.DiscardTile(player.ID, data.Tile)
    if err != nil {
        a.sendError(player, 4001, err.Error())
        return
    }
    
    // 3. 广播给其他玩家
    a.broadcast(player.RoomID, &Message{
        Type: "tile_discarded",
        Data: map[string]interface{}{
            "player_id": player.ID,
            "tile":      data.Tile,
        },
    })
}
```

---

**文档版本历史**：
- v1.0 (2026-02-28)：初始版本，完整PRD

**评审状态**：待评审

**预计开发周期**：
- 核心玩法：3周
- UI/UX：2周
- 测试优化：1周
- **总计：6周**
