import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

export interface ChatMessage {
  id: string;
  playerId: string;
  nickname: string;
  content: string;
  timestamp: number;
}

interface ChatState {
  messages: ChatMessage[];
  addMessage: (msg: Omit<ChatMessage, 'id' | 'timestamp'>) => void;
  clearMessages: () => void;
}

export const useChatStore = create<ChatState>()(
  devtools(
    (set) => ({
      messages: [],

      addMessage: (msg) =>
        set((state) => {
          // Deduplicate: skip if same content from same player within 2 seconds
          const now = Date.now();
          const isDuplicate = state.messages.some(
            (m) =>
              m.playerId === msg.playerId &&
              m.content === msg.content &&
              now - m.timestamp < 2000
          );
          if (isDuplicate) return state;

          return {
            messages: [
              ...state.messages,
              {
                ...msg,
                id: `${now}-${Math.random()}`,
                timestamp: now,
              },
            ],
          };
        }),

      clearMessages: () => set({ messages: [] }),
    }),
    { name: 'ChatStore' }
  )
);
