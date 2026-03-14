import React, { createContext, useContext, useEffect, useState } from 'react';
import { wsService, WebSocketService } from '@/services/websocket';
import { useConnectionStore } from '@/stores/connectionStore';
import { useUserStore } from '@/stores/userStore';
import { useRoomStore } from '@/stores/roomStore';
import { useGameStore } from '@/stores/gameStore';
import { useUIStore } from '@/stores/uiStore';
import { useChatStore } from '@/stores/chatStore';

interface WebSocketContextValue {
  service: WebSocketService;
  isConnected: boolean;
  connect: (playerId: string, nickname: string) => Promise<void>;
  disconnect: () => void;
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false);
  
  const connectionStore = useConnectionStore();
  const userStore = useUserStore();
  const roomStore = useRoomStore();
  const gameStore = useGameStore();
  const uiStore = useUIStore();
  const chatStore = useChatStore();

  useEffect(() => {
    // Handle connection events
    wsService.on('connected', () => {
      setIsConnected(true);
      connectionStore.setStatus('connected');
      connectionStore.resetReconnectAttempts();
    });

    wsService.on('disconnected', () => {
      setIsConnected(false);
      connectionStore.setStatus('disconnected');
    });

    wsService.on('reconnecting', () => {
      connectionStore.setStatus('reconnecting');
      connectionStore.incrementReconnectAttempts();
    });

    wsService.on('error', (error: any) => {
      console.error('[WebSocket] Error:', error);
      connectionStore.setError(error.message || 'Connection error');
      uiStore.addNotification({
        type: 'error',
        message: error.message || 'WebSocket connection error',
      });
    });

    // Handle room state changes
    wsService.on('room_state_changed', (data: any) => {
      console.log('[WebSocket] Room state changed:', data);
      
      const maxPlayers = data.game_type === 'ddz' ? 3 : data.game_type === 'mahjong' ? 4 : 4;
      
      roomStore.setRoom(data.room_id, data.game_type, maxPlayers);
      roomStore.updateRoomStatus(data.state);
      roomStore.updatePlayers(
        data.players.map((p: any) => ({
          seat: p.seat_number,
          player_id: p.player_id,
          nickname: p.nickname || p.player_id,
          ready: p.ready,
          offline: p.is_offline,
          is_bot: p.is_bot || false,
        }))
      );

      if (data.message) {
        uiStore.addNotification({
          type: 'info',
          message: data.message,
        });
      }
    });

    // Handle game state updates
    wsService.on('state_update', (data: any) => {
      console.log('[WebSocket] Game state update:', data);
      gameStore.setGameData(data);
    });

    // Handle game started
    wsService.on('game_started', (data: any) => {
      console.log('[WebSocket] Game started:', data);
      roomStore.updateRoomStatus('playing');
      // The game_started event contains the initial game state
      if (data && Object.keys(data).length > 0) {
        gameStore.setGameData(data);
      }
      
      uiStore.addNotification({
        type: 'success',
        message: '游戏开始！',
      });
    });

    // Handle game over
    wsService.on('game_over', (data: any) => {
      console.log('[WebSocket] Game over:', data);
      roomStore.updateRoomStatus('waiting');
      
      if (data.message) {
        uiStore.addNotification({
          type: 'info',
          message: data.message,
        });
      }
    });

    // Handle room left confirmation
    wsService.on('room_left', () => {
      console.log('[WebSocket] Room left confirmed');
      roomStore.leaveRoom();
      gameStore.resetGame();
      chatStore.clearMessages();
    });

    // Handle chat messages (backend broadcasts event as "chat")
    wsService.on('chat', (data: any) => {
      console.log('[WebSocket] Chat message:', data);
      chatStore.addMessage({
        playerId: data.player_id,
        nickname: data.nickname || data.player_id,
        content: data.content,
      });
    });

    return () => {
      wsService.disconnect();
    };
  }, []);

  const connect = async (playerId: string, nickname: string): Promise<void> => {
    try {
      connectionStore.setStatus('connecting');
      connectionStore.setPlayer(playerId, nickname);
      userStore.setUser(playerId, nickname);

      // Clear stale room/game state from any previous session
      roomStore.leaveRoom();
      gameStore.resetGame();
      chatStore.clearMessages();
      
      await wsService.connect(playerId, nickname);
    } catch (error: any) {
      connectionStore.setStatus('disconnected');
      connectionStore.setError(error.message || 'Failed to connect');
      throw error;
    }
  };

  const disconnect = (): void => {
    wsService.disconnect();
    connectionStore.reset();
    userStore.clearUser();
    roomStore.leaveRoom();
    gameStore.resetGame();
    chatStore.clearMessages();
  };

  return (
    <WebSocketContext.Provider value={{ service: wsService, isConnected, connect, disconnect }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocketContext = (): WebSocketContextValue => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocketContext must be used within WebSocketProvider');
  }
  return context;
};
