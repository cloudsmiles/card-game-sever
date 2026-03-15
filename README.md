# Card Game Server

一个基于 Go 和 WebSocket 的多人在线卡牌游戏服务器，支持斗地主、麻将等多种游戏模式。

## 功能特性

- **多游戏支持**：支持斗地主(Dou Dizhu)、麻将(Mahjong)、简易卡牌游戏(Simple)
- **实时通信**：基于 WebSocket 的全双工实时通信
- **房间管理**：创建/加入房间、玩家准备、选座系统
- **断线重连**：游戏中断线保护，支持 30 秒内重连恢复
- **聊天系统**：房间内实时聊天功能
- **游戏状态同步**：广播机制确保所有玩家状态一致

## 技术栈

- **语言**：Go 1.22+
- **通信协议**：WebSocket (gorilla/websocket)
- **架构模式**：模块化设计，接口抽象

## 项目结构

```
card-game-server/
├── internal/
│   ├── connection/      # WebSocket 连接管理
│   │   ├── connection.go
│   │   └── handler.go   # 消息处理器
│   ├── game/            # 游戏逻辑
│   │   ├── interfaces/  # 游戏接口定义
│   │   ├── simple/      # 简易游戏实现
│   │   ├── ddz/         # 斗地主游戏实现
│   │   ├── mahjong/     # 麻将游戏实现
│   │   └── factory.go   # 游戏工厂
│   ├── room/            # 房间管理
│   │   ├── room.go      # 房间实例
│   │   └── manager.go   # 房间管理器
│   └── types/           # 类型定义
│       ├── message.go   # 消息类型
│       ├── broadcast.go # 广播类型
│       ├── game_action.go
│       └── room_action.go
├── docs/
│   └── mahjong-prd.md   # 麻将产品需求文档
├── client.html          # 简易游戏测试客户端
├── ddz-client.html      # 斗地主测试客户端
├── mahjong-client.html  # 麻将测试客户端
├── main.go              # 入口文件
├── go.mod
└── go.sum
```

## 快速开始

### 安装依赖

```bash
cd card-game-server
go mod download
```

### 启动服务器

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动，WebSocket 端点为 `/ws`。

### 测试客户端

打开浏览器访问测试客户端：

- 简易游戏：`http://localhost:8080/client.html`
- 斗地主：`http://localhost:8080/ddz-client.html`
- 麻将：`http://localhost:8080/mahjong-client.html`

## WebSocket 协议

### 连接方式

```
ws://localhost:8080/ws?player=玩家ID
```

### 消息类型

#### 客户端 → 服务器

| 消息类型 | 说明 | 示例 |
|---------|------|------|
| `room.create` | 创建房间 | `{"type":"room.create","data":{"game_type":"ddz"}}` |
| `room.join` | 加入房间 | `{"type":"room.join","data":{"room_id":"abc123"}}` |
| `room.action` | 房间操作（准备/选座） | `{"type":"room.action","room_id":"abc123","data":{"action":"ready","data":{"ready":true}}}` |
| `game.action` | 游戏操作（出牌等） | `{"type":"game.action","room_id":"abc123","data":{"action":"play","card":"..."}}` |
| `chat` | 发送聊天消息 | `{"type":"chat","room_id":"abc123","data":{"content":"你好"}}` |

#### 服务器 → 客户端

| 消息类型 | 说明 |
|---------|------|
| `broadcast` | 广播消息（游戏状态变更、聊天等） |
| `error` | 错误消息 |

### 房间状态

- `waiting` - 等待中（可加入、可准备）
- `playing` - 游戏中
- `paused` - 游戏暂停（有玩家断线）
- `gameover` - 游戏结束

## 游戏类型

### 1. 简易游戏 (simple)

- 2 人对战
- 简单的卡牌对战逻辑
- 适合快速测试连接和房间功能

### 2. 斗地主 (ddz)

- 3 人对战
- 经典斗地主规则
- 支持叫分、出牌、炸弹等

### 3. 麻将 (mahjong)

- 4 人对战
- 国标麻将简化版（30种番种）
- 支持吃、碰、杠、胡
- 详见 [麻将 PRD](docs/mahjong-prd.md)

## 核心接口

### 游戏接口 (Game)

```go
type Game interface {
    ID() string
    Init(players []string) error
    ProcessAction(playerID string, action interface{}) (bool, error)
    CurrentTurn() string
    AdvanceTurn()
    GetState() interface{}
    GetStateForPlayer(playerID string) interface{}
    IsGameOver() bool
    Winner() string
    MaxPlayers() int
    MinPlayers() int
}
```

### 添加新游戏

1. 在 `internal/game/` 下创建新游戏目录
2. 实现 `interfaces.Game` 接口
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
    - [x] 胡了之后，不能正常游戏结束
    - [x] 增加等待时间以及超时处理
- [x] 榴莲忘返游戏
    - [x] 翻到猩猩牌处理不正确
    - [x] 点击继续游戏，状态同步有问题，会回到等待中页面
- [x] 重连时获取当前的gameState
- [ ] 消息有序

### v1.1.0
- [x] 游戏大厅
- [x] 可见的房间列表
- [x] 创建房间支持选择游戏类型
- [x] 斗地主增加机器人
- [x] 斗地主增加超时操作
- [x] 使用主流的前端游戏框架，重构游戏
- [x] 支持离开房间
    - [x] 玩家游玩时离开房间应该立即结束游戏
- [ ] 处理好断线重连机制
- [x] 麻将游戏重构
- [x] 榴莲忘贩游戏重构
- [ ] 聊天消息重构展示，使用透明窗口

### v1.2.0


## Agents 协同体系

本项目采用多Agent协同开发模式，各角色分工明确：

### 🎮 card-game-pm (产品经理)
- 负责产品规划和需求分析
- 制定游戏规则和平衡性设计
- 用户体验研究和市场调研

### 💻 card-game-backend (后端开发)
- 负责游戏逻辑和服务端架构
- 数据库设计和API开发
- 性能优化和系统稳定性

### 🎨 card-game-frontend (前端开发)
- 负责用户界面和交互体验
- 游戏客户端开发和优化
- 跨平台适配和技术选型

### 🎨 ui-artist (UI美术设计师) ⭐新加入
- 专门负责UI/UX设计和美术视觉创作
- 打造温馨童真、充满想象力的卡通风格
- 色彩体系规划和动效设计
- 视觉规范制定和组件库建设

## 许可证

MIT License
