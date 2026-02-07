import { create } from 'zustand';
import { AlertRecord } from '../types/api';
import { alertApi, AlertRule, AlertStats } from '../api/alert';

interface AlertState {
  rules: AlertRule[];
  records: AlertRecord[];
  stats: AlertStats | null;
  isLoading: boolean;
  error: string | null;
  selectedRecord: AlertRecord | null;
  total: number;
  pagination: {
    page: number;
    pageSize: number;
    total: number;
  };

  fetchRules: (params?: { page?: number; pageSize?: number }) => Promise<void>;
  fetchRuleById: (id: string) => Promise<AlertRule | null>;
  createRule: (data: { name: string; scope: string; severity: string }) => Promise<void>;
  updateRule: (id: string, data: Partial<AlertRule>) => Promise<void>;
  deleteRule: (id: string) => Promise<void>;
  fetchRecords: (params?: { page?: number; pageSize?: number }) => Promise<void>;
  fetchStats: () => Promise<void>;
  acknowledgeAlert: (id: string) => Promise<void>;
  resolveAlert: (id: string, resolution: string) => Promise<void>;
  createTestAlert: (data: { vmName: string; severity: string; metric: string; value: number }) => Promise<void>;
  selectRecord: (record: AlertRecord | null) => void;
  setPagination: (params: { page?: number; pageSize?: number }) => void;
}

export const useAlertStore = create<AlertState>()((set, get) => ({
  rules: [],
  records: [],
  stats: null,
  isLoading: false,
  error: null,
  selectedRecord: null,
  total: 0,
  pagination: {
    page: 1,
    pageSize: 10,
    total: 0,
  },

  fetchRules: async (params) => {
    set({ isLoading: true, error: null });
    try {
      const response = await alertApi.listRules(params);
      set({
        rules: response.rules || [],
        total: response.total || 0,
        pagination: {
          page: response.page || 1,
          pageSize: response.pageSize || 10,
          total: response.total || 0,
        },
        isLoading: false,
      });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取告警规则失败',
        isLoading: false,
      });
    }
  },

  fetchRuleById: async (id) => {
    set({ isLoading: true, error: null });
    try {
      const rule = await alertApi.getRule(id);
      set({ isLoading: false });
      return rule;
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取告警规则详情失败',
        isLoading: false,
      });
      return null;
    }
  },

  createRule: async (data) => {
    set({ isLoading: true, error: null });
    try {
      await alertApi.createRule(data);
      await get().fetchRules();
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '创建告警规则失败',
        isLoading: false,
      });
      throw error;
    }
  },

  updateRule: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      await alertApi.updateRule(id, data);
      await get().fetchRules();
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '更新告警规则失败',
        isLoading: false,
      });
      throw error;
    }
  },

  deleteRule: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await alertApi.deleteRule(id);
      await get().fetchRules();
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '删除告警规则失败',
        isLoading: false,
      });
      throw error;
    }
  },

  fetchRecords: async (params) => {
    set({ isLoading: true, error: null });
    try {
      const response = await alertApi.listRecords(params);
      set({
        records: response.records || [],
        pagination: {
          page: response.page || 1,
          pageSize: response.pageSize || 10,
          total: response.total || 0,
        },
        isLoading: false,
      });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取告警记录失败',
        isLoading: false,
      });
    }
  },

  fetchStats: async () => {
    set({ isLoading: true, error: null });
    try {
      const stats = await alertApi.getStats();
      set({ stats, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取告警统计失败',
        isLoading: false,
      });
    }
  },

  acknowledgeAlert: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await alertApi.acknowledge(id);
      await get().fetchRecords();
      await get().fetchStats();
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '确认告警失败',
        isLoading: false,
      });
      throw error;
    }
  },

  resolveAlert: async (id, resolution) => {
    set({ isLoading: true, error: null });
    try {
      await alertApi.resolve(id, resolution);
      await get().fetchRecords();
      await get().fetchStats();
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '解决告警失败',
        isLoading: false,
      });
      throw error;
    }
  },

  createTestAlert: async (data) => {
    set({ isLoading: true, error: null });
    try {
      await alertApi.createTestAlert(data);
      await get().fetchRecords();
      await get().fetchStats();
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '创建测试告警失败',
        isLoading: false,
      });
      throw error;
    }
  },

  selectRecord: (record) => set({ selectedRecord: record }),

  setPagination: (params) =>
    set({
      pagination: { ...get().pagination, ...params },
    }),
}));
