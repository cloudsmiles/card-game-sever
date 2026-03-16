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
  connect: (playerId: string, nickname: string, token?: string) => Promise<void>;
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
    // Define all handlers so we can clean them up
    const onConnected = () => {
      setIsConnected(true);
      connectionStore.setStatus('connected');
      connectionStore.resetReconnectAttempts();
    };

    const onDisconnected = () => {
      setIsConnected(false);
      connectionStore.setStatus('disconnected');
    };

    const onReconnecting = () => {
      connectionStore.setStatus('reconnecting');
      connectionStore.incrementReconnectAttempts();
    };

    const onError = (error: any) => {
      console.error('[WebSocket] Error:', error);
      connectionStore.setError(error.message || 'Connection error');
      uiStore.addNotification({
        type: 'error',
        message: error.message || 'WebSocket connection error',
      });
    };

    const onRoomStateChanged = (data: any) => {
      console.log('[WebSocket] Room state changed:', data);
      
      const maxPlayers = data.game_type === 'ddz' ? 3 : data.game_type === 'mahjong' ? 4 : data.game_type === 'durian' ? 7 : 4;
      
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

      // Show room messages in chat, but filter out game-over messages
      // (game over is already shown via notification + in-game UI)
      if (data.message && !data.message.startsWith('游戏结束')) {
        chatStore.addMessage({
          playerId: 'system',
          nickname: '系统',
          content: data.message,
        });
      }
    };

    const onStateUpdate = (data: any) => {
      console.log('[WebSocket] Game state update:', data);
      gameStore.setGameData(data);
    };

    const onGameStarted = (data: any) => {
      console.log('[WebSocket] Game started:', data);
      roomStore.updateRoomStatus('playing');
      if (data && Object.keys(data).length > 0) {
        gameStore.setGameData(data);
      }
      uiStore.addNotification({
        type: 'success',
        message: '游戏开始！',
      });
    };

    const onGameOver = (data: any) => {
      console.log('[WebSocket] Game over:', data);
      roomStore.updateRoomStatus('waiting');
      
      const winner = data?.winner || '未知';
      uiStore.addNotification({
        type: 'info',
        message: `游戏结束 - ${winner}`,
        duration: 8000,
      });
    };

    const onRoomLeft = () => {
      console.log('[WebSocket] Room left confirmed');
      roomStore.leaveRoom();
      gameStore.resetGame();
      chatStore.clearMessages();
    };

    const onChat = (data: any) => {
      console.log('[WebSocket] Chat message:', data);
      chatStore.addMessage({
        playerId: data.player_id,
        nickname: data.nickname || data.player_id,
        content: data.content,
      });
    };

    // Register all handlers
    wsService.on('connected', onConnected);
    wsService.on('disconnected', onDisconnected);
    wsService.on('reconnecting', onReconnecting);
    wsService.on('error', onError);
    wsService.on('room_state_changed', onRoomStateChanged);
    wsService.on('state_update', onStateUpdate);
    wsService.on('game_started', onGameStarted);
    wsService.on('game_over', onGameOver);
    wsService.on('room_left', onRoomLeft);
    wsService.on('chat', onChat);

    return () => {
      // Cleanup: remove all handlers to prevent duplicates on re-mount
      wsService.off('connected', onConnected);
      wsService.off('disconnected', onDisconnected);
      wsService.off('reconnecting', onReconnecting);
      wsService.off('error', onError);
      wsService.off('room_state_changed', onRoomStateChanged);
      wsService.off('state_update', onStateUpdate);
      wsService.off('game_started', onGameStarted);
      wsService.off('game_over', onGameOver);
      wsService.off('room_left', onRoomLeft);
      wsService.off('chat', onChat);
    };
  }, []);

  const connect = async (playerId: string, nickname: string, token?: string): Promise<void> => {
    try {
      connectionStore.setStatus('connecting');
      connectionStore.setPlayer(playerId, nickname);
      userStore.setUser(playerId, nickname);

      // Clear stale room/game state from any previous session
      roomStore.leaveRoom();
      gameStore.resetGame();
      chatStore.clearMessages();
      
      await wsService.connect(playerId, nickname, token);
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
