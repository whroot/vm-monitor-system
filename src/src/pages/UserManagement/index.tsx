import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Users, Shield, Plus, Edit, Trash2, Check, X, Key } from 'lucide-react';
import { permissionApi, Role, Permission } from '../../api/permission';
import { authApi } from '../../api/auth';

interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  status: string;
  roles: Role[];
}

const mockUsers: User[] = [];

const UserManagement: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'users' | 'roles'>('users');
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [showRoleModal, setShowRoleModal] = useState(false);
  const [showPermissionModal, setShowPermissionModal] = useState(false);
  const [showAddUserModal, setShowAddUserModal] = useState(false);
  const [showRolePermissionsModal, setShowRolePermissionsModal] = useState(false);
  const [rolePermissions, setRolePermissions] = useState<string[]>([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [usersData, rolesData, permissionsData] = await Promise.all([
        authApi.listUsers(),
        permissionApi.listRoles(),
        permissionApi.listPermissions(),
      ]);
      const mappedUsers: User[] = usersData.map((u: any) => ({
        id: u.id,
        username: u.username,
        email: u.email,
        name: u.name,
        status: u.status,
        roles: u.roles || [],
      }));
      setUsers(mappedUsers);
      setRoles(rolesData);
      setPermissions(permissionsData);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateUser = async (userData: { username: string; email: string; name: string; password: string }) => {
    try {
      await authApi.register(userData);
      await loadData();
      setShowAddUserModal(false);
    } catch (error: any) {
      console.error('创建用户失败:', error);
      alert(error.message || '创建用户失败');
    }
  };

  const handleAssignRoles = async (userId: string, roleIds: string[]) => {
    try {
      await permissionApi.assignUserRoles(userId, roleIds);
      await loadData();
      setShowRoleModal(false);
    } catch (error) {
      console.error('Failed to assign roles:', error);
    }
  };

  const handleCreateRole = async (roleData: Partial<Role>) => {
    try {
      await permissionApi.createRole(roleData);
      await loadData();
      setShowPermissionModal(false);
    } catch (error) {
      console.error('Failed to create role:', error);
    }
  };

  const handleDeleteRole = async (roleId: string) => {
    if (!confirm('确定要删除此角色吗？')) return;
    try {
      await permissionApi.deleteRole(roleId);
      await loadData();
    } catch (error) {
      console.error('Failed to delete role:', error);
    }
  };

  const handleDeleteUser = async (userId: string) => {
    if (!confirm('确定要删除此用户吗？此操作不可恢复。')) return;
    try {
      await authApi.deleteUser(userId);
      await loadData();
    } catch (error: any) {
      console.error('删除用户失败:', error);
      alert(error.message || '删除用户失败');
    }
  };

  const handleManageRolePermissions = async (role: Role) => {
    setSelectedRole(role);
    try {
      const rolePerms = await permissionApi.listRolePermissions(role.id);
      setRolePermissions(rolePerms.map(p => p.id));
      setShowRolePermissionsModal(true);
    } catch (error: any) {
      console.error('获取角色权限失败:', error);
      alert(error.message || '获取权限失败');
    }
  };

  const handleSaveRolePermissions = async () => {
    if (!selectedRole) return;
    try {
      await permissionApi.assignPermissionsToRole(selectedRole.id, rolePermissions);
      await loadData();
      setShowRolePermissionsModal(false);
    } catch (error: any) {
      console.error('保存权限失败:', error);
      alert(error.message || '保存权限失败');
    }
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
          <Shield className="w-6 h-6" />
          权限管理
        </h1>
      </div>

      <div className="flex gap-2 border-b border-gray-700">
        <button
          onClick={() => setActiveTab('users')}
          className={`px-4 py-2 flex items-center gap-2 border-b-2 transition-colors ${
            activeTab === 'users'
              ? 'border-primary text-primary'
              : 'border-transparent text-text-muted hover:text-white'
          }`}
        >
          <Users className="w-4 h-4" />
          用户管理
        </button>
        <button
          onClick={() => setActiveTab('roles')}
          className={`px-4 py-2 flex items-center gap-2 border-b-2 transition-colors ${
            activeTab === 'roles'
              ? 'border-primary text-primary'
              : 'border-transparent text-text-muted hover:text-white'
          }`}
        >
          <Shield className="w-4 h-4" />
          角色管理
        </button>
      </div>

      <div className="grid grid-cols-4 gap-6">
        <div className="card text-center">
          <div className="w-12 h-12 bg-success/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <Users className="w-6 h-6 text-success" />
          </div>
          <div className="text-3xl font-bold text-white">{users.length}</div>
          <div className="text-sm text-text-muted">总用户数</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-info/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <Shield className="w-6 h-6 text-info" />
          </div>
          <div className="text-3xl font-bold text-white">{roles.length}</div>
          <div className="text-sm text-text-muted">角色数</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-warning/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <Key className="w-6 h-6 text-warning" />
          </div>
          <div className="text-3xl font-bold text-white">{permissions.length}</div>
          <div className="text-sm text-text-muted">权限数</div>
        </div>
      </div>

      {activeTab === 'users' ? (
        <div className="card">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-white">用户列表</h3>
            <button
              onClick={() => setShowAddUserModal(true)}
              className="btn-primary flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              新建用户
            </button>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-text-muted text-sm border-b border-gray-700">
                  <th className="pb-3 font-medium">用户名</th>
                  <th className="pb-3 font-medium">邮箱</th>
                  <th className="pb-3 font-medium">姓名</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">角色</th>
                  <th className="pb-3 font-medium text-right">操作</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id} className="border-b border-gray-700/50 hover:bg-gray-700/20">
                    <td className="py-3 text-white font-medium">{user.username}</td>
                    <td className="py-3 text-text-muted">{user.email}</td>
                    <td className="py-3 text-white">{user.name}</td>
                    <td className="py-3">
                      <span className={`px-2 py-1 rounded-full text-xs ${
                        user.status === 'active' 
                          ? 'bg-success/20 text-success' 
                          : 'bg-error/20 text-error'
                      }`}>
                        {user.status === 'active' ? '活跃' : '禁用'}
                      </span>
                    </td>
                    <td className="py-3">
                      <div className="flex gap-1">
                        {user.roles.map((role) => (
                          <span key={role.id} className="px-2 py-0.5 bg-primary/20 text-primary rounded text-xs">
                            {role.name}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={() => {
                            setSelectedUser(user);
                            setShowRoleModal(true);
                          }}
                          className="p-1.5 hover:bg-gray-700 rounded text-text-muted hover:text-white transition-colors"
                          title="分配角色"
                        >
                          <Shield className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteUser(user.id)}
                          className="p-1.5 hover:bg-error/20 rounded text-text-muted hover:text-error transition-colors"
                          title="删除用户"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        <div className="card">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-white">角色列表</h3>
            <button
              onClick={() => {
                setSelectedRole(null);
                setShowPermissionModal(true);
              }}
              className="btn-primary flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              新建角色
            </button>
          </div>
          <div className="grid grid-cols-3 gap-4">
            {roles.map((role) => (
              <div key={role.id} className="border border-gray-700 rounded-lg p-4 hover:border-gray-600 transition-colors">
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <h4 className="font-medium text-white">{role.name}</h4>
                    <p className="text-sm text-text-muted mt-1">{role.description}</p>
                  </div>
                  {role.isSystem && (
                    <span className="px-2 py-0.5 bg-info/20 text-info rounded text-xs">系统</span>
                  )}
                </div>
                <div className="flex items-center justify-between text-sm text-text-muted">
                  <span>{role.userCount || 0} 用户</span>
                  <span>{role.permissions?.length || 0} 权限</span>
                </div>
                <div className="flex gap-2 mt-4 pt-3 border-t border-gray-700">
                  <button
                    onClick={() => {
                      setSelectedRole(role);
                      setShowPermissionModal(true);
                    }}
                    className="flex-1 flex items-center justify-center gap-1 py-1.5 bg-gray-700 hover:bg-gray-600 rounded text-sm transition-colors"
                  >
                    <Edit className="w-3.5 h-3.5" />
                    编辑
                  </button>
                  <button
                    onClick={() => handleManageRolePermissions(role)}
                    className="flex-1 flex items-center justify-center gap-1 py-1.5 bg-primary/20 hover:bg-primary/30 text-primary rounded text-sm transition-colors"
                  >
                    <Key className="w-3.5 h-3.5" />
                    权限
                  </button>
                  {!role.isSystem && (
                    <button
                      onClick={() => handleDeleteRole(role.id)}
                      className="flex items-center justify-center gap-1 px-3 py-1.5 bg-error/20 hover:bg-error/30 text-error rounded text-sm transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {showRoleModal && selectedUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white">分配角色 - {selectedUser.name}</h3>
              <button
                onClick={() => setShowRoleModal(false)}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {roles.map((role) => (
                <label
                  key={role.id}
                  className="flex items-center gap-3 p-3 rounded-lg border border-gray-700 hover:border-gray-600 cursor-pointer transition-colors"
                >
                  <input
                    type="checkbox"
                    checked={selectedUser.roles.some((r) => r.id === role.id)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setSelectedUser({
                          ...selectedUser,
                          roles: [...selectedUser.roles, role],
                        });
                      } else {
                        setSelectedUser({
                          ...selectedUser,
                          roles: selectedUser.roles.filter((r) => r.id !== role.id),
                        });
                      }
                    }}
                    className="w-4 h-4 rounded border-gray-600 bg-gray-700 text-primary focus:ring-primary"
                  />
                  <div className="flex-1">
                    <div className="text-white font-medium">{role.name}</div>
                    <div className="text-sm text-text-muted">{role.description}</div>
                  </div>
                  {role.isSystem && (
                    <span className="px-2 py-0.5 bg-info/20 text-info rounded text-xs">系统</span>
                  )}
                </label>
              ))}
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowRoleModal(false)}
                className="px-4 py-2 text-text-muted hover:text-white transition-colors"
              >
                取消
              </button>
              <button
                onClick={() => handleAssignRoles(selectedUser.id, selectedUser.roles.map((r) => r.id))}
                className="btn-primary flex items-center gap-2"
              >
                <Check className="w-4 h-4" />
                保存
              </button>
            </div>
          </div>
        </div>
      )}

      {showPermissionModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white">
                {selectedRole ? '编辑角色' : '新建角色'}
              </h3>
              <button
                onClick={() => {
                  setShowPermissionModal(false);
                  setSelectedRole(null);
                }}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-text-muted mb-1">角色名称</label>
                <input
                  type="text"
                  id="roleName"
                  defaultValue={selectedRole?.name}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入角色名称"
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">描述</label>
                <textarea
                  id="roleDesc"
                  defaultValue={selectedRole?.description}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入角色描述"
                  rows={3}
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">权限级别</label>
                <select
                  id="roleLevel"
                  defaultValue={selectedRole?.level || 10}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                >
                  <option value={1}>1 - 管理员</option>
                  <option value={5}>5 - 运维人员</option>
                  <option value={10}>10 - 普通用户</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => {
                  setShowPermissionModal(false);
                  setSelectedRole(null);
                }}
                className="px-4 py-2 text-text-muted hover:text-white transition-colors"
              >
                取消
              </button>
              <button
                onClick={() => {
                  const nameInput = document.getElementById('roleName') as HTMLInputElement;
                  const descInput = document.getElementById('roleDesc') as HTMLTextAreaElement;
                  const levelInput = document.getElementById('roleLevel') as HTMLSelectElement;
                  handleCreateRole({
                    name: nameInput?.value || '新角色',
                    description: descInput?.value || '新创建的角色',
                    level: parseInt(levelInput?.value || '10'),
                    path: '/user',
                  });
                }}
                className="btn-primary flex items-center gap-2"
              >
                <Check className="w-4 h-4" />
                保存
              </button>
            </div>
          </div>
        </div>
      )}

      {showAddUserModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white">新建用户</h3>
              <button
                onClick={() => setShowAddUserModal(false)}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-text-muted mb-1">用户名</label>
                <input
                  type="text"
                  id="newUsername"
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入用户名"
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">姓名</label>
                <input
                  type="text"
                  id="newName"
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入姓名"
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">邮箱</label>
                <input
                  type="email"
                  id="newEmail"
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入邮箱"
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">密码</label>
                <input
                  type="password"
                  id="newPassword"
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入密码"
                />
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowAddUserModal(false)}
                className="px-4 py-2 text-text-muted hover:text-white transition-colors"
              >
                取消
              </button>
              <button
                onClick={() => {
                  const usernameInput = document.getElementById('newUsername') as HTMLInputElement;
                  const nameInput = document.getElementById('newName') as HTMLInputElement;
                  const emailInput = document.getElementById('newEmail') as HTMLInputElement;
                  const passwordInput = document.getElementById('newPassword') as HTMLInputElement;
                  handleCreateUser({
                    username: usernameInput?.value || '',
                    name: nameInput?.value || '',
                    email: emailInput?.value || '',
                    password: passwordInput?.value || '',
                  });
                }}
                className="btn-primary flex items-center gap-2"
              >
                <Check className="w-4 h-4" />
                创建
              </button>
            </div>
          </div>
        </div>
      )}

      {showRolePermissionsModal && selectedRole && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-lg animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white">权限分配 - {selectedRole.name}</h3>
              <button
                onClick={() => setShowRolePermissionsModal(false)}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>
            <div className="max-h-96 overflow-y-auto">
              <div className="space-y-2">
                {permissions.map((perm) => (
                  <label
                    key={perm.id}
                    className="flex items-center gap-3 p-3 rounded-lg border border-gray-700 hover:border-gray-600 cursor-pointer transition-colors"
                  >
                    <input
                      type="checkbox"
                      checked={rolePermissions.includes(perm.id)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setRolePermissions([...rolePermissions, perm.id]);
                        } else {
                          setRolePermissions(rolePermissions.filter(id => id !== perm.id));
                        }
                      }}
                      className="w-4 h-4 rounded border-gray-600 bg-gray-700 text-primary focus:ring-primary"
                    />
                    <div className="flex-1">
                      <div className="text-white font-medium">{perm.name}</div>
                      <div className="text-sm text-text-muted">{perm.description || `${perm.resource}:${perm.action}`}</div>
                    </div>
                    <span className="px-2 py-0.5 bg-gray-700 text-text-muted rounded text-xs">
                      {perm.resource}
                    </span>
                  </label>
                ))}
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowRolePermissionsModal(false)}
                className="px-4 py-2 text-text-muted hover:text-white transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleSaveRolePermissions}
                className="btn-primary flex items-center gap-2"
              >
                <Check className="w-4 h-4" />
                保存
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default UserManagement;
