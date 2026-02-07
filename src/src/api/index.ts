export * from './auth';
export * from './vm';
export * from './realtime';
export * from './alert';
export * from './dashboard';

// 重新导出所有API
export { default as apiClient } from './client';
export { authApi } from './auth';
export { vmApi } from './vm';
export { realtimeApi, historyApi } from './realtime';
export { alertApi } from './alert';
export { dashboardApi } from './dashboard';