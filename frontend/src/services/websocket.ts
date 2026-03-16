import type { WSMessage } from '@/types/websocket';

type EventHandler = (data: any) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private eventHandlers: Map<string, Set<EventHandler>> = new Map();
  private reconnectTimer: number | null = null;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private baseReconnectDelay: number = 3000;
  private playerId: string | null = null;
  private nickname: string | null = null;
  private url: string = '';
  private shouldReconnect: boolean = true;
  private isConnecting: boolean = false;

  /**
   * Connect to WebSocket server
   */
  connect(playerId: string, nickname?: string): Promise<void> {
    // 防止并发连接
    if (this.isConnecting) {
      return Promise.reject(new Error('Already connecting'));
    }

    // 关闭旧连接（不触发自动重连）
    if (this.ws) {
      const oldWs = this.ws;
      oldWs.onclose = null; // 移除 onclose 防止触发重连
      oldWs.onerror = null;
      oldWs.onmessage = null;
      oldWs.close();
      this.ws = null;
    }

    // 清除待执行的重连定时器
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    this.isConnecting = true;
    this.shouldReconnect = true;
    this.playerId = playerId;
    this.nickname = nickname || null;

    const wsUrl = (import.meta as any).env?.VITE_WS_URL || 'ws://localhost:8080/ws';
    const params = new URLSearchParams({ player: playerId });
    if (nickname) {
      params.set('nickname', nickname);
    }
    this.url = `${wsUrl}?${params.toString()}`;

    console.log('[WebSocket] Connecting to:', this.url);

    return new Promise((resolve, reject) => {
      try {
        const ws = new WebSocket(this.url);

        ws.onopen = () => {
          console.log('[WebSocket] Connected');
          this.ws = ws;
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.emit('connected', null);
          resolve();
        };

        ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };

        ws.onerror = (error) => {
          console.error('[WebSocket] Error:', error);
          this.isConnecting = false;
          this.emit('error', error);
          reject(error);
        };

        ws.onclose = () => {
          console.log('[WebSocket] Disconnected');
          // 只有当前活跃的 ws 关闭时才处理
          if (this.ws === ws) {
            this.ws = null;
            this.emit('disconnected', null);
            if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
              this.scheduleReconnect();
            }
          }
        };
      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  /**
   * Disconnect from server
   */
  disconnect(): void {
    this.shouldReconnect = false;
    this.isConnecting = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.onclose = null;
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Send message to server
   */
  send(message: WSMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.error('[WebSocket] Cannot send message: not connected');
    }
  }

  on(event: string, handler: EventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  off(event: string, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  private emit(event: string, data: any): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach((handler) => handler(data));
    }
  }

  private handleMessage(data: string): void {
    try {
      const message = JSON.parse(data);

      if (message.type === 'broadcast') {
        const broadcast = message.data as { event: string; content: any };
        const { event, content } = broadcast;
        this.emit(event, content);
        this.emit('broadcast', { event, content });
      } else if (message.type === 'error') {
        const errorData = message.data as { code: string; message: string };
        this.emit('error', errorData);
      } else {
        this.emit(message.type, message.data);
      }
    } catch (error) {
      console.error('[WebSocket] Failed to parse message:', error);
    }
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    if (!this.playerId || this.isConnecting) return;

    this.reconnectAttempts++;
    const delay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(
      `[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    this.emit('reconnecting', { attempt: this.reconnectAttempts, delay });

    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      this.connect(this.playerId!, this.nickname || undefined)
        .then(() => {
          console.log('[WebSocket] Reconnected successfully');
        })
        .catch((error) => {
          console.error('[WebSocket] Reconnection failed:', error);
        });
    }, delay);
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
export const wsService = new WebSocketService();
