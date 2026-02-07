import apiClient from './client';
import { LoginRequest, LoginResponse, User, ChangePasswordRequest } from '../types/api';

const MOCK_MODE = false;

export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    if (MOCK_MODE) {
      await new Promise(r => setTimeout(r, 800));
      if (data.username === 'admin' && data.password === 'password') {
        return {
          user: { id: 'usr_001', username: 'admin', email: 'admin@company.com', name: '系统管理员', department: 'IT部', roles: [{ id: 'role_admin', name: '系统管理员', description: '拥有所有权限', level: 1, path: '/admin', isSystem: true, createdAt: '', updatedAt: '' }], status: 'active' as const, mustChangePassword: false, mfaEnabled: false, preferences: { language: data.language || 'zh-CN', theme: 'dark', timezone: 'Asia/Shanghai', dateFormat: 'YYYY-MM-DD' }, createdAt: '2026-02-03T10:30:00Z', updatedAt: '2026-02-03T10:30:00Z' },
          accessToken: 'mock-access-token-' + Date.now(),
          refreshToken: 'mock-refresh-token-' + Date.now(),
          expiresIn: 3600,
          permissions: ['vm:read', 'vm:write', 'alert:read', 'alert:write', 'user:read', 'user:write', '*'],
        };
      }
      throw new Error('用户名或密码错误');
    }
    const loginData = await apiClient.post('/auth/login', data) as { user: User; accessToken: string; refreshToken: string; expiresIn: number; tokenType: string };
    return {
      user: loginData.user,
      accessToken: loginData.accessToken,
      refreshToken: loginData.refreshToken,
      expiresIn: loginData.expiresIn,
      permissions: ['vm:read', 'vm:write', 'alert:read', 'alert:write', 'user:read', 'user:write'],
    };
  },

  logout: async (): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.post('/auth/logout');
  },

  refreshToken: async (refreshToken: string): Promise<{ accessToken: string; refreshToken: string; expiresIn: number }> => {
    if (MOCK_MODE) return { accessToken: 'new-token', refreshToken: 'new-refresh', expiresIn: 3600 };
    const response = await apiClient.post('/auth/refresh', { refreshToken }) as unknown;
    const data = response as { accessToken: string; refreshToken: string; expiresIn: number };
    return {
      accessToken: data.accessToken,
      refreshToken: data.refreshToken,
      expiresIn: data.expiresIn,
    };
  },

  getMe: async (): Promise<{ user: User; permissions: string[] }> => {
    if (MOCK_MODE) return { user: null as unknown as User, permissions: [] };
    const apiResponse = await apiClient.get('/auth/profile') as User;
    return {
      user: apiResponse,
      permissions: ['vm:read', 'vm:write', 'alert:read', 'alert:write', 'user:read', 'user:write'],
    };
  },

  changePassword: async (data: ChangePasswordRequest): Promise<{ passwordChangedAt: string }> => {
    if (MOCK_MODE) return { passwordChangedAt: new Date().toISOString() };
    await apiClient.put('/auth/password', data);
    return { passwordChangedAt: new Date().toISOString() };
  },

  checkPermission: async (_permission: string, _resource?: string): Promise<{ allowed: boolean }> => {
    if (MOCK_MODE) return { allowed: true };
    return { allowed: true };
  },

  register: async (data: { username: string; email: string; name: string; password: string }): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.post('/auth/register', data);
  },

  listUsers: async (): Promise<any[]> => {
    if (MOCK_MODE) return [];
    const response = await apiClient.get('/users') as any[];
    return response;
  },

  deleteUser: async (userId: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.delete(`/users/${userId}`, { data: {} });
  },

  updateProfile: async (data: { name?: string; email?: string; phone?: string; department?: string }): Promise<Partial<User>> => {
    if (MOCK_MODE) return data;
    const apiResponse = await apiClient.put('/auth/profile', data) as User;
    return apiResponse;
  },

  uploadAvatar: async (formData: FormData): Promise<{ avatarUrl: string }> => {
    if (MOCK_MODE) return { avatarUrl: 'https://via.placeholder.com/150' };
    const response = await apiClient.post('/auth/avatar', formData) as { avatarUrl: string };
    return response;
  },
};

export const mockAuth = {
  login: async (credentials: { username: string; password: string }) => {
    await new Promise(r => setTimeout(r, 800));
    if (credentials.username === 'admin' && credentials.password === 'password') {
      return {
        user: { id: 'usr_001', username: 'admin', email: 'admin@company.com', name: '系统管理员', department: 'IT部', roles: [{ id: 'role_admin', name: '系统管理员', description: '拥有所有权限', level: 1, path: '/admin', isSystem: true, createdAt: '', updatedAt: '' }], status: 'active' as const, mustChangePassword: false, mfaEnabled: false, preferences: { language: 'zh-CN', theme: 'dark', timezone: 'Asia/Shanghai', dateFormat: 'YYYY-MM-DD' }, createdAt: '2026-02-03T10:30:00Z', updatedAt: '2026-02-03T10:30:00Z' },
        accessToken: 'mock-access-token-' + Date.now(),
        refreshToken: 'mock-refresh-token-' + Date.now(),
        expiresIn: 3600,
        permissions: ['vm:read', 'vm:write', 'alert:read', 'alert:write', 'user:read', 'user:write', '*'],
      };
    }
    throw new Error('用户名或密码错误');
  },
  logout: async () => {},
  getMe: async () => ({ user: null, permissions: [] }),
};
