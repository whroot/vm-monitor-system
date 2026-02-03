import apiClient from './client';
import { SystemOverview, RealtimeMetrics, ExportTask } from '@types/api';

export const realtimeApi = {
  // 获取VM当前指标
  getVMMetrics: async (vmId: string): Promise<RealtimeMetrics> => {
    const response = await apiClient.get(`/realtime/vms/${vmId}`);
    return response as RealtimeMetrics;
  },

  // 批量获取VM指标
  batchGetMetrics: async (vmIds: string[], metrics?: string[]): Promise<RealtimeMetrics[]> => {
    const response = await apiClient.post('/realtime/vms/batch', { vmIds, metrics });
    return (response as { metrics: RealtimeMetrics[] }).metrics;
  },

  // 获取分组聚合指标
  getGroupMetrics: async (groupId: string): Promise<unknown> => {
    const response = await apiClient.get(`/realtime/groups/${groupId}`);
    return response;
  },

  // 获取全局概览
  getOverview: async (): Promise<SystemOverview> => {
    const response = await apiClient.get('/realtime/overview');
    return response as SystemOverview;
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
    const response = await apiClient.post('/history/query', params);
    return response;
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
    const response = await apiClient.post('/history/aggregate', params);
    return response;
  },

  // 趋势分析
  getTrends: async (params: {
    vmIds?: string[];
    startTime: string;
    endTime: string;
    metrics: string[];
    forecast?: { enabled: boolean; horizon: number; method: string };
  }): Promise<unknown> => {
    const response = await apiClient.post('/history/trends', params);
    return response;
  },

  // 导出数据
  export: async (params: {
    vmIds: string[];
    startTime: string;
    endTime: string;
    format: 'csv' | 'excel' | 'json';
    aggregation: string;
  }): Promise<{ id: string; status: string }> => {
    const response = await apiClient.post('/history/export', params);
    return response as { id: string; status: string };
  },

  // 获取导出任务
  getExportTask: async (taskId: string): Promise<ExportTask> => {
    const response = await apiClient.get(`/history/export/${taskId}`);
    return response as ExportTask;
  },
};