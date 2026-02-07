import { create } from 'zustand';
import { VM, VMGroup, VMListRequest, VMMetrics, VMMetricsHistoryResponse } from '../types/api';
import { vmApi } from '../api/vm';

const MOCK_MODE = false;

interface VMState {
  vms: VM[];
  total: number;
  isLoading: boolean;
  error: string | null;
  groups: VMGroup[];
  selectedVM: VM | null;
  queryParams: VMListRequest;
  metrics: VMMetrics[];
  isLoadingMetrics: boolean;
  metricsHistory: VMMetricsHistoryResponse | null;
  
  fetchVMs: (params?: VMListRequest) => Promise<void>;
  fetchVMById: (id: string) => Promise<void>;
  createVM: (data: Partial<VM>) => Promise<void>;
  updateVM: (id: string, data: Partial<VM>) => Promise<void>;
  deleteVM: (id: string) => Promise<void>;
  selectVM: (vm: VM | null) => void;
  fetchGroups: () => Promise<void>;
  createGroup: (data: Partial<VMGroup>) => Promise<void>;
  updateGroup: (id: string, data: Partial<VMGroup>) => Promise<void>;
  deleteGroup: (id: string) => Promise<void>;
  setQueryParams: (params: Partial<VMListRequest>) => void;
  fetchAllMetrics: () => Promise<void>;
  fetchMetricsHistory: (vmId: string, period?: string) => Promise<void>;
}

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

export const useVMStore = create<VMState>()((set, get) => ({
  vms: [],
  total: 0,
  isLoading: false,
  error: null,
  groups: [
    { id: 'g1', name: '生产集群', description: '核心生产虚拟机', type: 'cluster', vmCount: 50, onlineCount: 48, offlineCount: 1, errorCount: 1, isSystem: true, createdAt: '', updatedAt: '' },
    { id: 'g2', name: '测试集群', description: '测试环境虚拟机', type: 'cluster', vmCount: 30, onlineCount: 25, offlineCount: 3, errorCount: 2, isSystem: true, createdAt: '', updatedAt: '' },
    { id: 'g3', name: '开发环境', description: '开发测试机器', type: 'custom', vmCount: 20, onlineCount: 18, offlineCount: 2, errorCount: 0, isSystem: false, createdAt: '', updatedAt: '' },
  ],
  selectedVM: null,
  queryParams: { page: 1, pageSize: 20 },
  metrics: [],
  isLoadingMetrics: false,
  metricsHistory: null,

  fetchVMs: async (params) => {
    set({ isLoading: true, error: null });
    try {
      if (MOCK_MODE) {
        await new Promise(r => setTimeout(r, 500));
        set({
          vms: mockVMs,
          total: mockVMs.length,
          isLoading: false,
        });
      } else {
        const queryParams = { ...get().queryParams, ...params };
        const response = await vmApi.list(queryParams);
        set({ vms: response.list, total: response.pagination.total, queryParams, isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : '获取VM列表失败', isLoading: false });
    }
  },

  fetchVMById: async (id) => {
    set({ isLoading: true, error: null });
    try {
      if (MOCK_MODE) {
        await new Promise(r => setTimeout(r, 300));
        const vm = mockVMs.find(v => v.id === id) || mockVMs[0];
        set({ selectedVM: vm, isLoading: false });
      } else {
        const vm = await vmApi.get(id);
        set({ selectedVM: vm, isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : '获取VM详情失败', isLoading: false });
    }
  },

  createVM: async (data) => {
    set({ isLoading: true, error: null });
    try {
      if (!MOCK_MODE) await vmApi.create(data);
      await get().fetchVMs();
      set({ isLoading: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : '创建VM失败', isLoading: false });
      throw error;
    }
  },

  updateVM: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      if (!MOCK_MODE) await vmApi.update(id, data);
      set({ selectedVM: { ...get().selectedVM!, ...data } as VM, isLoading: false });
      await get().fetchVMs();
    } catch (error) {
      set({ error: error instanceof Error ? error.message : '更新VM失败', isLoading: false });
      throw error;
    }
  },

  deleteVM: async (id) => {
    set({ isLoading: true, error: null });
    try {
      if (!MOCK_MODE) await vmApi.delete(id);
      await get().fetchVMs();
      if (get().selectedVM?.id === id) set({ selectedVM: null });
      set({ isLoading: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : '删除VM失败', isLoading: false });
      throw error;
    }
  },

  selectVM: (vm) => set({ selectedVM: vm }),

  fetchGroups: async () => {
    if (MOCK_MODE) return;
    try {
      const groups = await vmApi.getGroups();
      set({ groups });
    } catch (error) {
      console.error('Failed to fetch groups:', error);
    }
  },

  createGroup: async (data) => {
    if (MOCK_MODE) return;
    try {
      await vmApi.createGroup(data);
      await get().fetchGroups();
    } catch (error) {
      console.error('Failed to create group:', error);
      throw error;
    }
  },

  updateGroup: async (id, data) => {
    if (MOCK_MODE) return;
    try {
      await vmApi.updateGroup(id, data);
      await get().fetchGroups();
    } catch (error) {
      console.error('Failed to update group:', error);
      throw error;
    }
  },

  deleteGroup: async (id) => {
    if (MOCK_MODE) return;
    try {
      await vmApi.deleteGroup(id);
      await get().fetchGroups();
    } catch (error) {
      console.error('Failed to delete group:', error);
      throw error;
    }
  },

  setQueryParams: (params) => set({ queryParams: { ...get().queryParams, ...params } }),

  fetchAllMetrics: async () => {
    set({ isLoadingMetrics: true, error: null });
    try {
      const metrics = await vmApi.getAllMetrics();
      set({ metrics, isLoadingMetrics: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : '获取指标失败', isLoadingMetrics: false });
    }
  },

  fetchMetricsHistory: async (vmId, period = '24h') => {
    set({ isLoadingMetrics: true, error: null });
    try {
      const history = await vmApi.getMetricsHistory(vmId, period);
      set({ metricsHistory: history, isLoadingMetrics: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : '获取历史指标失败', isLoadingMetrics: false });
    }
  },
}));
