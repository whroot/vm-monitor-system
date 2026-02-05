import apiClient from './client';
import { SystemOverview, RealtimeMetrics, ExportTask } from '../types/api';

const MOCK_MODE = false;

const mockOverview: SystemOverview = {
  timestamp: new Date().toISOString(),
  status: 'healthy',
  healthScore: { value: 92, level: 'excellent', trend: 'stable' },
  vmMonitoring: { totalVMs: 150, onlineVMs: 140, offlineVMs: 5, errorVMs: 3, collectionRate: 99.5, avgCollectionTime: 150 },
  alerts: { critical: 1, high: 3, medium: 8, low: 15 },
};

export const realtimeApi = {
  getVMMetrics: async (vmId: string): Promise<RealtimeMetrics> => {
    if (MOCK_MODE) {
      await new Promise(r => setTimeout(r, 300));
      return {
        vmId,
        timestamp: new Date().toISOString(),
        dataSources: { vsphere: true, guestOS: true },
        cpu: { usageMHz: 2400, usagePercent: 42.5, ready: 2.5, wait: 5.2, load1min: 1.2, load5min: 0.9, load15min: 0.7 },
        memory: { usagePercent: 68, usedMB: 11264, freeMB: 5248, totalMB: 16384, buffersMB: 512, cachedMB: 2048 },
        disk: { usagePercent: 45, readLatency: 8, writeLatency: 12, readIOPS: 150, writeIOPS: 80 },
        network: { inBps: 52428800, outBps: 104857600, inBytes: 52428800, outBytes: 104857600 },
      };
    }
    return apiClient.get(`/realtime/vms/${vmId}`) as unknown as Promise<RealtimeMetrics>;
  },

  batchGetMetrics: async (vmIds: string[], metrics?: string[]): Promise<RealtimeMetrics[]> => {
    if (MOCK_MODE) return Promise.all(vmIds.map(id => realtimeApi.getVMMetrics(id)));
    return apiClient.post('/realtime/vms/batch', { vmIds, metrics }) as unknown as Promise<RealtimeMetrics[]>;
  },

  getGroupMetrics: async (groupId: string): Promise<unknown> => {
    if (MOCK_MODE) return { cpuUsage: 45, memoryUsage: 62, diskUsage: 38 };
    return apiClient.get(`/realtime/groups/${groupId}`) as unknown as Promise<unknown>;
  },

  getOverview: async (): Promise<SystemOverview> => {
    if (MOCK_MODE) {
      await new Promise(r => setTimeout(r, 500));
      return mockOverview;
    }
    return apiClient.get('/realtime/overview') as unknown as Promise<SystemOverview>;
  },
};

export const historyApi = {
  query: async (params: { vmIds: string[]; startTime: string; endTime: string; metrics?: string[]; aggregation?: string; page?: number; pageSize?: number; }): Promise<unknown> => {
    if (MOCK_MODE) return { list: [], pagination: { page: 1, pageSize: 100, total: 0, totalPages: 0 } };
    return apiClient.post('/history/query', params) as unknown as Promise<unknown>;
  },

  aggregate: async (params: { vmIds?: string[]; startTime: string; endTime: string; metrics: string[]; dimensions: string[]; groupBy?: string; }): Promise<unknown> => {
    if (MOCK_MODE) return { avgCpu: 45.2, avgMemory: 62.1 };
    return apiClient.post('/history/aggregate', params) as unknown as Promise<unknown>;
  },

  getTrends: async (params: { vmIds?: string[]; startTime: string; endTime: string; metrics: string[]; forecast?: { enabled: boolean; horizon: number; method: string }; }): Promise<unknown> => {
    if (MOCK_MODE) return { trend: 'up', forecast: [] };
    return apiClient.post('/history/trends', params) as unknown as Promise<unknown>;
  },

  export: async (params: { vmIds: string[]; startTime: string; endTime: string; format: 'csv' | 'excel' | 'json'; aggregation: string; }): Promise<{ id: string; status: string }> => {
    if (MOCK_MODE) return { id: 'export-001', status: 'completed' };
    return apiClient.post('/history/export', params) as unknown as Promise<{ id: string; status: string }>;
  },

  getExportTask: async (taskId: string): Promise<ExportTask> => {
    if (MOCK_MODE) {
      return { id: taskId, status: 'completed', format: 'csv', filename: 'export.csv', result: { fileUrl: '/downloads/export.csv', fileSize: 1024, recordCount: 100, expiresAt: new Date(Date.now() + 86400000).toISOString() }, createdAt: new Date().toISOString() };
    }
    return apiClient.get(`/history/export/${taskId}`) as unknown as Promise<ExportTask>;
  },
};