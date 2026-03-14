import { useWebSocketContext } from '@/contexts/WebSocketContext';
import type { WSMessage } from '@/types/websocket';

export const useWebSocket = () => {
  const { service, isConnected, connect, disconnect } = useWebSocketContext();

  const send = (message: WSMessage) => {
    service.send(message);
  };

  const createRoom = (gameType: 'ddz' | 'mahjong' | 'durian') => {
    send({
      type: 'room.create',
      data: {
        game_type: gameType,
      },
    });
  };

  const joinRoom = (roomId: string) => {
    send({
      type: 'room.join',
      data: {
        room_id: roomId,
      },
    });
  };

  const leaveRoom = (roomId: string) => {
    send({
      type: 'room.leave',
      room_id: roomId,
      data: {},
    });
  };

  const setReady = (roomId: string, ready: boolean) => {
    send({
      type: 'room.action',
      room_id: roomId,
      data: {
        action: 'ready',
        data: { ready },
      },
    });
  };

  const selectSeat = (roomId: string, seatNumber: number) => {
    send({
      type: 'room.action',
      room_id: roomId,
      data: {
        action: 'select_seat',
        data: { seat_number: seatNumber },
      },
    });
  };

  const sendChat = (roomId: string, content: string) => {
    send({
      type: 'chat',
      room_id: roomId,
      data: {
        content,
      },
    });
  };

  const sendGameAction = (roomId: string, action: string, data?: any) => {
    send({
      type: 'game.action',
      room_id: roomId,
      data: {
        action,
        ...data,
      },
    });
  };

  const addBot = (roomId: string) => {
    send({
      type: 'room.action',
      room_id: roomId,
      data: {
        action: 'add_bot',
      },
    });
  };

  return {
    isConnected,
    connect,
    disconnect,
    send,
    createRoom,
    joinRoom,
    leaveRoom,
    setReady,
    selectSeat,
    sendChat,
    sendGameAction,
    addBot,
  };
};
