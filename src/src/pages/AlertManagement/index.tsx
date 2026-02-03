import React from 'react';
import { useTranslation } from 'react-i18next';
import { AlertCircle, Plus, Bell } from 'lucide-react';

const AlertManagement: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
          <Bell className="w-6 h-6" />
          {t('alerts')}
        </h1>
        <button className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          新建规则
        </button>
      </div>

      {/* Alert Stats */}
      <div className="grid grid-cols-4 gap-6">
        <div className="card text-center">
          <div className="w-12 h-12 bg-danger/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <AlertCircle className="w-6 h-6 text-danger" />
          </div>
          <div className="text-3xl font-bold text-danger">3</div>
          <div className="text-sm text-text-muted">严重告警</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-warning/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <AlertCircle className="w-6 h-6 text-warning" />
          </div>
          <div className="text-3xl font-bold text-warning">8</div>
          <div className="text-sm text-text-muted">高优先级</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-info/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <AlertCircle className="w-6 h-6 text-info" />
          </div>
          <div className="text-3xl font-bold text-info">15</div>
          <div className="text-sm text-text-muted">中等优先级</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-success/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <AlertCircle className="w-6 h-6 text-success" />
          </div>
          <div className="text-3xl font-bold text-success">45</div>
          <div className="text-sm text-text-muted">规则总数</div>
        </div>
      </div>

      {/* Alert List Placeholder */}
      <div className="card">
        <div className="flex gap-4 mb-6">
          <button className="px-4 py-2 bg-success text-white rounded-lg text-sm font-medium">
            告警记录
          </button>
          <button className="px-4 py-2 bg-surface border border-border text-text-secondary rounded-lg text-sm font-medium">
            告警规则
          </button>
        </div>
        <div className="h-64 flex items-center justify-center text-text-muted">
          告警列表区域
        </div>
      </div>
    </div>
  );
};

export default AlertManagement;