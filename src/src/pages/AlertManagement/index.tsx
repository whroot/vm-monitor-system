import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { AlertCircle, Plus, Bell, Check, X, Clock, Server } from 'lucide-react';
import { alertApi, AlertRule, AlertRecord, AlertStats } from '../../api/alert';

type AlertTab = 'records' | 'rules';

const AlertManagement: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<AlertTab>('records');
  const [stats, setStats] = useState<AlertStats | null>(null);
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [records, setRecords] = useState<AlertRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [showRuleModal, setShowRuleModal] = useState(false);
  const [selectedRule, setSelectedRule] = useState<AlertRule | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [statsData, rulesData, recordsData] = await Promise.all([
        alertApi.getStats(),
        alertApi.listRules(),
        alertApi.listRecords(),
      ]);
      setStats(statsData);
      setRules(rulesData.rules || []);
      setRecords(recordsData.records || []);
    } catch (error) {
      console.error('Failed to load alert data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAcknowledge = async (alertId: string) => {
    try {
      await alertApi.acknowledge(alertId);
      await loadData();
    } catch (error) {
      console.error('Failed to acknowledge alert:', error);
    }
  };

  const handleResolve = async (alertId: string) => {
    const resolution = prompt('请输入解决说明:');
    if (resolution === null) return;
    try {
      await alertApi.resolve(alertId, resolution || '已解决');
      await loadData();
    } catch (error) {
      console.error('Failed to resolve alert:', error);
    }
  };

  const handleToggleRule = async (rule: AlertRule) => {
    try {
      await alertApi.updateRule(rule.ID, { Enabled: !rule.Enabled });
      await loadData();
    } catch (error) {
      console.error('Failed to toggle rule:', error);
    }
  };

  const handleDeleteRule = async (ruleId: string) => {
    if (!confirm('确定要删除此告警规则吗？')) return;
    try {
      await alertApi.deleteRule(ruleId);
      await loadData();
    } catch (error) {
      console.error('Failed to delete rule:', error);
    }
  };

  const handleCreateRule = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    try {
      await alertApi.createRule({
        name: formData.get('name') as string,
        scope: formData.get('scope') as string,
        severity: formData.get('severity') as string,
      });
      await loadData();
      setShowRuleModal(false);
    } catch (error) {
      console.error('Failed to create rule:', error);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'danger';
      case 'high': return 'warning';
      case 'medium': return 'info';
      default: return 'success';
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active': return <span className="px-2 py-0.5 bg-danger/20 text-danger rounded text-xs">活跃</span>;
      case 'acknowledged': return <span className="px-2 py-0.5 bg-warning/20 text-warning rounded text-xs">已确认</span>;
      case 'resolved': return <span className="px-2 py-0.5 bg-success/20 text-success rounded text-xs">已解决</span>;
      default: return null;
    }
  };

  if (loading) {
    return <div className="text-white p-8">加载中...</div>;
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
          <Bell className="w-6 h-6" />
          {t('alerts')}
        </h1>
        <button
          onClick={() => setShowRuleModal(true)}
          className="btn-primary flex items-center gap-2"
        >
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
          <div className="text-3xl font-bold text-danger">{stats?.Critical || 0}</div>
          <div className="text-sm text-text-muted">严重告警</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-warning/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <AlertCircle className="w-6 h-6 text-warning" />
          </div>
          <div className="text-3xl font-bold text-warning">{stats?.Warning || 0}</div>
          <div className="text-sm text-text-muted">高优先级</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-info/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <AlertCircle className="w-6 h-6 text-info" />
          </div>
          <div className="text-3xl font-bold text-info">{(stats?.Active || 0) - (stats?.Critical || 0) - (stats?.Warning || 0)}</div>
          <div className="text-sm text-text-muted">中等优先级</div>
        </div>
        <div className="card text-center">
          <div className="w-12 h-12 bg-success/20 rounded-xl flex items-center justify-center mx-auto mb-3">
            <AlertCircle className="w-6 h-6 text-success" />
          </div>
          <div className="text-3xl font-bold text-success">{stats?.Total || rules.length}</div>
          <div className="text-sm text-text-muted">规则总数</div>
        </div>
      </div>

      {/* Alert List */}
      <div className="card">
        <div className="flex gap-4 mb-6">
          <button
            onClick={() => setActiveTab('records')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === 'records'
                ? 'bg-success text-white'
                : 'bg-surface border border-border text-text-secondary hover:text-white'
            }`}
          >
            告警记录
          </button>
          <button
            onClick={() => setActiveTab('rules')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === 'rules'
                ? 'bg-success text-white'
                : 'bg-surface border border-border text-text-secondary hover:text-white'
            }`}
          >
            告警规则
          </button>
        </div>

        {/* Alert Records Tab */}
        {activeTab === 'records' && (
          <div className="overflow-x-auto">
            {records.length === 0 ? (
              <div className="h-64 flex items-center justify-center text-text-muted">
                暂无告警记录
              </div>
            ) : (
              <table className="w-full">
                <thead>
                  <tr className="text-left text-text-muted text-sm border-b border-gray-700">
                    <th className="pb-3 font-medium">告警名称</th>
                    <th className="pb-3 font-medium">虚拟机</th>
                    <th className="pb-3 font-medium">指标</th>
                    <th className="pb-3 font-medium">触发值</th>
                    <th className="pb-3 font-medium">阈值</th>
                    <th className="pb-3 font-medium">级别</th>
                    <th className="pb-3 font-medium">状态</th>
                    <th className="pb-3 font-medium">触发时间</th>
                    <th className="pb-3 font-medium text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {records.map((record) => (
                    <tr key={record.id} className="border-b border-gray-700/50 hover:bg-gray-700/20">
                      <td className="py-3 text-white font-medium">{record.ruleName}</td>
                      <td className="py-3 text-text-muted flex items-center gap-2">
                        <Server className="w-3.5 h-3.5" />
                        {record.vmName || record.vmId}
                      </td>
                      <td className="py-3 text-text-muted">{record.metric}</td>
                      <td className="py-3 text-warning">{record.triggerValue}%</td>
                      <td className="py-3 text-text-muted">{record.threshold}%</td>
                      <td className="py-3">
                        <span className={`px-2 py-0.5 rounded text-xs ${
                          getSeverityColor(record.severity) === 'danger' ? 'bg-danger/20 text-danger' :
                          getSeverityColor(record.severity) === 'warning' ? 'bg-warning/20 text-warning' :
                          'bg-info/20 text-info'
                        }`}>
                          {record.severity}
                        </span>
                      </td>
                      <td className="py-3">{getStatusBadge(record.status)}</td>
                      <td className="py-3 text-text-muted text-sm">
                        {new Date(record.triggeredAt).toLocaleString('zh-CN')}
                      </td>
                      <td className="py-3 text-right">
                        <div className="flex items-center justify-end gap-1">
                          {record.status === 'active' && (
                            <button
                              onClick={() => handleAcknowledge(record.id)}
                              className="p-1.5 hover:bg-warning/20 rounded text-warning hover:text-warning transition-colors"
                              title="确认"
                            >
                              <Check className="w-4 h-4" />
                            </button>
                          )}
                          {record.status !== 'resolved' && (
                            <button
                              onClick={() => handleResolve(record.id)}
                              className="p-1.5 hover:bg-success/20 rounded text-success hover:text-success transition-colors"
                              title="解决"
                            >
                              <AlertCircle className="w-4 h-4" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {/* Alert Rules Tab */}
        {activeTab === 'rules' && (
          <div className="overflow-x-auto">
            {rules.length === 0 ? (
              <div className="h-64 flex items-center justify-center text-text-muted">
                暂无告警规则，点击"新建规则"添加
              </div>
            ) : (
              <table className="w-full">
                <thead>
                  <tr className="text-left text-text-muted text-sm border-b border-gray-700">
                    <th className="pb-3 font-medium">规则名称</th>
                    <th className="pb-3 font-medium">描述</th>
                    <th className="pb-3 font-medium">范围</th>
                    <th className="pb-3 font-medium">级别</th>
                    <th className="pb-3 font-medium">冷却时间</th>
                    <th className="pb-3 font-medium">状态</th>
                    <th className="pb-3 font-medium">创建时间</th>
                    <th className="pb-3 font-medium text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {rules.map((rule) => (
                    <tr key={rule.ID} className="border-b border-gray-700/50 hover:bg-gray-700/20">
                      <td className="py-3 text-white font-medium">{rule.Name}</td>
                      <td className="py-3 text-text-muted">{rule.Description || '-'}</td>
                      <td className="py-3 text-text-muted">{rule.Scope}</td>
                      <td className="py-3">
                        <span className={`px-2 py-0.5 rounded text-xs ${
                          getSeverityColor(rule.Severity) === 'danger' ? 'bg-danger/20 text-danger' :
                          getSeverityColor(rule.Severity) === 'warning' ? 'bg-warning/20 text-warning' :
                          'bg-info/20 text-info'
                        }`}>
                          {rule.Severity}
                        </span>
                      </td>
                      <td className="py-3 text-text-muted">{rule.Cooldown}秒</td>
                      <td className="py-3">
                        <button
                          onClick={() => handleToggleRule(rule)}
                          className={`px-2 py-0.5 rounded text-xs transition-colors ${
                            rule.Enabled
                              ? 'bg-success/20 text-success'
                              : 'bg-gray-700 text-text-muted'
                          }`}
                        >
                          {rule.Enabled ? '启用' : '禁用'}
                        </button>
                      </td>
                      <td className="py-3 text-text-muted text-sm">
                        {new Date(rule.CreatedAt).toLocaleDateString('zh-CN')}
                      </td>
                      <td className="py-3 text-right">
                        <button
                          onClick={() => handleDeleteRule(rule.ID)}
                          className="p-1.5 hover:bg-error/20 rounded text-text-muted hover:text-error transition-colors"
                          title="删除"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}
      </div>

      {/* Create Rule Modal */}
      {showRuleModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white">新建告警规则</h3>
              <button
                onClick={() => setShowRuleModal(false)}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>
            <form onSubmit={handleCreateRule} className="space-y-4">
              <div>
                <label className="block text-sm text-text-muted mb-1">规则名称</label>
                <input
                  type="text"
                  name="name"
                  required
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入规则名称"
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">监控范围</label>
                <select
                  name="scope"
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                >
                  <option value="vm">虚拟机</option>
                  <option value="host">宿主机</option>
                  <option value="cluster">集群</option>
                </select>
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">告警级别</label>
                <select
                  name="severity"
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                >
                  <option value="critical">严重</option>
                  <option value="high">高</option>
                  <option value="medium">中</option>
                  <option value="low">低</option>
                </select>
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={() => setShowRuleModal(false)}
                  className="px-4 py-2 text-text-muted hover:text-white transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="btn-primary flex items-center gap-2"
                >
                  <Plus className="w-4 h-4" />
                  创建
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default AlertManagement;
