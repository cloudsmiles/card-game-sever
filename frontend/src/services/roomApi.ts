import type { GameType, RoomInfo } from '@/types/room';

interface GetRoomsParams {
  game_type?: GameType;
  status?: 'waiting' | 'playing';
  page?: number;
  limit?: number;
}

interface GetRoomsResponse {
  rooms: RoomInfo[];
  total: number;
  page: number;
  limit: number;
}

class RoomApiService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = (import.meta as any).env?.VITE_API_URL || 'http://localhost:8080';
  }

  /**
   * Get list of rooms with optional filters
   */
  async getRooms(params?: GetRoomsParams): Promise<GetRoomsResponse> {
    try {
      const queryParams = new URLSearchParams();
      
      if (params?.game_type) {
        queryParams.append('game_type', params.game_type);
      }
      if (params?.status) {
        queryParams.append('status', params.status);
      }
      if (params?.page !== undefined) {
        queryParams.append('page', params.page.toString());
      }
      if (params?.limit !== undefined) {
        queryParams.append('limit', params.limit.toString());
      }

      const url = `${this.baseUrl}/api/rooms${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
      
      const response = await fetch(url);
      
      if (!response.ok) {
        throw new Error(`Failed to fetch rooms: ${response.statusText}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error('[RoomAPI] Error fetching rooms:', error);
      throw error;
    }
  }

  /**
   * Get room details by ID
   */
  async getRoomById(roomId: string): Promise<RoomInfo> {
    try {
      const url = `${this.baseUrl}/api/rooms/${roomId}`;
      
      const response = await fetch(url);
      
      if (!response.ok) {
        throw new Error(`Failed to fetch room: ${response.statusText}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error('[RoomAPI] Error fetching room:', error);
      throw error;
    }
  }
}

export const roomApi = new RoomApiService();
