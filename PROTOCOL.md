# 多人卡牌游戏通信协议

## 概述

本文档定义了前端与后端之间的 WebSocket 通信协议。

## 连接建立

```
ws://localhost:8080/ws?player=PLAYER_ID&nickname=昵称
```

- `player`: 玩家唯一ID（必填）
- `nickname`: 玩家昵称（可选，默认使用 player ID）

## 消息格式

### 统一消息结构

```typescript
interface WSMessage {
  type: string;           // 消息类型
  room_id?: string;       // 房间ID（部分消息需要）
  player_id?: string;     // 玩家ID（服务端填充，通过URL参数传递）
  data: any;              // 消息数据
}
```

---

## 客户端 -> 服务器消息

### 1. 创建房间 (room.create)

```typescript
{
  type: 'room.create',
  data: {
    game_type: 'simple' | 'ddz' | 'mahjong' | 'durian'
  }
}
```

### 2. 加入房间 (room.join)

```typescript
{
  type: 'room.join',
  data: {
    room_id: string
  }
}
```

### 3. 离开房间 (room.leave)

```typescript
{
  type: 'room.leave',
  room_id: string,
  data: {}
}
```

### 4. 房间操作 (room.action)

```typescript
{
  type: 'room.action',
  room_id: string,
  data: {
    action: 'ready' | 'select_seat' | 'add_bot' | 'kick_bot',
    data?: any
  }
}
```

#### 准备/取消准备

```typescript
{
  type: 'room.action',
  room_id: 'xxx',
  data: {
    action: 'ready',
    data: { ready: true }   // true=准备, false=取消准备
  }
}
```

#### 选座

```typescript
{
  type: 'room.action',
  room_id: 'xxx',
  data: {
    action: 'select_seat',
    data: { seat_number: 0 }
  }
}
```

#### 添加机器人

```typescript
{
  type: 'room.action',
  room_id: 'xxx',
  data: {
    action: 'add_bot'
  }
}
```

#### 踢出机器人

```typescript
{
  type: 'room.action',
  room_id: 'xxx',
  data: {
    action: 'kick_bot',
    data: { bot_id: 'bot_0' }
  }
}
```

### 5. 游戏操作 (game.action)

```typescript
{
  type: 'game.action',
  room_id: string,
  data: {
    action: string,      // 游戏特定动作
    card?: any,          // 动作数据（打出的牌等）
    [key: string]: any   // 其他游戏特定字段
  }
}
```

### 6. 聊天 (chat)

```typescript
{
  type: 'chat',
  room_id: string,
  data: {
    content: string
  }
}
```

---

## 服务器 -> 客户端消息

### 广播消息格式

所有服务器推送都使用 `broadcast` 类型：

```typescript
{
  type: 'broadcast',
  data: {
    event: string,       // 事件类型
    content: any         // 事件内容
  }
}
```

---

### 1. 房间状态变更 (room_state_changed)

**触发时机：** 玩家加入/离开房间、准备/取消准备、选座、添加/踢出机器人、开始游戏

```typescript
{
  type: 'broadcast',
  data: {
    event: 'room_state_changed',
    content: {
      room_id: string,
      game_type: string,                    // 游戏类型
      state: 'waiting' | 'playing',
      players: PlayerSeatInfo[],
      message: string
    }
  }
}

interface PlayerSeatInfo {
  player_id: string;
  nickname: string;
  seat_number: number;    // 座位号 (0, 1, 2, ...)
  ready: boolean;
  is_offline: boolean;    // 是否断线
  is_bot: boolean;        // 是否是机器人
}
```

### 2. 游戏状态更新 (state_update)

**触发时机：** 游戏中有任何动作（出牌、摸牌等）

```typescript
{
  type: 'broadcast',
  data: {
    event: 'state_update',
    content: { /* 游戏特定的状态对象，每个玩家收到个性化内容 */ }
  }
}
```

### 3. 游戏开始 (game_started)

```typescript
{
  type: 'broadcast',
  data: {
    event: 'game_started',
    content: {
      message: '游戏开始！'
    }
  }
}
```

### 4. 游戏结束 (game_over)

```typescript
{
  type: 'broadcast',
  data: {
    event: 'game_over',
    content: {
      winner: string,
      message: string
      // 游戏特定的结算信息
    }
  }
}
```

### 5. 聊天消息 (chat)

```typescript
{
  type: 'broadcast',
  data: {
    event: 'chat',
    content: {
      room_id: string,
      player_id: string,
      nickname: string,
      content: string
    }
  }
}
```

### 6. 离开房间确认 (room_left)

```typescript
{
  type: 'broadcast',
  data: {
    event: 'room_left',
    content: {
      room_id: string
    }
  }
}
```

### 7. 错误消息 (error)

```typescript
{
  type: 'error',
  data: {
    code: number,
    message: string
  }
}
```

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 消息格式错误 / 缺少消息类型 |
| 1001 | 玩家已在其他房间 |
| 1002 | 房间不存在 |
| 1003 | 加入房间失败 |
| 1004 | 房间已满 |
| 1005 | 游戏进行中，无法加入 |
| 1006 | 玩家不在房间中 |
| 500 | 服务器内部错误 |

---

## 前端类型定义

```typescript
// types/websocket.ts
export interface WSMessage {
  type: string;
  [key: string]: any;
}

export interface WSBroadcast {
  type: 'broadcast';
  event: string;
  content: any;
}

export interface WSError {
  type: 'error';
  code: number;
  message: string;
}

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';
```

```typescript
// types/room.ts
export interface SeatPlayerInfo {
  seat: number;
  player_id: string;
  nickname: string;
  ready: boolean;
  offline: boolean;
  is_bot?: boolean;
}

export type GameType = 'ddz' | 'mahjong' | 'durian';
export type RoomStatus = 'waiting' | 'playing';
```

---

## 断线重连

- 连接断开后前端自动重连（指数退避，最大间隔30秒）
- 重连时使用相同的 `player` 和 `nickname` 参数
- 如果玩家在游戏中断线，后端标记为离线状态
- 重连成功后自动恢复到游戏房间，获取当前游戏状态
- 非游戏中断线则直接移除玩家

## 心跳检测

- 服务端定期发送 Ping 帧
- 客户端自动回复 Pong 帧
- 超时未收到 Pong 则断开连接
