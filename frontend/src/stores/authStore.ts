import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import type { UserInfo } from '@/types/auth';

interface AuthState {
  // State
  token: string | null;
  user: UserInfo | null;
  isLoggedIn: boolean;

  // Actions
  setAuth: (token: string, user: UserInfo) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set) => ({
        token: null,
        user: null,
        isLoggedIn: false,

        setAuth: (token, user) => set({ token, user, isLoggedIn: true }),

        clearAuth: () => set({ token: null, user: null, isLoggedIn: false }),
      }),
      {
        name: 'auth-storage',
      }
    ),
    { name: 'AuthStore' }
  )
);
