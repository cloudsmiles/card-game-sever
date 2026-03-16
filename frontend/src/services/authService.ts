import type { LoginResponse, ApiResponse, UserInfo } from '@/types/auth';

const API_BASE = (import.meta as any).env?.VITE_API_URL || 'http://localhost:8080';

class AuthService {
  private token: string | null = null;

  constructor() {
    // 从 localStorage 恢复 token
    this.token = localStorage.getItem('auth_token');
  }

  getToken(): string | null {
    return this.token;
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('auth_token', token);
    } else {
      localStorage.removeItem('auth_token');
    }
  }

  private async request<T>(path: string, options?: RequestInit): Promise<ApiResponse<T>> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    const resp = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers: { ...headers, ...options?.headers },
    });
    const data = await resp.json();
    if (data.code !== 0) {
      throw new Error(data.message || '请求失败');
    }
    return data;
  }

  // 游客登录
  async guestLogin(nickname: string): Promise<LoginResponse> {
    const resp = await this.request<LoginResponse>('/api/auth/guest', {
      method: 'POST',
      body: JSON.stringify({ nickname }),
    });
    this.setToken(resp.data.token);
    return resp.data;
  }

  // 微信登录
  async wechatLogin(code: string, platform: 'h5' | 'open' = 'h5'): Promise<LoginResponse> {
    const resp = await this.request<LoginResponse>('/api/auth/wechat', {
      method: 'POST',
      body: JSON.stringify({ code, platform }),
    });
    this.setToken(resp.data.token);
    return resp.data;
  }

  // 获取当前用户信息
  async getMe(): Promise<UserInfo> {
    const resp = await this.request<UserInfo>('/api/auth/me');
    return resp.data;
  }

  // 登出
  logout() {
    this.setToken(null);
  }

  // 是否已登录
  isLoggedIn(): boolean {
    return !!this.token;
  }
}

export const authService = new AuthService();
