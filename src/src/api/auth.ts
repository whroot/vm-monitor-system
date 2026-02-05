import apiClient from './client';
import { LoginRequest, LoginResponse, User, ChangePasswordRequest } from '../types/api';

const MOCK_MODE = true;

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
    return apiClient.post('/auth/login', data) as unknown as Promise<LoginResponse>;
  },

  logout: async (): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.post('/auth/logout');
  },

  refreshToken: async (refreshToken: string): Promise<{ accessToken: string; refreshToken: string; expiresIn: number }> => {
    if (MOCK_MODE) return { accessToken: 'new-token', refreshToken: 'new-refresh', expiresIn: 3600 };
    return apiClient.post('/auth/refresh', { refreshToken }) as unknown as Promise<{ accessToken: string; refreshToken: string; expiresIn: number }>;
  },

  getMe: async (): Promise<{ user: User; permissions: string[] }> => {
    if (MOCK_MODE) return { user: null as unknown as User, permissions: [] };
    return apiClient.get('/auth/me') as unknown as Promise<{ user: User; permissions: string[] }>;
  },

  changePassword: async (data: ChangePasswordRequest): Promise<{ passwordChangedAt: string }> => {
    if (MOCK_MODE) return { passwordChangedAt: new Date().toISOString() };
    return apiClient.put('/auth/password', data) as unknown as Promise<{ passwordChangedAt: string }>;
  },

  checkPermission: async (permission: string, resource?: string): Promise<{ allowed: boolean }> => {
    if (MOCK_MODE) return { allowed: true };
    return apiClient.get('/auth/check', { params: { permission, resource } }) as unknown as Promise<{ allowed: boolean }>;
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
