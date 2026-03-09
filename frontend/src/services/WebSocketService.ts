import { 
  WSMessage, 
  MessageTypes, 
  BroadcastEvents,
  CreateRoomData,
  JoinRoomData,
  RoomActionData,
  GameActionData,
  ChatData,
  BroadcastData,
  RoomInfo,
  PlayerInfo
} from '../types/websocket';

class WebSocketService {
  private ws: WebSocket | null = null;
  private playerID: string = '';
  private currentRoomID: string = '';
  private listeners: Map<string, ((data: any) => void)[]> = new Map();
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectDelay: number = 3000;

  // 连接到WebSocket服务器
  connect(playerName: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.playerID = playerName;
      
      // 使用相对路径，让Vite代理处理
      const wsUrl = `/ws?player=${encodeURIComponent(playerName)}`;

      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket连接已建立');
        this.reconnectAttempts = 0;
        resolve();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('解析WebSocket消息失败:', error);
        }
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket连接已关闭:', event.code, event.reason);
        this.handleDisconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket连接错误:', error);
        reject(error);
      };
    });
  }

  // 处理接收到的消息
  private handleMessage(message: WSMessage) {
    console.log('收到WebSocket消息:', message);

    switch (message.type) {
      case MessageTypes.BROADCAST:
        this.handleBroadcast(message.data as BroadcastData);
        break;
      case MessageTypes.ERROR:
        this.handleError(message.data);
        break;
      default:
        // 其他类型的消息可以通过事件系统处理
        this.emit('message', message);
    }
  }

  // 处理广播消息
  private handleBroadcast(data: BroadcastData) {
    console.log('处理广播消息:', data.event, data.content);
    
    switch (data.event) {
      case BroadcastEvents.ROOM_STATE_CHANGED:
        this.emit('roomStateChanged', data.content);
        break;
      case BroadcastEvents.CHAT_MESSAGE:
        this.emit('chatMessage', data.content);
        break;
      case BroadcastEvents.GAME_STATE_UPDATE:
        this.emit('gameStateUpdate', data.content);
        break;
      case BroadcastEvents.PLAYER_JOINED:
        this.emit('playerJoined', data.content);
        break;
      case BroadcastEvents.PLAYER_LEFT:
        this.emit('playerLeft', data.content);
        break;
      default:
        this.emit('broadcast', data);
    }
  }

  // 处理错误消息
  private handleError(error: any) {
    console.error('WebSocket错误:', error);
    this.emit('error', error);
  }

  // 处理连接断开
  private handleDisconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`尝试重新连接 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      
      setTimeout(() => {
        this.connect(this.playerID).catch(err => {
          console.error('重连失败:', err);
        });
      }, this.reconnectDelay);
    } else {
      console.error('达到最大重连次数，停止重连');
      this.emit('disconnected', { reason: '达到最大重连次数' });
    }
  }

  // 发送消息
  private sendMessage(message: WSMessage) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.error('WebSocket未连接，无法发送消息');
      throw new Error('WebSocket未连接');
    }
  }

  // 创建房间
  createRoom(gameType: string) {
    const message: WSMessage = {
      type: MessageTypes.ROOM_CREATE,
      data: {
        game_type: gameType
      } as CreateRoomData
    };
    this.sendMessage(message);
  }

  // 加入房间
  joinRoom(roomID: string) {
    const message: WSMessage = {
      type: MessageTypes.ROOM_JOIN,
      data: {
        room_id: roomID
      } as JoinRoomData
    };
    this.sendMessage(message);
  }

  // 房间操作
  roomAction(action: string, data?: any) {
    const message: WSMessage = {
      type: MessageTypes.ROOM_ACTION,
      room_id: this.currentRoomID,
      data: {
        action,
        data
      } as RoomActionData
    };
    this.sendMessage(message);
  }

  // 游戏操作
  gameAction(action: string, data?: any) {
    const message: WSMessage = {
      type: MessageTypes.GAME_ACTION,
      room_id: this.currentRoomID,
      data: {
        action,
        data
      } as GameActionData
    };
    this.sendMessage(message);
  }

  // 发送聊天消息
  sendChat(content: string) {
    const message: WSMessage = {
      type: MessageTypes.CHAT,
      room_id: this.currentRoomID,
      data: {
        content
      } as ChatData
    };
    this.sendMessage(message);
  }

  // 设置当前房间ID
  setCurrentRoomID(roomID: string) {
    this.currentRoomID = roomID;
  }

  // 获取当前房间ID
  getCurrentRoomID(): string {
    return this.currentRoomID;
  }

  // 获取玩家ID
  getPlayerID(): string {
    return this.playerID;
  }

  // 关闭连接
  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.listeners.clear();
  }

  // 事件监听
  on(event: string, callback: (data: any) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event)?.push(callback);
  }

  // 移除事件监听
  off(event: string, callback: (data: any) => void) {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    }
  }

  // 触发事件
  private emit(event: string, data: any) {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`事件 ${event} 的回调函数执行出错:`, error);
        }
      });
    }
  }

  // 检查连接状态
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

// 导出单例实例
export const wsService = new WebSocketService();

// Hook用于在组件中使用WebSocket服务
export const useWebSocket = () => {
  return wsService;
};