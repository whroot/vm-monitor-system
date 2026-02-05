import apiClient from './client';
import {
  LoginRequest,
  LoginResponse,
  User,
  ChangePasswordRequest
} from '../types/api';

export const authApi = {
  // 登录
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    return apiClient.post('/auth/login', data) as unknown as Promise<LoginResponse>;
  },

  // 登出
  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout');
  },

  // 刷新Token
  refreshToken: async (refreshToken: string): Promise<{ accessToken: string; refreshToken: string; expiresIn: number }> => {
    return apiClient.post('/auth/refresh', { refreshToken }) as unknown as Promise<{ accessToken: string; refreshToken: string; expiresIn: number }>;
  },

  // 获取当前用户信息
  getMe: async (): Promise<{ user: User; permissions: string[] }> => {
    return apiClient.get('/auth/me') as unknown as Promise<{ user: User; permissions: string[] }>;
  },

  // 修改密码
  changePassword: async (data: ChangePasswordRequest): Promise<{ passwordChangedAt: string }> => {
    return apiClient.put('/auth/password', data) as unknown as Promise<{ passwordChangedAt: string }>;
  },

  // 检查权限
  checkPermission: async (permission: string, resource?: string): Promise<{ allowed: boolean }> => {
    return apiClient.get('/auth/check', {
      params: { permission, resource }
    }) as unknown as Promise<{ allowed: boolean }>;
  },
};
