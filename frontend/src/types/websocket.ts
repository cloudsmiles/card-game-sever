// WebSocket message types based on PROTOCOL.md

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
  code: string;
  message: string;
}

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';
