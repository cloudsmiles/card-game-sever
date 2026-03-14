import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import type { Notification } from '@/types/common';

interface UIState {
  // State
  theme: 'light' | 'dark';
  notifications: Notification[];
  chatOpen: boolean;
  rulesOpen: boolean;

  // Actions
  toggleTheme: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  addNotification: (notification: Omit<Notification, 'id'>) => void;
  removeNotification: (id: string) => void;
  toggleChat: () => void;
  setChat: (open: boolean) => void;
  toggleRules: () => void;
  setRules: (open: boolean) => void;
}

export const useUIStore = create<UIState>()(
  devtools(
    persist(
      (set) => ({
        // Initial state
        theme: 'light',
        notifications: [],
        chatOpen: true,
        rulesOpen: false,

        // Actions
        toggleTheme: () =>
          set((state) => ({
            theme: state.theme === 'light' ? 'dark' : 'light',
          })),

        setTheme: (theme) => set({ theme }),

        addNotification: (notification) =>
          set((state) => ({
            notifications: [
              ...state.notifications,
              {
                ...notification,
                id: `${Date.now()}-${Math.random()}`,
              },
            ],
          })),

        removeNotification: (id) =>
          set((state) => ({
            notifications: state.notifications.filter((n) => n.id !== id),
          })),

        toggleChat: () => set((state) => ({ chatOpen: !state.chatOpen })),

        setChat: (open) => set({ chatOpen: open }),

        toggleRules: () => set((state) => ({ rulesOpen: !state.rulesOpen })),

        setRules: (open) => set({ rulesOpen: open }),
      }),
      {
        name: 'ui-storage',
        partialize: (state) => ({ theme: state.theme }),
      }
    ),
    { name: 'UIStore' }
  )
);
