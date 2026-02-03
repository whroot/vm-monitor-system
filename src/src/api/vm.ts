import apiClient from './client';
import { VM, VMListRequest, VMListResponse, VMGroup } from '@types/api';

export const vmApi = {
  // 获取VM列表
  list: async (params: VMListRequest): Promise<VMListResponse> => {
    const response = await apiClient.get('/vms', { params });
    return response as VMListResponse;
  },

  // 获取VM详情
  get: async (id: string): Promise<VM> => {
    const response = await apiClient.get(`/vms/${id}`);
    return response as VM;
  },

  // 创建VM
  create: async (data: Partial<VM>): Promise<VM> => {
    const response = await apiClient.post('/vms', data);
    return response as VM;
  },

  // 更新VM
  update: async (id: string, data: Partial<VM>): Promise<VM> => {
    const response = await apiClient.put(`/vms/${id}`, data);
    return response as VM;
  },

  // 删除VM
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/vms/${id}`);
  },

  // 同步VMware信息
  sync: async (data: { type: 'full' | 'incremental'; datacenterId?: string; clusterId?: string; hostId?: string }): Promise<{ syncId: string; status: string }> => {
    const response = await apiClient.post('/vms/sync', data);
    return response as { syncId: string; status: string };
  },

  // 获取VM统计
  getStatistics: async (): Promise<unknown> => {
    const response = await apiClient.get('/vms/statistics');
    return response;
  },

  // 批量操作
  batch: async (data: { action: 'start' | 'stop' | 'restart' | 'delete'; vmIds: string[]; force?: boolean }): Promise<{ taskId: string; status: string }> => {
    const response = await apiClient.post('/vms/batch', data);
    return response as { taskId: string; status: string };
  },

  // ========== 分组管理 ==========

  // 获取分组列表
  getGroups: async (): Promise<VMGroup[]> => {
    const response = await apiClient.get('/vms/groups');
    return response as VMGroup[];
  },

  // 创建分组
  createGroup: async (data: Partial<VMGroup>): Promise<VMGroup> => {
    const response = await apiClient.post('/vms/groups', data);
    return response as VMGroup;
  },

  // 更新分组
  updateGroup: async (id: string, data: Partial<VMGroup>): Promise<VMGroup> => {
    const response = await apiClient.put(`/vms/groups/${id}`, data);
    return response as VMGroup;
  },

  // 删除分组
  deleteGroup: async (id: string): Promise<void> => {
    await apiClient.delete(`/vms/groups/${id}`);
  },
};
