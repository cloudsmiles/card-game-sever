# Card Game Server

一个基于 Go 和 WebSocket 的多人在线卡牌游戏服务器，支持斗地主、麻将、榴莲忘返等多种游戏模式，配套 React 前端。

## 功能特性

- **多游戏支持**：斗地主(DDZ)、麻将(Mahjong)、榴莲忘返(Durian)
- **实时通信**：基于 WebSocket 的全双工实时通信
- **房间管理**：创建/加入房间、玩家准备、选座系统
- **断线重连**：游戏中断线保护，支持重连恢复
- **聊天系统**：房间内实时聊天功能
- **游戏状态同步**：广播机制确保所有玩家状态一致
- **机器人系统**：支持添加 AI 机器人填充空位
- **超时处理**：回合超时自动操作，防止游戏卡死

## 技术栈

### 后端
- **语言**：Go 1.25+
- **通信协议**：WebSocket (gorilla/websocket)
- **架构模式**：模块化设计，接口抽象

### 前端
- **框架**：React 18 + TypeScript
- **构建工具**：Vite 5
- **UI 库**：Material-UI 5
- **状态管理**：Zustand
- **动画**：Framer Motion
- **路由**：React Router 6

## 项目结构

```
card-game-server/
├── backend/                 # Go 后端
│   ├── internal/
│   │   ├── connection/      # WebSocket 连接管理
│   │   │   ├── connection.go  # 连接追踪、心跳检测
│   │   │   ├── handler.go     # 消息处理器
│   │   │   └── http.go        # HTTP 路由
│   │   ├── game/            # 游戏逻辑
│   │   │   ├── interfaces/  # 游戏接口定义 (Game, BotPlayer, TimeoutHandler)
│   │   │   ├── simple/      # 简易游戏实现
│   │   │   ├── ddz/         # 斗地主游戏实现
│   │   │   ├── mahjong/     # 麻将游戏实现
│   │   │   ├── durian/      # 榴莲忘返游戏实现
│   │   │   └── factory.go   # 游戏工厂
│   │   ├── room/            # 房间管理
│   │   │   ├── room.go      # 房间实例
│   │   │   ├── manager.go   # 房间管理器
│   │   │   ├── bot.go       # 机器人管理
│   │   │   └── timer.go     # 超时计时器
│   │   └── types/           # 类型定义
│   │       ├── message.go   # 消息类型
│   │       ├── broadcast.go # 广播类型
│   │       ├── error.go     # 错误类型
│   │       └── room_action.go
│   ├── main.go              # 入口文件
│   ├── go.mod
│   └── go.sum
├── frontend/                # React 前端
│   ├── src/
│   │   ├── components/      # 通用组件 (chat, common, room, layout)
│   │   ├── contexts/        # WebSocket Context
│   │   ├── games/           # 游戏组件
│   │   │   ├── ddz/         # 斗地主 UI
│   │   │   ├── mahjong/     # 麻将 UI
│   │   │   └── durian/      # 榴莲忘返 UI
│   │   ├── hooks/           # 自定义 Hooks
│   │   ├── pages/           # 页面 (Connect, Lobby, Room)
│   │   ├── services/        # WebSocket 服务、API
│   │   ├── stores/          # Zustand 状态管理
│   │   └── types/           # TypeScript 类型定义
│   ├── package.json
│   └── index.html
├── PROTOCOL.md              # WebSocket 协议文档
└── README.md
```

## 快速开始

### 启动后端

```bash
cd card-game-server/backend
go mod download
go run main.go
```

服务器将在 `http://localhost:8080` 启动，WebSocket 端点为 `/ws`。

### 启动前端

```bash
cd card-game-server/frontend
npm install
npm run dev
```

## WebSocket 协议

### 连接方式

```
ws://localhost:8080/ws?player=玩家ID&nickname=昵称
```

### 消息类型

#### 客户端 → 服务器

| 消息类型 | 说明 | 示例 |
|---------|------|------|
| `room.create` | 创建房间 | `{"type":"room.create","data":{"game_type":"ddz"}}` |
| `room.join` | 加入房间 | `{"type":"room.join","data":{"room_id":"abc123"}}` |
| `room.leave` | 离开房间 | `{"type":"room.leave","room_id":"abc123"}` |
| `room.action` | 房间操作（准备/选座/加机器人） | `{"type":"room.action","room_id":"abc123","data":{"action":"ready","data":{"ready":true}}}` |
| `game.action` | 游戏操作 | `{"type":"game.action","room_id":"abc123","data":{"action":"play","card":"..."}}` |
| `chat` | 发送聊天消息 | `{"type":"chat","room_id":"abc123","data":{"content":"你好"}}` |

#### 服务器 → 客户端

| 消息类型 | 说明 |
|---------|------|
| `broadcast` | 广播消息（游戏状态变更、聊天等） |
| `error` | 错误消息 |

详见 [PROTOCOL.md](PROTOCOL.md)。

## 游戏类型

### 1. 简易游戏 (simple)
- 2 人对战
- 简单的卡牌对战逻辑，适合测试

### 2. 斗地主 (ddz)
- 3 人对战（逆时针出牌顺序）
- 经典斗地主规则，支持叫分、出牌、炸弹等
- 支持机器人、回合超时自动操作

### 3. 麻将 (mahjong)
- 4 人对战
- 国标麻将简化版（30种番种）
- 支持吃、碰、杠、胡
- 支持回合超时自动操作

### 4. 榴莲忘返 (durian)
- 2~7 人对战
- 水果订单卡牌游戏，含猩猩兄妹特殊牌
- 猩猩牌效果：南茜(香蕉库存无限)、米奇(去除数量为3的订单)、墨菲(无事发生)
- 翻到猩猩牌触发交换订单效果（每张牌只能翻转一次）
- 支持机器人、回合超时自动操作

## 核心接口

### 游戏接口 (Game)

```go
type Game interface {
    ID() string
    Init(players []string) error
    ProcessAction(playerID string, action interface{}) (bool, error)
    CurrentTurn() string
    GetState() interface{}
    GetStateForPlayer(playerID string) interface{}
    IsGameOver() bool
    Winner() string
    MaxPlayers() int
    MinPlayers() int
}
```

### 可选接口

- `BotPlayer`：机器人行为接口，实现 `GetBotAction(botID string) *Action`
- `TimeoutHandler`：超时处理接口，实现 `HandleTurnTimeout` / `HandlePendingTimeout`

### 添加新游戏

1. 在 `backend/internal/game/` 下创建新游戏目录
2. 实现 `interfaces.Game` 接口（可选实现 `BotPlayer`、`TimeoutHandler`）
3. 在 `factory.go` 中注册游戏类型

```go
// factory.go
switch gameType {
case "your_game":
    return yourgame.New(), nil
}
```

## 开发计划

### v1.0.0
- [x] 基础 WebSocket 通信
- [x] 房间管理系统
- [x] 支持多种游戏
- [x] 支持断线重连
- [x] 支持心跳检测
- [x] 简易游戏
- [x] 斗地主游戏
- [x] 麻将游戏
- [x] 榴莲忘返游戏
- [x] 重连时获取当前的 gameState
- [ ] 消息有序

### v1.1.0
- [x] 游戏大厅
- [x] 可见的房间列表
- [x] 创建房间支持选择游戏类型
- [x] 斗地主增加机器人
- [x] 斗地主增加超时操作
- [x] React 前端重构
- [x] 支持离开房间（游玩时离开立即结束游戏）
- [x] 麻将游戏 UI 重构
- [x] 榴莲忘返游戏 UI 重构
- [x] 断线重连机制优化

### v1.2.0
- [x] 会话持久化（刷新页面自动重连，不丢失身份）
- [x] 优化超时逻辑，所有操作都需要有超时，超时后会自动操作，尽量避免游戏卡死
- [x] 聊天消息重构展示，使用透明浮层
- [ ] 持久化用户信息
- [ ] 支持微信登录
- [ ] 玩家积分系统与排行榜
- [ ] 游戏回放/观战模式
- [x] 游戏内表情包快捷发送
- [ ] 房间密码/私密房间

### v1.3.0
- [ ] 音效与动画增强（出牌音效、结算动画）
- [ ] 移动端适配优化

## 许可证

MIT License
