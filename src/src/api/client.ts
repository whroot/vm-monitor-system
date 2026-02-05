import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { ApiResponse } from '../types/api';
import { useAuthStore } from '../stores/authStore';

// 创建 axios 实例
const apiClient: AxiosInstance = axios.create({
  baseURL: (import.meta.env as unknown as { VITE_API_BASE_URL?: string }).VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 添加 Token
    const token = localStorage.getItem('accessToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    // 添加语言
    const language = localStorage.getItem('language') || 'zh-CN';
    config.headers['Accept-Language'] = language;
    
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => {
    // 如果响应格式正确，返回 data
    if (response.data && typeof (response.data as ApiResponse<unknown>).code === 'number') {
      const apiResponse = response.data as ApiResponse<unknown>;
      if (apiResponse.code >= 200 && apiResponse.code < 300) {
        return apiResponse.data;
      } else {
        return Promise.reject(new Error(apiResponse.message));
      }
    }
    return response.data;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // 401 未授权 - Token过期
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // 尝试刷新 Token
        const refreshToken = localStorage.getItem('refreshToken');
        if (!refreshToken) {
          throw new Error('No refresh token');
        }

        const response = await axios.post(
          `${apiClient.defaults.baseURL}/auth/refresh`,
          { refreshToken },
          { headers: { 'Content-Type': 'application/json' } }
        );

        const { accessToken, refreshToken: newRefreshToken } = response.data.data as { 
          accessToken: string; 
          refreshToken: string;
        };

        // 更新 Token
        localStorage.setItem('accessToken', accessToken);
        localStorage.setItem('refreshToken', newRefreshToken);

        // 重试原请求
        originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        return apiClient(originalRequest);
      } catch (refreshError) {
        // 刷新失败，登出
        useAuthStore.getState().logout();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // 403 无权限
    if (error.response?.status === 403) {
      console.error('Permission denied');
    }

    // 其他错误
    const message = (error.response?.data as ApiResponse<unknown>)?.message || error.message || 'Unknown error';
    return Promise.reject(new Error(message));
  }
);

export default apiClient;

