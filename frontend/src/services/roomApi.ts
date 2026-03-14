import type { GameType } from '@/types/room';

// 后端返回的房间信息格式
export interface ApiRoomInfo {
  id: string;
  game_type: string;
  state: string;
  players: number;
  max_players: number;
}

// 后端统一响应格式
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface GetRoomsParams {
  game_type?: GameType;
  status?: 'waiting' | 'playing';
}

class RoomApiService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = (import.meta as any).env?.VITE_API_URL || 'http://localhost:8080/api';
  }

  async getRooms(params?: GetRoomsParams): Promise<ApiRoomInfo[]> {
    try {
      const queryParams = new URLSearchParams();

      if (params?.game_type) {
        queryParams.append('game_type', params.game_type);
      }
      if (params?.status) {
        queryParams.append('status', params.status);
      }

      const qs = queryParams.toString();
      const url = `${this.baseUrl}/rooms${qs ? `?${qs}` : ''}`;

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`Failed to fetch rooms: ${response.statusText}`);
      }

      const result: ApiResponse<ApiRoomInfo[]> = await response.json();

      if (result.code !== 0) {
        throw new Error(result.message || 'Unknown error');
      }

      return result.data || [];
    } catch (error) {
      console.error('[RoomAPI] Error fetching rooms:', error);
      throw error;
    }
  }
}

export const roomApi = new RoomApiService();
