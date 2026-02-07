import apiClient from './client';

export interface Permission {
  id: string;
  name: string;
  description?: string;
  resource: string;
  action: string;
  level: string;
  scope: string;
  createdAt: string;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
  parentId?: string;
  level: number;
  path: string;
  isSystem: boolean;
  permissions?: Permission[];
  userCount?: number;
  createdAt: string;
  updatedAt: string;
}

const MOCK_MODE = false;

const mockPermissions: Permission[] = [
  { id: 'vm:read', name: '查看虚拟机', resource: 'vm', action: 'read', level: 'read', scope: 'global', createdAt: '' },
  { id: 'vm:write', name: '管理虚拟机', resource: 'vm', action: 'write', level: 'write', scope: 'global', createdAt: '' },
  { id: 'alert:read', name: '查看告警', resource: 'alert', action: 'read', level: 'read', scope: 'global', createdAt: '' },
  { id: 'alert:write', name: '管理告警规则', resource: 'alert', action: 'write', level: 'write', scope: 'global', createdAt: '' },
];

const mockRoles: Role[] = [
  { id: 'r1', name: '系统管理员', description: '拥有所有权限', level: 1, path: '/admin', isSystem: true, userCount: 5, createdAt: '', updatedAt: '' },
  { id: 'r2', name: '运维人员', description: '可以管理VM和告警', level: 5, path: '/ops', isSystem: true, userCount: 10, createdAt: '', updatedAt: '' },
  { id: 'r3', name: '只读用户', description: '只能查看信息', level: 10, path: '/viewer', isSystem: true, userCount: 20, createdAt: '', updatedAt: '' },
];

export const permissionApi = {
  // Permissions
  listPermissions: async (): Promise<Permission[]> => {
    if (MOCK_MODE) return mockPermissions;
    const response = await apiClient.get('/permissions') as Permission[];
    return response;
  },

  getPermission: async (id: string): Promise<Permission> => {
    if (MOCK_MODE) return mockPermissions.find(p => p.id === id) || mockPermissions[0];
    const response = await apiClient.get(`/permissions/${id}`) as Permission;
    return response;
  },

  createPermission: async (data: Omit<Permission, 'id' | 'createdAt'>): Promise<Permission> => {
    if (MOCK_MODE) return { ...data, id: `perm_${Date.now()}`, createdAt: new Date().toISOString() };
    const response = await apiClient.post('/permissions', data) as Permission;
    return response;
  },

  deletePermission: async (id: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.delete(`/permissions/${id}`);
  },

  listRolePermissions: async (roleId: string): Promise<Permission[]> => {
    if (MOCK_MODE) return mockPermissions;
    const response = await apiClient.get(`/permissions/role/${roleId}`) as Permission[];
    return response;
  },

  assignPermissionsToRole: async (roleId: string, permissionIds: string[]): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.post(`/permissions/role/${roleId}`, { permissionIds });
  },

  // Roles
  listRoles: async (): Promise<Role[]> => {
    if (MOCK_MODE) return mockRoles;
    const response = await apiClient.get('/roles') as Role[];
    return response;
  },

  getRole: async (id: string): Promise<Role> => {
    if (MOCK_MODE) return mockRoles.find(r => r.id === id) || mockRoles[0];
    const response = await apiClient.get(`/roles/${id}`) as Role;
    return response;
  },

  createRole: async (data: Partial<Role>): Promise<Role> => {
    if (MOCK_MODE) return { id: `r_${Date.now()}`, name: data.name || '新角色', level: 10, path: '/user', isSystem: false, createdAt: '', updatedAt: '', ...data };
    const apiResponse = await apiClient.post('/roles', data) as { roleId: string };
    return { id: apiResponse.roleId, name: data.name || '', level: data.level || 10, path: data.path || '', isSystem: false, createdAt: '', updatedAt: '', ...data };
  },

  updateRole: async (id: string, data: Partial<Role>): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.put(`/roles/${id}`, data);
  },

  deleteRole: async (id: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.delete(`/roles/${id}`);
  },

  getRoleUsers: async (roleId: string): Promise<unknown[]> => {
    if (MOCK_MODE) return [];
    const response = await apiClient.get(`/roles/${roleId}/users`) as unknown[];
    return response;
  },

  assignUserRoles: async (userId: string, roleIds: string[]): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.post('/roles/users', { userId, roleIds });
  },

  getUserRoles: async (userId: string): Promise<Role[]> => {
    if (MOCK_MODE) return mockRoles;
    const response = await apiClient.get(`/users/${userId}/roles`) as Role[];
    return response;
  },

  getUserPermissions: async (userId: string): Promise<Permission[]> => {
    if (MOCK_MODE) return mockPermissions;
    const response = await apiClient.get(`/users/${userId}/permissions`) as Permission[];
    return response;
  },
};
