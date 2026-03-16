// 用户信息
export interface UserInfo {
  id: number;
  nickname: string;
  avatar_url?: string;
  login_type: 'guest' | 'wechat_h5' | 'wechat_open';
  created_at: string;
}

// 登录响应
export interface LoginResponse {
  token: string;
  user: UserInfo;
}

// API 响应包装
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}
