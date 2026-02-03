import React from 'react';
import { useTranslation } from 'react-i18next';
import { Users, Plus, Shield } from 'lucide-react';

const UserManagement: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
          <Users className="w-6 h-6" />
          {t('users')}
        </h1>
        <button className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          新建用户
        </button>
      </div>

      {/* User Stats */}
      <div className="grid grid-cols-4 gap-6">
        <div className="card text-center">
          <div className="w-12 h-12 bg-success/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <Users className="w-6 h-6 text-success" />
          </div>
          <div className="text-3xl font-bold text-white">45</div>
          <div className="text-sm text-text-muted">总用户数</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-info/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <Shield className="w-6 h-6 text-info" />
          </div>
          <div className="text-3xl font-bold text-white">5</div>
          <div className="text-sm text-text-muted">角色数</div>
        </div>
      </div>

      {/* User List Placeholder */}
      <div className="card">
        <div className="h-96 flex items-center justify-center text-text-muted">
          用户列表区域
        </div>
      </div>
    </div>
  );
};

export default UserManagement;