// Common types used across the application

export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
  duration?: number;
}

export interface GameLog {
  id: string;
  timestamp: number;
  message: string;
  playerId?: string;
}
