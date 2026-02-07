import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Settings, Database, Shield, Info, Save, Check } from 'lucide-react';

type SettingsTab = 'basic' | 'security' | 'about';

interface SystemConfig {
  interval: number;
  retentionDays: number;
  sessionTimeout: number;
  maxLoginAttempts: number;
  passwordExpiry: number;
}

const defaultConfig: SystemConfig = {
  interval: 30,
  retentionDays: 730,
  sessionTimeout: 60,
  maxLoginAttempts: 5,
  passwordExpiry: 90,
};

const STORAGE_KEY = 'vm_monitor_settings';

const SystemSettings: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<SettingsTab>('basic');
  const [saved, setSaved] = useState(false);
  const [config, setConfig] = useState<SystemConfig>(defaultConfig);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const savedConfig = localStorage.getItem(STORAGE_KEY);
    if (savedConfig) {
      try {
        setConfig({ ...defaultConfig, ...JSON.parse(savedConfig) });
      } catch (e) {
        console.error('Failed to parse settings:', e);
      }
    }
    setLoading(false);
  }, []);

  const handleSave = () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const handleReset = () => {
    if (confirm('确定要恢复默认设置吗？')) {
      setConfig(defaultConfig);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(defaultConfig));
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    }
  };

  const handleChange = (key: keyof SystemConfig, value: number) => {
    setConfig({ ...config, [key]: value });
    setSaved(false);
  };

  if (loading) {
    return <div className="text-white">加载中...</div>;
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <h1 className="text-2xl font-bold text-white flex items-center gap-3">
        <Settings className="w-6 h-6" />
        {t('system')}
      </h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="card lg:col-span-1">
          <nav className="space-y-2">
            <button
              onClick={() => setActiveTab('basic')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                activeTab === 'basic'
                  ? 'bg-success text-white'
                  : 'text-text-tertiary hover:bg-surface hover:text-white'
              }`}
            >
              <Database className="w-5 h-5" />
              <span className="font-medium">基本设置</span>
            </button>
            <button
              onClick={() => setActiveTab('security')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                activeTab === 'security'
                  ? 'bg-success text-white'
                  : 'text-text-tertiary hover:bg-surface hover:text-white'
              }`}
            >
              <Shield className="w-5 h-5" />
              <span className="font-medium">安全设置</span>
            </button>
            <button
              onClick={() => setActiveTab('about')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                activeTab === 'about'
                  ? 'bg-success text-white'
                  : 'text-text-tertiary hover:bg-surface hover:text-white'
              }`}
            >
              <Info className="w-5 h-5" />
              <span className="font-medium">关于系统</span>
            </button>
          </nav>
        </div>

        <div className="card lg:col-span-2">
          {activeTab === 'basic' && (
            <>
              <h3 className="text-lg font-semibold text-white mb-6">系统配置</h3>
              <div className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-text-muted mb-2">
                    采集间隔（秒）
                  </label>
                  <input
                    type="number"
                    value={config.interval}
                    onChange={(e) => handleChange('interval', parseInt(e.target.value) || 30)}
                    className="input max-w-xs"
                    min="10"
                    max="300"
                  />
                  <p className="text-xs text-text-muted mt-1">虚拟机监控数据的采集间隔，范围10-300秒</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-text-muted mb-2">
                    数据保留天数
                  </label>
                  <input
                    type="number"
                    value={config.retentionDays}
                    onChange={(e) => handleChange('retentionDays', parseInt(e.target.value) || 730)}
                    className="input max-w-xs"
                    min="30"
                    max="3650"
                  />
                  <p className="text-xs text-text-muted mt-1">历史监控数据的保留时间，范围30-3650天</p>
                </div>
                <div className="flex justify-end gap-3">
                  <button onClick={handleReset} className="btn-secondary">
                    恢复默认
                  </button>
                  <button onClick={handleSave} className="btn-primary flex items-center gap-2">
                    {saved ? <Check className="w-4 h-4" /> : <Save className="w-4 h-4" />}
                    {saved ? '已保存' : '保存配置'}
                  </button>
                </div>
              </div>
            </>
          )}

          {activeTab === 'security' && (
            <>
              <h3 className="text-lg font-semibold text-white mb-6">安全配置</h3>
              <div className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-text-muted mb-2">
                    会话超时（分钟）
                  </label>
                  <input
                    type="number"
                    value={config.sessionTimeout}
                    onChange={(e) => handleChange('sessionTimeout', parseInt(e.target.value) || 60)}
                    className="input max-w-xs"
                    min="5"
                    max="480"
                  />
                  <p className="text-xs text-text-muted mt-1">用户无操作后自动登出的时间，范围5-480分钟</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-text-muted mb-2">
                    最大登录尝试次数
                  </label>
                  <input
                    type="number"
                    value={config.maxLoginAttempts}
                    onChange={(e) => handleChange('maxLoginAttempts', parseInt(e.target.value) || 5)}
                    className="input max-w-xs"
                    min="3"
                    max="10"
                  />
                  <p className="text-xs text-text-muted mt-1">超过此次数后账户将被锁定，范围3-10次</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-text-muted mb-2">
                    密码过期天数
                  </label>
                  <input
                    type="number"
                    value={config.passwordExpiry}
                    onChange={(e) => handleChange('passwordExpiry', parseInt(e.target.value) || 90)}
                    className="input max-w-xs"
                    min="30"
                    max="365"
                  />
                  <p className="text-xs text-text-muted mt-1">强制要求更改密码的周期，范围30-365天</p>
                </div>
                <div className="flex justify-end gap-3">
                  <button onClick={handleReset} className="btn-secondary">
                    恢复默认
                  </button>
                  <button onClick={handleSave} className="btn-primary flex items-center gap-2">
                    {saved ? <Check className="w-4 h-4" /> : <Save className="w-4 h-4" />}
                    {saved ? '已保存' : '保存配置'}
                  </button>
                </div>
              </div>
            </>
          )}

          {activeTab === 'about' && (
            <>
              <h3 className="text-lg font-semibold text-white mb-6">关于系统</h3>
              <div className="space-y-4 text-text-muted">
                <div className="p-4 bg-surface rounded-lg">
                  <p className="text-white font-medium text-lg mb-2">VM监控系统</p>
                  <p className="text-sm">版本 1.0.0</p>
                  <p className="text-sm mt-2">一个用于监控和管理虚拟机的综合平台</p>
                </div>
                <div className="pt-4 border-t border-gray-700">
                  <p className="text-sm font-medium text-white mb-2">技术栈</p>
                  <ul className="list-disc list-inside mt-2 space-y-1 text-sm">
                    <li>后端：Go + Gin + PostgreSQL</li>
                    <li>前端：React + TypeScript + Vite</li>
                    <li>ORM：GORM</li>
                    <li>认证：JWT</li>
                  </ul>
                </div>
                <div className="pt-4 border-t border-gray-700">
                  <p className="text-sm font-medium text-white mb-2">当前配置</p>
                  <div className="grid grid-cols-2 gap-2 text-sm mt-2">
                    <span>采集间隔：</span>
                    <span>{config.interval}秒</span>
                    <span>数据保留：</span>
                    <span>{config.retentionDays}天</span>
                    <span>会话超时：</span>
                    <span>{config.sessionTimeout}分钟</span>
                    <span>密码过期：</span>
                    <span>{config.passwordExpiry}天</span>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default SystemSettings;