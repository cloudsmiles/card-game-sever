// WebSocket消息类型定义
export interface WSMessage {
  type: string;
  room_id?: string;
  data: any;
}

// 游戏类型枚举
export type GameType = 'simple' | 'ddz' | 'mahjong' | 'durian';

// 房间信息
export interface RoomInfo {
  id: string;
  game_type: GameType;
  players: PlayerInfo[];
  max_players: number;
  status: 'waiting' | 'playing' | 'gameover';
  created_at: string;
}

// 玩家信息
export interface PlayerInfo {
  id: string;
  name: string;
  is_ready: boolean;
  seat_index?: number;
}

// WebSocket消息类型常量
export const MessageTypes = {
  // 客户端发送
  ROOM_CREATE: 'room.create',
  ROOM_JOIN: 'room.join',
  ROOM_ACTION: 'room.action',
  GAME_ACTION: 'game.action',
  CHAT: 'chat',
  
  // 服务器发送
  BROADCAST: 'broadcast',
  ERROR: 'error'
} as const;

// 广播事件类型
export const BroadcastEvents = {
  ROOM_STATE_CHANGED: 'room_state_changed',
  CHAT_MESSAGE: 'chat_message',
  GAME_STATE_UPDATE: 'game_state_update',
  PLAYER_JOINED: 'player_joined',
  PLAYER_LEFT: 'player_left'
} as const;

export type BroadcastEvent = typeof BroadcastEvents[keyof typeof BroadcastEvents];

// 房间操作类型
export const RoomActions = {
  READY: 'ready',
  UNREADY: 'unready',
  SELECT_SEAT: 'select_seat'
} as const;

// 创建房间数据
export interface CreateRoomData {
  game_type: GameType;
}

// 加入房间数据
export interface JoinRoomData {
  room_id: string;
}

// 房间操作数据
export interface RoomActionData {
  action: keyof typeof RoomActions;
  data?: any;
}

// 游戏操作数据
export interface GameActionData {
  action: string;
  data?: any;
}

// 聊天数据
export interface ChatData {
  content: string;
}

// 错误数据
export interface ErrorData {
  code: string;
  message: string;
}

// 广播数据
export interface BroadcastData {
  event: BroadcastEvent;
  content: any;
}