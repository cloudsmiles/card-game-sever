import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { ConnectionStatus } from '@/types/websocket';

interface ConnectionState {
  // State
  status: ConnectionStatus;
  playerId: string | null;
  nickname: string | null;
  error: string | null;
  reconnectAttempts: number;

  // Actions
  setStatus: (status: ConnectionStatus) => void;
  setPlayer: (playerId: string, nickname: string) => void;
  setError: (error: string | null) => void;
  incrementReconnectAttempts: () => void;
  resetReconnectAttempts: () => void;
  reset: () => void;
}

export const useConnectionStore = create<ConnectionState>()(
  devtools(
    (set) => ({
      // Initial state
      status: 'disconnected',
      playerId: null,
      nickname: null,
      error: null,
      reconnectAttempts: 0,

      // Actions
      setStatus: (status) => set({ status }),

      setPlayer: (playerId, nickname) => set({ playerId, nickname }),

      setError: (error) => set({ error }),

      incrementReconnectAttempts: () =>
        set((state) => ({ reconnectAttempts: state.reconnectAttempts + 1 })),

      resetReconnectAttempts: () => set({ reconnectAttempts: 0 }),

      reset: () =>
        set({
          status: 'disconnected',
          playerId: null,
          nickname: null,
          error: null,
          reconnectAttempts: 0,
        }),
    }),
    { name: 'ConnectionStore' }
  )
);
