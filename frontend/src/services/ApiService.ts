// API服务 - 封装HTTP请求

// 开发环境使用相对路径（Vite代理），生产环境使用环境变量
const getApiBaseUrl = () => {
  if (import.meta.env.DEV) {
    return '/api';
  }
  return (import.meta.env.VITE_API_BASE_URL || '') + '/api';
};

const API_BASE_URL = getApiBaseUrl();

export interface RoomListResponse {
  code: number;
  message: string;
  data: RoomInfo[];
}

// 后端返回的房间信息（简化版）
export interface RoomInfo {
  id: string;
  game_type: string;
  state: string;
  players: number;
  max_players: number;
}

export const apiService = {
  // 获取房间列表
  async getRoomList(): Promise<RoomInfo[]> {
    try {
      const response = await fetch(`${API_BASE_URL}/rooms`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data: RoomListResponse = await response.json();
      if (data.code !== 0) {
        throw new Error(data.message || '获取房间列表失败');
      }
      return data.data;
    } catch (error) {
      console.error('获取房间列表失败:', error);
      throw error;
    }
  }
};