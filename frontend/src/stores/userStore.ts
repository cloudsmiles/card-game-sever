import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

interface UserState {
  // State
  playerId: string | null;
  nickname: string | null;

  // Actions
  setUser: (playerId: string, nickname: string) => void;
  clearUser: () => void;
}

export const useUserStore = create<UserState>()(
  devtools(
    persist(
      (set) => ({
        // Initial state
        playerId: null,
        nickname: null,

        // Actions
        setUser: (playerId, nickname) => set({ playerId, nickname }),

        clearUser: () => set({ playerId: null, nickname: null }),
      }),
      {
        name: 'user-storage',
      }
    ),
    { name: 'UserStore' }
  )
);
