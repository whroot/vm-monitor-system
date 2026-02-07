import apiClient from './client';
import { VM, VMListRequest, VMListResponse, VMGroup, VMMetrics, VMMetricsHistoryResponse } from '../types/api';

const MOCK_MODE = false;

const mockVMs: VM[] = Array.from({ length: 20 }, (_, i) => ({
  id: `vm_${i + 1}`,
  vmwareId: `vmware-${1000 + i}`,
  name: `prod-app-${String(i + 1).padStart(2, '0')}`,
  ip: `192.168.1.${100 + i}`,
  os: (i % 2 === 0 ? 'Linux' : 'Windows') as 'Linux' | 'Windows',
  osVersion: i % 2 === 0 ? 'Ubuntu 22.04 LTS' : 'Windows Server 2022',
  cpuCores: 4 + (i % 4),
  memoryGB: 8 + (i % 8) * 4,
  diskGB: 100 + i * 50,
  networkAdapters: 1 + (i % 3),
  powerState: (i % 5 === 0 ? 'poweredOff' : 'poweredOn') as 'poweredOn' | 'poweredOff' | 'suspended',
  hostId: `host-${i % 3 + 1}`,
  hostName: `esxi-${i % 3 + 1}.company.com`,
  datacenterId: `dc-${i % 2 + 1}`,
  datacenterName: i % 2 === 0 ? '数据中心-北京' : '数据中心-上海',
  clusterId: `cluster-${i % 2 + 1}`,
  clusterName: i % 2 === 0 ? '生产集群' : '测试集群',
  status: (i % 10 === 0 ? 'error' : i % 5 === 0 ? 'offline' : 'online') as 'online' | 'offline' | 'error' | 'unknown',
  lastSeen: new Date(Date.now() - i * 60000).toISOString(),
  vmwareToolsStatus: 'running' as const,
  tags: i % 3 === 0 ? ['核心', '数据库'] : i % 2 === 0 ? ['Web服务器'] : ['应用服务器'],
  description: `生产环境虚拟机 #${i + 1}`,
  createdAt: '2025-01-15T10:30:00Z',
  updatedAt: new Date().toISOString(),
}));

export const vmApi = {
  list: async (params: VMListRequest): Promise<VMListResponse> => {
    if (MOCK_MODE) {
      await new Promise(r => setTimeout(r, 300));
      return { list: mockVMs, pagination: { page: params.page || 1, pageSize: params.pageSize || 20, total: mockVMs.length, totalPages: Math.ceil(mockVMs.length / (params.pageSize || 20)) }, summary: { total: mockVMs.length, online: mockVMs.filter(v => v.status === 'online').length, offline: mockVMs.filter(v => v.status === 'offline').length, error: mockVMs.filter(v => v.status === 'error').length } };
    }
    const response = await apiClient.get('/vms', { params }) as { vms: VM[]; total: number; page: number; pageSize: number };
    const runningCount = response.vms.filter((v: VM) => v.status === 'online').length;
    const stoppedCount = response.vms.filter((v: VM) => v.status === 'offline').length;
    const warningCount = response.vms.filter((v: VM) => v.status === 'error').length;
    return {
      list: response.vms,
      pagination: {
        page: response.page,
        pageSize: response.pageSize,
        total: response.total,
        totalPages: Math.ceil(response.total / response.pageSize),
      },
      summary: {
        total: response.total,
        online: runningCount,
        offline: stoppedCount,
        error: warningCount,
      },
    };
  },

  get: async (id: string): Promise<VM> => {
    if (MOCK_MODE) {
      await new Promise(r => setTimeout(r, 200));
      return mockVMs.find(v => v.id === id) || mockVMs[0];
    }
    const response = await apiClient.get(`/vms/${id}`) as VM;
    return response;
  },

  create: async (data: Partial<VM>): Promise<VM> => {
    if (MOCK_MODE) return { ...mockVMs[0], ...data, id: `vm_${Date.now()}`, name: data.name || 'new-vm' } as VM;
    const response = await apiClient.post('/vms', data) as VM;
    return response;
  },

  update: async (id: string, data: Partial<VM>): Promise<VM> => {
    if (MOCK_MODE) return { ...mockVMs[0], ...data, id } as VM;
    const response = await apiClient.put(`/vms/${id}`, data) as VM;
    return response;
  },

  delete: async (id: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.delete(`/vms/${id}`);
  },

  sync: async (data: { type: 'full' | 'incremental'; datacenterId?: string; clusterId?: string; hostId?: string }): Promise<{ syncId: string; status: string }> => {
    if (MOCK_MODE) return { syncId: `sync_${Date.now()}`, status: 'running' };
    return apiClient.post('/vms/sync', data) as unknown as Promise<{ syncId: string; status: string }>;
  },

  getStatistics: async (): Promise<unknown> => {
    if (MOCK_MODE) return { total: 150, online: 140, offline: 5, error: 3, cpuUsage: 45.2, memoryUsage: 62.1 };
    return apiClient.get('/vms/stats') as unknown as Promise<unknown>;
  },

  batch: async (data: { action: 'start' | 'stop' | 'restart' | 'delete'; vmIds: string[]; force?: boolean }): Promise<{ taskId: string; status: string }> => {
    if (MOCK_MODE) return { taskId: `task_${Date.now()}`, status: 'completed' };
    return apiClient.post('/vms/batch', data) as unknown as Promise<{ taskId: string; status: string }>;
  },

  getGroups: async (): Promise<VMGroup[]> => {
    if (MOCK_MODE) return [
      { id: 'g1', name: '生产集群', description: '核心生产虚拟机', type: 'cluster', vmCount: 50, onlineCount: 48, offlineCount: 1, errorCount: 1, isSystem: true, createdAt: '', updatedAt: '' },
      { id: 'g2', name: '测试集群', description: '测试环境虚拟机', type: 'cluster', vmCount: 30, onlineCount: 25, offlineCount: 3, errorCount: 2, isSystem: true, createdAt: '', updatedAt: '' },
      { id: 'g3', name: '开发环境', description: '开发测试机器', type: 'custom', vmCount: 20, onlineCount: 18, offlineCount: 2, errorCount: 0, isSystem: false, createdAt: '', updatedAt: '' },
    ];
    return apiClient.get('/vms/groups') as unknown as Promise<VMGroup[]>;
  },

  createGroup: async (data: Partial<VMGroup>): Promise<VMGroup> => {
    if (MOCK_MODE) return { id: `g_${Date.now()}`, name: data.name || '新分组', vmCount: 0, onlineCount: 0, offlineCount: 0, errorCount: 0, isSystem: false, createdAt: '', updatedAt: '', ...data } as VMGroup;
    return apiClient.post('/vms/groups', data) as unknown as Promise<VMGroup>;
  },

  updateGroup: async (id: string, data: Partial<VMGroup>): Promise<VMGroup> => {
    if (MOCK_MODE) return { id, name: data.name || '分组', vmCount: 0, onlineCount: 0, offlineCount: 0, errorCount: 0, isSystem: false, createdAt: '', updatedAt: '', ...data } as VMGroup;
    return apiClient.put(`/vms/groups/${id}`, data) as unknown as Promise<VMGroup>;
  },

  deleteGroup: async (id: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.delete(`/vms/groups/${id}`);
  },

  getAllMetrics: async (): Promise<VMMetrics[]> => {
    if (MOCK_MODE) {
      return mockVMs.map((vm) => ({
        vmId: vm.id,
        vmName: vm.name,
        cpuUsage: Math.round(Math.random() * 100 * 10) / 10,
        memoryUsage: Math.round(Math.random() * 100 * 10) / 10,
        diskUsage: Math.round(Math.random() * 100 * 10) / 10,
        diskReadMbps: Math.round(Math.random() * 200 * 10) / 10,
        diskWriteMbps: Math.round(Math.random() * 150 * 10) / 10,
        networkInMbps: Math.round(Math.random() * 100 * 10) / 10,
        networkOutMbps: Math.round(Math.random() * 100 * 10) / 10,
        temperature: Math.round(35 + Math.random() * 30),
        updatedAt: new Date().toISOString(),
      }));
    }
    const response = await apiClient.get('/vms/metrics/all') as VMMetrics[];
    return response;
  },

  getMetricsHistory: async (vmId: string, period: string = '24h', startTime?: string, endTime?: string): Promise<VMMetricsHistoryResponse> => {
    if (MOCK_MODE) {
      const vm = mockVMs.find(v => v.id === vmId) || mockVMs[0];
      let hours = period === '1h' ? 1 : period === '6h' ? 6 : period === '24h' ? 24 : period === '7d' ? 168 : 720;
      const metrics = Array.from({ length: hours }, (_, i) => ({
        timestamp: new Date(Date.now() - (hours - i - 1) * 3600000).toISOString(),
        cpuUsage: Math.round(Math.random() * 100 * 10) / 10,
        memoryUsage: Math.round(Math.random() * 100 * 10) / 10,
        diskUsage: Math.round(Math.random() * 100 * 10) / 10,
        diskReadMbps: Math.round(Math.random() * 200 * 10) / 10,
        diskWriteMbps: Math.round(Math.random() * 150 * 10) / 10,
        networkInMbps: Math.round(Math.random() * 100 * 10) / 10,
        networkOutMbps: Math.round(Math.random() * 100 * 10) / 10,
        temperature: Math.round(35 + Math.random() * 30),
      }));
      return { vmId, vmName: vm.name, period, metrics };
    }
    const params: any = { period };
    if (startTime) params.startTime = startTime;
    if (endTime) params.endTime = endTime;
    const response = await apiClient.get(`/vms/${vmId}/metrics/history`, { params }) as any;
    return {
      vmId,
      vmName: response.vmName || '',
      period: response.period || period,
      metrics: response.metrics || [],
    };
  },
};

export const mockVM = { mockVMs };
