// Mahjong Game Types

export type TileSuit = 'wan' | 'tiao' | 'tong' | 'zi';

export interface MahjongTile {
  suit: TileSuit;
  rank: number;
}

export interface MeldData {
  type: string; // '吃' | '碰' | '明杠' | '暗杠' | '补杠'
  tiles: MahjongTile[];
  from_player: number;
}

export interface MahjongPlayerInfo {
  player_id: string;
  seat: number;
  hand_count: number;
  melds: MeldData[];
  discards: MahjongTile[];
  is_dealer: boolean;
}

export interface KongOption {
  tile: MahjongTile;
  type: 'concealed' | 'extended';
}

export interface FanItem {
  name: string;
  fan: number;
}

export interface WinResult {
  winner_seat: number;
  win_type: string;
  loser_seat: number;
  fan_list: FanItem[];
  total_fan: number;
}

export interface MahjongServerState {
  phase: 'play' | 'pending';
  current_seat: number;
  current: string;
  dealer_seat: number;
  wall_remaining: number;
  players: MahjongPlayerInfo[];
  game_over: boolean;
  winner: string;
  my_seat: number;
  my_hand: MahjongTile[];
  available_actions?: string[];
  last_discard?: MahjongTile;
  last_discard_seat?: number;
  chow_options?: MahjongTile[][];
  kong_options?: KongOption[];
  win_result?: WinResult;
  turn_deadline: number;    // Unix毫秒时间戳
  pending_deadline: number; // Unix毫秒时间戳
}

// Tile display helpers
const suitNames: Record<TileSuit, string> = {
  wan: '万',
  tiao: '条',
  tong: '筒',
  zi: '',
};

const ziNames: Record<number, string> = {
  1: '东', 2: '南', 3: '西', 4: '北',
  5: '中', 6: '发', 7: '白',
};

export const tileToString = (tile: MahjongTile): string => {
  if (tile.suit === 'zi') return ziNames[tile.rank] || '?';
  return `${tile.rank}${suitNames[tile.suit]}`;
};

export const tileToEmoji = (tile: MahjongTile): string => {
  if (tile.suit === 'zi') return ziNames[tile.rank] || '?';
  return `${tile.rank}`;
};

export const suitLabel = (suit: TileSuit): string => suitNames[suit];

export const isSameTile = (a: MahjongTile, b: MahjongTile): boolean =>
  a.suit === b.suit && a.rank === b.rank;
