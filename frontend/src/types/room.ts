// Room-related types

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

export interface RoomInfo {
  room_id: string;
  game_type: GameType;
  status: RoomStatus;
  players: SeatPlayerInfo[];
  player_count: number;
  max_players: number;
}
