import apiClient from './client';
import { SystemOverview, RealtimeMetrics, ExportTask } from '../types/api';

export const realtimeApi = {
  // 获取VM当前指标
  getVMMetrics: async (vmId: string): Promise<RealtimeMetrics> => {
    return apiClient.get(`/realtime/vms/${vmId}`) as unknown as Promise<RealtimeMetrics>;
  },

  // 批量获取VM指标
  batchGetMetrics: async (vmIds: string[], metrics?: string[]): Promise<RealtimeMetrics[]> => {
    return apiClient.post('/realtime/vms/batch', { vmIds, metrics }) as unknown as Promise<RealtimeMetrics[]>;
  },

  // 获取分组聚合指标
  getGroupMetrics: async (groupId: string): Promise<unknown> => {
    return apiClient.get(`/realtime/groups/${groupId}`) as unknown as Promise<unknown>;
  },

  // 获取全局概览
  getOverview: async (): Promise<SystemOverview> => {
    return apiClient.get('/realtime/overview') as unknown as Promise<SystemOverview>;
  },
};

export const historyApi = {
  // 查询历史数据
  query: async (params: {
    vmIds: string[];
    startTime: string;
    endTime: string;
    metrics?: string[];
    aggregation?: string;
    page?: number;
    pageSize?: number;
  }): Promise<unknown> => {
    return apiClient.post('/history/query', params) as unknown as Promise<unknown>;
  },

  // 聚合统计
  aggregate: async (params: {
    vmIds?: string[];
    startTime: string;
    endTime: string;
    metrics: string[];
    dimensions: string[];
    groupBy?: string;
  }): Promise<unknown> => {
    return apiClient.post('/history/aggregate', params) as unknown as Promise<unknown>;
  },

  // 趋势分析
  getTrends: async (params: {
    vmIds?: string[];
    startTime: string;
    endTime: string;
    metrics: string[];
    forecast?: { enabled: boolean; horizon: number; method: string };
  }): Promise<unknown> => {
    return apiClient.post('/history/trends', params) as unknown as Promise<unknown>;
  },

  // 导出数据
  export: async (params: {
    vmIds: string[];
    startTime: string;
    endTime: string;
    format: 'csv' | 'excel' | 'json';
    aggregation: string;
  }): Promise<{ id: string; status: string }> => {
    return apiClient.post('/history/export', params) as unknown as Promise<{ id: string; status: string }>;
  },

  // 获取导出任务
  getExportTask: async (taskId: string): Promise<ExportTask> => {
    return apiClient.get(`/history/export/${taskId}`) as unknown as Promise<ExportTask>;
  },
};