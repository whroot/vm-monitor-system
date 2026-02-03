import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User, LoginRequest } from '@types/api';
import { authApi } from '@api/auth';

interface AuthState {
  // 状态
  user: User | null;
  isAuthenticated: boolean;
  permissions: string[];
  isLoading: boolean;
  error: string | null;

  // 方法
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  fetchUser: () => Promise<void>;
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>;
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
          const response = await authApi.login(credentials);
          
          // 保存 Token
          localStorage.setItem('accessToken', response.accessToken);
          localStorage.setItem('refreshToken', response.refreshToken);
          localStorage.setItem('language', response.user.preferences.language);
          
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
          await authApi.logout();
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          // 清除本地存储
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
          const response = await authApi.getMe();
          set({
            user: response.user,
            permissions: response.permissions,
            isAuthenticated: true,
          });
        } catch (error) {
          // Token 无效，登出
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
