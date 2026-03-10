# 多人卡牌游戏通信协议

## 概述

本文档定义了前端与后端之间的 WebSocket 通信协议。

## 连接建立

```
ws://localhost:8080/ws?player=PLAYER_ID
```

## 消息格式

### 统一消息结构

```typescript
interface WSMessage {
  type: string;           // 消息类型
  room_id?: string;       // 房间ID（部分消息需要）
  player_id?: string;    // 玩家ID（通过URL参数传递）
  data: any;             // 消息数据
}
```

---

## 客户端 -> 服务器消息

### 1. 创建房间 (room.create)

```typescript
// 请求
{
  type: 'room.create',
  data: {
    game_type: 'simple' | 'ddz' | 'mahjong' | 'durian'
  }
}

// 服务器响应：通过 room_state_changed 广播房间状态
```

### 2. 加入房间 (room.join)

```typescript
// 请求
{
  type: 'room.join',
  data: {
    room_id: string
  }
}

// 服务器响应：通过 room_state_changed 广播房间状态
```

### 3. 房间操作 (room.action)

```typescript
// 请求
{
  type: 'room.action',
  room_id: string,
  data: {
    action: 'ready' | 'unready' | 'select_seat',
    data?: {
      ready?: boolean,
      seat_number?: number
    }
  }
}

// 准备操作
{
  type: 'room.action',
  room_id: 'xxx',
  data: {
    action: 'ready',
    data: { ready: true }
  }
}

// 选座操作
{
  type: 'room.action',
  room_id: 'xxx',
  data: {
    action: 'select_seat',
    data: { seat_number: 0 }
  }
}
```

### 4. 游戏操作 (game.action)

```typescript
// 请求
{
  type: 'game.action',
  room_id: string,
  data: {
    action: string,      // 游戏特定动作
    card?: any           // 动作数据（打出的牌等）
  }
}

// 游戏开始
{
  type: 'game.action',
  room_id: 'xxx',
  data: { action: 'start' }
}

// 离开游戏
{
  type: 'game.action',
  room_id: 'xxx',
  data: { action: 'leave' }
}
```

### 5. 聊天 (chat)

```typescript
// 请求
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

**服务器推送时机：** 玩家加入/离开房间、准备/取消准备、选座、开始游戏

```typescript
// 内容结构
{
  room_id: string,           // 房间ID
  state: 'waiting' | 'playing', // 房间状态
  players: SeatPlayerInfo[], // 玩家列表
  message: string            // 提示信息
}

// 玩家信息
interface SeatPlayerInfo {
  player_id: string;    // 玩家ID
  seat_number: number;  // 座位号 (0, 1, 2, ...)
  ready: boolean;        // 是否已准备
  is_offline: boolean;   // 是否断线
}

// 示例
{
  type: 'broadcast',
  data: {
    event: 'room_state_changed',
    content: {
      room_id: 'room_abc123',
      state: 'waiting',
      players: [
        { player_id: 'player1', seat_number: 0, ready: true, is_offline: false },
        { player_id: 'player2', seat_number: 1, ready: false, is_offline: false }
      ],
      message: '玩家 player1 加入了房间'
    }
  }
}
```

### 2. 游戏状态更新 (state_update)

**服务器推送时机：** 游戏中有任何动作（出牌、摸牌等）

```typescript
// 内容为游戏特定的状态对象
// 不同游戏类型有不同的结构，参见各游戏协议
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
      winner: string,        // 获胜者ID
      message: string        // 提示信息
      // 游戏特定的结算信息
    }
  }
}
```

### 5. 聊天消息 (chat_message)

```typescript
{
  type: 'broadcast',
  data: {
    event: 'chat_message',
    content: {
      room_id: string,
      player_id: string,
      content: string
    }
  }
}
```

### 6. 错误消息 (error)

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
| 400 | 消息格式错误 |
| 401 | 缺少消息类型 |
| 1001 | 玩家已在其他房间 |
| 1002 | 房间不存在 |
| 1003 | 加入房间失败 |
| 1004 | 房间已满 |
| 1005 | 游戏进行中，无法加入 |
| 500 | 服务器内部错误 |

---

## 前端类型定义

```typescript
// types/websocket.ts

// 游戏类型
export type GameType = 'simple' | 'ddz' | 'mahjong' | 'durian';

// 玩家座位信息
export interface SeatPlayerInfo {
  player_id: string;
  seat_number: number;
  ready: boolean;
  is_offline: boolean;
}

// 房间状态变更内容
export interface RoomStateContent {
  room_id: string;
  state: 'waiting' | 'playing';
  players: SeatPlayerInfo[];
  message: string;
}

// WebSocket 消息
export interface WSMessage {
  type: string;
  room_id?: string;
  player_id?: string;
  data: any;
}

// 消息类型常量
export const MessageTypes = {
  ROOM_CREATE: 'room.create',
  ROOM_JOIN: 'room.join',
  ROOM_ACTION: 'room.action',
  GAME_ACTION: 'game.action',
  CHAT: 'chat',
  BROADCAST: 'broadcast',
  ERROR: 'error'
} as const;

// 广播事件常量
export const BroadcastEvents = {
  ROOM_STATE_CHANGED: 'room_state_changed',
  CHAT_MESSAGE: 'chat_message',
  GAME_STATE_UPDATE: 'state_update',
  GAME_STARTED: 'game_started',
  GAME_OVER: 'game_over',
  PLAYER_JOINED: 'player_joined',
  PLAYER_LEFT: 'player_left'
} as const;
```

---

## 处理 room_state_changed 示例

前端处理 `room_state_changed` 事件时，直接使用 `content` 字段：

```typescript
wsService.on('roomStateChanged', (data: RoomStateContent) => {
  // data 直接就是 RoomStateContent 结构
  console.log('房间ID:', data.room_id);
  console.log('房间状态:', data.state);
  console.log('玩家列表:', data.players);
  console.log('提示信息:', data.message);

  // 更新房间信息
  setRoomInfo({
    id: data.room_id,
    game_type: currentGameType,
    players: data.players.map(p => ({
      id: p.player_id,
      name: p.player_id, // 可选：如有昵称则使用昵称
      is_ready: p.ready,
      seat_index: p.seat_number
    })),
    max_players: data.players.length, // 或根据游戏类型确定
    status: data.state,
    created_at: ''
  });
});
```