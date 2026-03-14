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
  private url: string = '';
  private shouldReconnect: boolean = true;

  /**
   * Connect to WebSocket server
   */
  connect(playerId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.playerId = playerId;
      const wsUrl = (import.meta as any).env?.VITE_WS_URL || 'ws://localhost:8080/ws';
      this.url = `${wsUrl}?player=${playerId}`;
      
      console.log('[WebSocket] Connecting to:', this.url);

      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('[WebSocket] Connected');
          this.reconnectAttempts = 0;
          this.emit('connected', null);
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };

        this.ws.onerror = (error) => {
          console.error('[WebSocket] Error:', error);
          this.emit('error', error);
          reject(error);
        };

        this.ws.onclose = () => {
          console.log('[WebSocket] Disconnected');
          this.emit('disconnected', null);
          this.ws = null;

          if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnect();
          }
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Disconnect from server
   */
  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
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

  /**
   * Register event handler
   */
  on(event: string, handler: EventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  /**
   * Unregister event handler
   */
  off(event: string, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  /**
   * Emit event to handlers
   */
  private emit(event: string, data: any): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach((handler) => handler(data));
    }
  }

  /**
   * Handle incoming message
   */
  private handleMessage(data: string): void {
    try {
      const message = JSON.parse(data);

      if (message.type === 'broadcast') {
        const broadcast = message.data as { event: string; content: any };
        const { event, content } = broadcast;
        
        // Emit specific event
        this.emit(event, content);
        
        // Also emit generic broadcast event
        this.emit('broadcast', { event, content });
      } else if (message.type === 'error') {
        const errorData = message.data as { code: string; message: string };
        this.emit('error', errorData);
      } else {
        // Handle other message types
        this.emit(message.type, message.data);
      }
    } catch (error) {
      console.error('[WebSocket] Failed to parse message:', error);
    }
  }

  /**
   * Handle reconnection with exponential backoff
   */
  private reconnect(): void {
    if (!this.playerId) {
      console.error('[WebSocket] Cannot reconnect: no player ID');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.getReconnectDelay();

    console.log(
      `[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    this.emit('reconnecting', { attempt: this.reconnectAttempts, delay });

    this.reconnectTimer = window.setTimeout(() => {
      this.connect(this.playerId!)
        .then(() => {
          console.log('[WebSocket] Reconnected successfully');
        })
        .catch((error) => {
          console.error('[WebSocket] Reconnection failed:', error);
        });
    }, delay);
  }

  /**
   * Calculate reconnect delay with exponential backoff
   */
  private getReconnectDelay(): number {
    return this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
export const wsService = new WebSocketService();
