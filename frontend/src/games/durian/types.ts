// Durian (榴莲忘返) Game Types

export type FruitType = 'durian' | 'banana' | 'grape' | 'strawberry';
export type CardType = 'fruit' | 'gorilla';

export interface FruitCardData {
  id: number;
  card_type: 'fruit';
  left_fruit: FruitType;
  left_count: number;
  right_fruit: FruitType;
  right_count: number;
}

export interface GorillaCardData {
  id: number;
  card_type: 'gorilla';
  ability: string;
  description: string;
}

export type CardData = FruitCardData | GorillaCardData;

export interface PlayerInfo {
  player_id: string;
  anger_score: number;
  anger_tokens: number[];
}

export interface OtherPlayerCard {
  player_id: string;
  card: CardData;
}

export interface PlayedCardRecord {
  player_name: string;
  card: CardData;
  chosen_side: 'left' | 'right';
  timestamp: string;
  round_number: number;
  swapped?: boolean;
}

export interface GorillaEffect {
  player_id: string;
  ability: string;
  cancelled_orders: number;
}

export interface HolderCardInfo {
  player_id: string;
  card: CardData;
}

export interface SettlementData {
  bell_ringer: string;
  all_holder_cards: HolderCardInfo[];
  gorilla_effects: GorillaEffect[];
  orders_before_cancel: Record<string, number>;
  orders_after_cancel: Record<string, number>;
  inventory: Record<string, number>;
  is_shortage: boolean;
  shortage_fruits: string[];
  punished_player: string;
  punish_reason: string;
  anger_token_given: number;
}

export interface DurianServerState {
  phase: 'dealing' | 'playing' | 'settlement' | 'round_end' | 'waiting_continue' | 'game_over';
  round_number: number;
  current: string;
  deck_remaining: number;
  orders: Record<string, number>;
  anger_token_pool: number[];
  players: PlayerInfo[];
  game_over: boolean;
  winner: string;
  current_drawn_card: CardData | null;
  played_cards_history: PlayedCardRecord[];
  waiting_for_swap_order: boolean;
  your_card: null; // always null (can't see own card)
  others_cards: OtherPlayerCard[];
  last_settlement?: SettlementData;
  waiting_for_continue?: string[];
  turnDeadline?: number;
}

// Fruit display helpers
export const fruitEmoji: Record<FruitType, string> = {
  durian: '🥭',
  banana: '🍌',
  grape: '🍇',
  strawberry: '🍓',
};

export const fruitName: Record<FruitType, string> = {
  durian: '榴莲',
  banana: '香蕉',
  grape: '葡萄',
  strawberry: '草莓',
};

export const fruitColor: Record<FruitType, string> = {
  durian: '#F7931E',
  banana: '#FFD54F',
  grape: '#7B1FA2',
  strawberry: '#E53935',
};
