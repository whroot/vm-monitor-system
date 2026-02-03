import apiClient from './client';
import { 
  LoginRequest, 
  LoginResponse, 
  ApiResponse,
  User,
  ChangePasswordRequest
} from '@types/api';

export const authApi = {
  // 登录
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/auth/login', data);
    return response as LoginResponse;
  },

  // 登出
  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout');
  },

  // 刷新Token
  refreshToken: async (refreshToken: string): Promise<{ accessToken: string; refreshToken: string; expiresIn: number }> => {
    const response = await apiClient.post('/auth/refresh', { refreshToken });
    return response as { accessToken: string; refreshToken: string; expiresIn: number };
  },

  // 获取当前用户信息
  getMe: async (): Promise<{ user: User; permissions: string[] }> => {
    const response = await apiClient.get('/auth/me');
    return response as { user: User; permissions: string[] };
  },

  // 修改密码
  changePassword: async (data: ChangePasswordRequest): Promise<{ passwordChangedAt: string }> => {
    const response = await apiClient.put('/auth/password', data);
    return response as { passwordChangedAt: string };
  },

  // 检查权限
  checkPermission: async (permission: string, resource?: string): Promise<{ allowed: boolean }> => {
    const response = await apiClient.get('/auth/check', {
      params: { permission, resource }
    });
    return response as { allowed: boolean };
  },
};
