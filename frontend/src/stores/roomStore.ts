import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { SeatPlayerInfo, GameType, RoomStatus } from '@/types/room';

interface RoomState {
  // State
  currentRoomId: string | null;
  gameType: GameType | null;
  roomStatus: RoomStatus;
  players: SeatPlayerInfo[];
  maxPlayers: number;

  // Actions
  setRoom: (roomId: string, gameType: GameType, maxPlayers: number) => void;
  updateRoomStatus: (status: RoomStatus) => void;
  updatePlayers: (players: SeatPlayerInfo[]) => void;
  addPlayer: (player: SeatPlayerInfo) => void;
  removePlayer: (playerId: string) => void;
  updatePlayerReady: (playerId: string, ready: boolean) => void;
  updatePlayerOffline: (playerId: string, offline: boolean) => void;
  leaveRoom: () => void;
}

export const useRoomStore = create<RoomState>()(
  devtools(
    (set) => ({
      // Initial state
      currentRoomId: null,
      gameType: null,
      roomStatus: 'waiting',
      players: [],
      maxPlayers: 0,

      // Actions
      setRoom: (roomId, gameType, maxPlayers) =>
        set({ currentRoomId: roomId, gameType, maxPlayers }),

      updateRoomStatus: (status) => set({ roomStatus: status }),

      updatePlayers: (players) => set({ players }),

      addPlayer: (player) =>
        set((state) => ({
          players: [...state.players, player],
        })),

      removePlayer: (playerId) =>
        set((state) => ({
          players: state.players.filter((p) => p.player_id !== playerId),
        })),

      updatePlayerReady: (playerId, ready) =>
        set((state) => ({
          players: state.players.map((p) =>
            p.player_id === playerId ? { ...p, ready } : p
          ),
        })),

      updatePlayerOffline: (playerId, offline) =>
        set((state) => ({
          players: state.players.map((p) =>
            p.player_id === playerId ? { ...p, offline } : p
          ),
        })),

      leaveRoom: () =>
        set({
          currentRoomId: null,
          gameType: null,
          roomStatus: 'waiting',
          players: [],
          maxPlayers: 0,
        }),
    }),
    { name: 'RoomStore' }
  )
);
