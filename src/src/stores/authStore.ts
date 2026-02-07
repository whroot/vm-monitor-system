import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User, LoginRequest } from '../types/api';
import { authApi, mockAuth } from '../api/auth';

const MOCK_MODE = false;

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  permissions: string[];
  isLoading: boolean;
  error: string | null;

  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  fetchUser: () => Promise<void>;
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>;
  updateProfile: (data: { name?: string; email?: string; phone?: string; department?: string }) => Promise<void>;
  hasPermission: (permission: string) => boolean;
  setLanguage: (language: string) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      permissions: [],
      isLoading: false,
      error: null,

      login: async (credentials) => {
        set({ isLoading: true, error: null });
        try {
          const response = MOCK_MODE ? await mockAuth.login(credentials) : await authApi.login(credentials);
          
          localStorage.setItem('accessToken', response.accessToken);
          localStorage.setItem('refreshToken', response.refreshToken);
          localStorage.setItem('language', response.user.preferences?.language || 'zh-CN');
          
          set({
            user: response.user,
            isAuthenticated: true,
            permissions: response.permissions,
            isLoading: false,
          });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : '登录失败',
            isLoading: false,
          });
          throw error;
        }
      },

      logout: async () => {
        try {
          if (!MOCK_MODE) await authApi.logout();
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          
          set({
            user: null,
            isAuthenticated: false,
            permissions: [],
            error: null,
          });
        }
      },

      fetchUser: async () => {
        try {
          if (MOCK_MODE) {
            const token = localStorage.getItem('accessToken');
            if (token?.startsWith('mock-')) {
              set({
                user: {
                  id: 'usr_001',
                  username: 'admin',
                  email: 'admin@company.com',
                  name: '系统管理员',
                  department: 'IT部',
                  roles: [{ id: 'role_admin', name: '系统管理员', description: '拥有所有权限', level: 1, path: '/admin', isSystem: true, createdAt: '', updatedAt: '' }],
                  status: 'active' as const,
                  mustChangePassword: false,
                  mfaEnabled: false,
                  preferences: { language: 'zh-CN', theme: 'dark', timezone: 'Asia/Shanghai', dateFormat: 'YYYY-MM-DD' },
                  createdAt: '2026-02-03T10:30:00Z',
                  updatedAt: '2026-02-03T10:30:00Z',
                },
                permissions: ['vm:read', 'vm:write', 'alert:read', 'alert:write', 'user:read', 'user:write', '*'],
                isAuthenticated: true,
              });
              return;
            }
          }
          const response = MOCK_MODE ? { user: null, permissions: [] } : await authApi.getMe();
          const backendUser = response.user;
          const defaultRoles = [{ id: 'role_user', name: '普通用户', description: '普通用户角色', level: 10, path: '/user', isSystem: true, createdAt: '', updatedAt: '' }];
          set({
            user: {
              id: backendUser?.id || 'unknown',
              username: backendUser?.username || 'unknown',
              email: backendUser?.email || '',
              name: backendUser?.name || 'Unknown',
              phone: backendUser?.phone || '',
              department: backendUser?.department || '',
              avatar: backendUser?.avatar || '',
              roles: backendUser?.roles?.length ? backendUser.roles : defaultRoles,
              status: backendUser?.status || 'active',
              mustChangePassword: false,
              mfaEnabled: false,
              preferences: backendUser?.preferences || { language: 'zh-CN', theme: 'dark', timezone: 'Asia/Shanghai', dateFormat: 'YYYY-MM-DD' },
              createdAt: backendUser?.createdAt || new Date().toISOString(),
              updatedAt: backendUser?.updatedAt || new Date().toISOString(),
            },
            permissions: response.permissions,
            isAuthenticated: true,
          });
        } catch (error) {
          get().logout();
        }
      },

      changePassword: async (oldPassword, newPassword) => {
        set({ isLoading: true, error: null });
        try {
          await authApi.changePassword({
            oldPassword,
            newPassword,
            confirmPassword: newPassword,
          });
          set({ isLoading: false });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : '密码修改失败',
            isLoading: false,
          });
          throw error;
        }
      },

      updateProfile: async (data) => {
        set({ isLoading: true, error: null });
        try {
          const updatedUser = await authApi.updateProfile(data);
          set((state) => ({
            user: state.user ? { ...state.user, ...(updatedUser as Partial<User>) } : state.user,
            isLoading: false,
          }));
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : '资料更新失败',
            isLoading: false,
          });
          throw error;
        }
      },

      hasPermission: (permission) => {
        const { permissions } = get();
        return permissions.includes(permission) || permissions.includes('*');
      },

      setLanguage: (language) => {
        const { user } = get();
        if (user) {
          user.preferences.language = language;
          set({ user: { ...user } });
        }
        localStorage.setItem('language', language);
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        permissions: state.permissions,
      }),
    }
  )
);
