import React from 'react';
import { useTranslation } from 'react-i18next';
import { Settings, Database, Shield, Info } from 'lucide-react';

const SystemSettings: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="space-y-6 animate-fade-in">
      <h1 className="text-2xl font-bold text-white flex items-center gap-3">
        <Settings className="w-6 h-6" />
        {t('system')}
      </h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Settings Menu */}
        <div className="card lg:col-span-1">
          <nav className="space-y-2">
            <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl bg-success text-white">
              <Database className="w-5 h-5" />
              <span className="font-medium">基本设置</span>
            </button>
            <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-text-tertiary hover:bg-surface hover:text-white transition-all">
              <Shield className="w-5 h-5" />
              <span className="font-medium">安全设置</span>
            </button>
            <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-text-tertiary hover:bg-surface hover:text-white transition-all">
              <Info className="w-5 h-5" />
              <span className="font-medium">关于系统</span>
            </button>
          </nav>
        </div>

        {/* Settings Content */}
        <div className="card lg:col-span-2">
          <h3 className="text-lg font-semibold text-white mb-6">系统配置</h3>
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                采集间隔（秒）
              </label>
              <input type="number" defaultValue={30} className="input max-w-xs" />
            </div>
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                数据保留天数
              </label>
              <input type="number" defaultValue={730} className="input max-w-xs" />
            </div>
            <div className="flex justify-end">
              <button className="btn-primary">
                {t('save')}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SystemSettings;