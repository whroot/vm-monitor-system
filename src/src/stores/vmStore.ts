import { create } from 'zustand';
import { VM, VMGroup, VMListRequest, VMListResponse } from '@types/api';
import { vmApi } from '@api/vm';

interface VMState {
  // VM列表
  vms: VM[];
  total: number;
  isLoading: boolean;
  error: string | null;
  
  // 分组
  groups: VMGroup[];
  
  // 选中的VM
  selectedVM: VM | null;
  
  // 查询参数
  queryParams: VMListRequest;
  
  // 方法
  fetchVMs: (params?: VMListRequest) => Promise<void>;
  fetchVMById: (id: string) => Promise<void>;
  createVM: (data: Partial<VM>) => Promise<void>;
  updateVM: (id: string, data: Partial<VM>) => Promise<void>;
  deleteVM: (id: string) => Promise<void>;
  selectVM: (vm: VM | null) => void;
  
  // 分组方法
  fetchGroups: () => Promise<void>;
  createGroup: (data: Partial<VMGroup>) => Promise<void>;
  updateGroup: (id: string, data: Partial<VMGroup>) => Promise<void>;
  deleteGroup: (id: string) => Promise<void>;
  
  // 设置查询参数
  setQueryParams: (params: Partial<VMListRequest>) => void;
}

export const useVMStore = create<VMState>()((set, get) => ({
  vms: [],
  total: 0,
  isLoading: false,
  error: null,
  groups: [],
  selectedVM: null,
  queryParams: {
    page: 1,
    pageSize: 20,
  },

  fetchVMs: async (params) => {
    set({ isLoading: true, error: null });
    try {
      const queryParams = { ...get().queryParams, ...params };
      const response = await vmApi.list(queryParams);
      set({
        vms: response.list,
        total: response.pagination.total,
        queryParams,
        isLoading: false,
      });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取VM列表失败',
        isLoading: false,
      });
    }
  },

  fetchVMById: async (id) => {
    set({ isLoading: true, error: null });
    try {
      const vm = await vmApi.get(id);
      set({
        selectedVM: vm,
        isLoading: false,
      });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取VM详情失败',
        isLoading: false,
      });
    }
  },

  createVM: async (data) => {
    set({ isLoading: true, error: null });
    try {
      await vmApi.create(data);
      // 刷新列表
      await get().fetchVMs();
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '创建VM失败',
        isLoading: false,
      });
      throw error;
    }
  },

  updateVM: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      const updated = await vmApi.update(id, data);
      set({
        selectedVM: updated,
        isLoading: false,
      });
      // 刷新列表
      await get().fetchVMs();
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '更新VM失败',
        isLoading: false,
      });
      throw error;
    }
  },

  deleteVM: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await vmApi.delete(id);
      // 刷新列表
      await get().fetchVMs();
      // 如果删除的是当前选中的VM，清空选中
      if (get().selectedVM?.id === id) {
        set({ selectedVM: null });
      }
      set({ isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '删除VM失败',
        isLoading: false,
      });
      throw error;
    }
  },

  selectVM: (vm) => {
    set({ selectedVM: vm });
  },

  // 分组方法
  fetchGroups: async () => {
    try {
      const groups = await vmApi.getGroups();
      set({ groups });
    } catch (error) {
      console.error('Failed to fetch groups:', error);
    }
  },

  createGroup: async (data) => {
    try {
      await vmApi.createGroup(data);
      await get().fetchGroups();
    } catch (error) {
      console.error('Failed to create group:', error);
      throw error;
    }
  },

  updateGroup: async (id, data) => {
    try {
      await vmApi.updateGroup(id, data);
      await get().fetchGroups();
    } catch (error) {
      console.error('Failed to update group:', error);
      throw error;
    }
  },

  deleteGroup: async (id) => {
    try {
      await vmApi.deleteGroup(id);
      await get().fetchGroups();
    } catch (error) {
      console.error('Failed to delete group:', error);
      throw error;
    }
  },

  setQueryParams: (params) => {
    set({ queryParams: { ...get().queryParams, ...params } });
  },
}));
