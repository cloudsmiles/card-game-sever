import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { GameLog } from '@/types/common';

interface GameState {
  // State
  gameData: any | null;
  currentTurn: string | null;
  phase: string | null;
  logs: GameLog[];

  // Actions
  setGameData: (data: any) => void;
  updateGameState: (updates: Partial<Omit<GameState, 'setGameData' | 'updateGameState' | 'addLog' | 'clearLogs' | 'resetGame'>>) => void;
  addLog: (message: string, playerId?: string) => void;
  clearLogs: () => void;
  resetGame: () => void;
}

export const useGameStore = create<GameState>()(
  devtools(
    (set) => ({
      // Initial state
      gameData: null,
      currentTurn: null,
      phase: null,
      logs: [],

      // Actions
      setGameData: (data) => set({ gameData: data }),

      updateGameState: (updates) => set(updates),

      addLog: (message, playerId) =>
        set((state) => ({
          logs: [
            ...state.logs,
            {
              id: `${Date.now()}-${Math.random()}`,
              timestamp: Date.now(),
              message,
              playerId,
            },
          ],
        })),

      clearLogs: () => set({ logs: [] }),

      resetGame: () =>
        set({
          gameData: null,
          currentTurn: null,
          phase: null,
          logs: [],
        }),
    }),
    { name: 'GameStore' }
  )
);
