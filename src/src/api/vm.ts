import apiClient from './client';
import { VM, VMListRequest, VMListResponse, VMGroup } from '../types/api';

export const vmApi = {
  // 获取VM列表
  list: async (params: VMListRequest): Promise<VMListResponse> => {
    return apiClient.get('/vms', { params }) as unknown as Promise<VMListResponse>;
  },

  // 获取VM详情
  get: async (id: string): Promise<VM> => {
    return apiClient.get(`/vms/${id}`) as unknown as Promise<VM>;
  },

  // 创建VM
  create: async (data: Partial<VM>): Promise<VM> => {
    return apiClient.post('/vms', data) as unknown as Promise<VM>;
  },

  // 更新VM
  update: async (id: string, data: Partial<VM>): Promise<VM> => {
    return apiClient.put(`/vms/${id}`, data) as unknown as Promise<VM>;
  },

  // 删除VM
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/vms/${id}`);
  },

  // 同步VMware信息
  sync: async (data: { type: 'full' | 'incremental'; datacenterId?: string; clusterId?: string; hostId?: string }): Promise<{ syncId: string; status: string }> => {
    return apiClient.post('/vms/sync', data) as unknown as Promise<{ syncId: string; status: string }>;
  },

  // 获取VM统计
  getStatistics: async (): Promise<unknown> => {
    return apiClient.get('/vms/statistics') as unknown as Promise<unknown>;
  },

  // 批量操作
  batch: async (data: { action: 'start' | 'stop' | 'restart' | 'delete'; vmIds: string[]; force?: boolean }): Promise<{ taskId: string; status: string }> => {
    return apiClient.post('/vms/batch', data) as unknown as Promise<{ taskId: string; status: string }>;
  },

  // ========== 分组管理 ==========

  // 获取分组列表
  getGroups: async (): Promise<VMGroup[]> => {
    return apiClient.get('/vms/groups') as unknown as Promise<VMGroup[]>;
  },

  // 创建分组
  createGroup: async (data: Partial<VMGroup>): Promise<VMGroup> => {
    return apiClient.post('/vms/groups', data) as unknown as Promise<VMGroup>;
  },

  // 更新分组
  updateGroup: async (id: string, data: Partial<VMGroup>): Promise<VMGroup> => {
    return apiClient.put(`/vms/groups/${id}`, data) as unknown as Promise<VMGroup>;
  },

  // 删除分组
  deleteGroup: async (id: string): Promise<void> => {
    await apiClient.delete(`/vms/groups/${id}`);
  },
};
