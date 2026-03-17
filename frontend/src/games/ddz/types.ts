// DDZ (斗地主) Game Types

export type Suit = 'hearts' | 'diamonds' | 'clubs' | 'spades' | 'joker';
export type Rank = '3' | '4' | '5' | '6' | '7' | '8' | '9' | '10' | 'J' | 'Q' | 'K' | 'A' | '2' | 'JOKER' | 'joker';

export interface Card {
  suit: Suit;
  rank: Rank;
  id: string;
}

export type PlayerPosition = 'left' | 'right' | 'bottom';

export interface DDZPlayerInfo {
  player_id: string;
  nickname?: string;
  seat_number: number;
  is_landlord: boolean;
  hand_count: number;
}

// Backend card object with suit info
export interface CardObj {
  Value: number;
  Suit: string; // "1"=♦, "2"=♣, "3"=♥, "4"=♠, "小王"/"大王"
}

export interface LastPlay {
  Type: string;
  Cards: CardObj[];
}

export interface PlayerAction {
  type: 'play' | 'pass' | 'call';
  cards?: CardObj[];
  score?: number;
}

// Backend state_update structure
export interface DDZServerState {
  phase: 'call' | 'play';
  players: string[];
  player_infos: DDZPlayerInfo[];
  player_seats: Record<string, number>;
  landlord: string;
  current: string;
  current_seat: number;
  last_play: LastPlay;
  last_actions: Record<string, PlayerAction>;
  game_over: boolean;
  winner: string;
  hand_counts: Record<string, number>;
  bottom: CardObj[] | null;
  my_hand: CardObj[];
  turn_deadline: number; // Unix毫秒时间戳
}

export interface DDZAction {
  type: 'call_landlord' | 'play_cards' | 'pass';
  data?: number | Array<{ value: number }>;
}

// Card value to display mapping
export const VALUE_DISPLAY: Record<number, { rank: Rank; suit: Suit }> = {};

// Value to rank string
export const valueToRank = (value: number): string => {
  if (value <= 10) return String(value);
  if (value === 11) return 'J';
  if (value === 12) return 'Q';
  if (value === 13) return 'K';
  if (value === 14) return 'A';
  if (value === 15) return '2';
  if (value === 16) return '小王';
  if (value === 17) return '大王';
  return String(value);
};

// Check if value is a red card (for display color)
export const isRedValue = (value: number): boolean => {
  return value === 16 || value === 17; // Jokers are special
};
