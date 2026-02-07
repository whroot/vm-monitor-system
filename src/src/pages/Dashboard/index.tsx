import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  PieChart, Pie, Cell, ResponsiveContainer,
  XAxis, YAxis, Tooltip, AreaChart, Area
} from 'recharts';
import {
  TrendingUp, TrendingDown,
  Activity, Cpu, HardDrive, Network,
  AlertCircle, ShieldCheck, Server, Clock
} from 'lucide-react';
import { realtimeApi } from '../../api/realtime';
import { DashboardMetrics } from '../../types/api';

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const [mode, setMode] = useState<'normal' | 'fault'>('normal');
  const [overview, setOverview] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchOverview();
    const interval = setInterval(fetchOverview, 30000); // 30秒刷新
    return () => clearInterval(interval);
  }, []);

  const fetchOverview = async () => {
    try {
      const data = await realtimeApi.getOverview();
      setOverview(data);
    } catch (error) {
      console.error('Failed to fetch overview:', error);
    } finally {
      setLoading(false);
    }
  };

  // 模拟数据
  const healthScore = overview?.healthScore?.value || 92;
  
  const metrics = [
    { label: t('cpu'), value: '42.5%', trend: '+2.4%', up: true, icon: Cpu, color: '#2196f3' },
    { label: t('memory'), value: '12.8 GB', trend: '-1.1%', up: false, icon: Activity, color: '#4caf50' },
    { label: t('disk'), value: '78.2%', trend: '+0.5%', up: true, icon: HardDrive, color: '#ff9800' },
    { label: t('network'), value: '1.2 Gbps', trend: '+12%', up: true, icon: Network, color: '#9c27b0' },
  ];

  const vmStatsData = [
    { name: t('online'), value: overview?.vmMonitoring?.onlineVMs || 85, color: '#00d4aa' },
    { name: t('warningStatus'), value: 10, color: '#ff9800' },
    { name: t('errorStatus'), value: overview?.vmMonitoring?.errorVMs || 3, color: '#f44336' },
    { name: t('offline'), value: overview?.vmMonitoring?.offlineVMs || 2, color: '#607d8b' },
  ];

  const performanceData = [
    { time: '00:00', cpu: 32, mem: 45 },
    { time: '04:00', cpu: 45, mem: 52 },
    { time: '08:00', cpu: 82, mem: 75 },
    { time: '12:00', cpu: 55, mem: 60 },
    { time: '16:00', cpu: 65, mem: 68 },
    { time: '20:00', cpu: 42, mem: 55 },
    { time: '23:59', cpu: 38, mem: 48 },
  ];

  const recentAlerts = [
    { id: '1', vm: 'prod-db-01', type: 'High CPU Load', level: 'critical', time: '2 mins ago' },
    { id: '2', vm: 'app-server-cluster', type: 'Disk Pressure', level: 'high', time: '15 mins ago' },
    { id: '3', vm: 'web-server-03', type: 'Memory Usage', level: 'medium', time: '1 hour ago' },
  ];

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-text-muted">{t('loading')}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Mode Switcher */}
      <div className="flex items-center justify-between">
        <div className="flex items-center bg-surface p-1.5 rounded-2xl border border-border w-fit">
          <button
            onClick={() => setMode('normal')}
            className={`px-4 py-2 rounded-xl text-sm font-semibold transition-all ${
              mode === 'normal'
                ? 'bg-success text-white'
                : 'text-text-tertiary hover:text-white'
            }`}
          >
            {t('normalMode')}
          </button>
          <button
            onClick={() => setMode('fault')}
            className={`px-4 py-2 rounded-xl text-sm font-semibold transition-all ${
              mode === 'fault'
                ? 'bg-danger text-white'
                : 'text-text-tertiary hover:text-white'
            }`}
          >
            {t('faultMode')}
          </button>
        </div>

        <div className="flex items-center gap-2 text-text-muted text-sm">
          <Clock className="w-4 h-4" />
          <span>{new Date().toLocaleString()}</span>
        </div>
      </div>

      {mode === 'normal' ? (
        <>
          {/* Health Score & Metrics */}
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            {/* Health Score */}
            <div className="lg:col-span-1 card flex flex-col items-center justify-center relative overflow-hidden">
              <div className="absolute top-0 right-0 p-4 opacity-10">
                <ShieldCheck className="w-32 h-32" />
              </div>
              <div className="text-text-muted text-sm font-semibold mb-2 uppercase tracking-wider">
                {t('healthScore')}
              </div>
              <div className="relative w-40 h-40">
                <svg className="w-full h-full transform -rotate-90">
                  <circle cx="80" cy="80" r="70" fill="transparent" stroke="#2a3441" strokeWidth="12" />
                  <circle
                    cx="80" cy="80" r="70" fill="transparent"
                    stroke={healthScore >= 80 ? '#00d4aa' : healthScore >= 60 ? '#ff9800' : '#f44336'}
                    strokeWidth="12"
                    strokeDasharray={440}
                    strokeDashoffset={440 - (440 * healthScore) / 100}
                    strokeLinecap="round"
                    className="transition-all duration-1000 ease-out"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center flex-col">
                  <span className="text-4xl font-bold text-white">{healthScore}</span>
                  <span className={`text-xs font-bold ${
                    healthScore >= 80 ? 'text-success' : healthScore >= 60 ? 'text-warning' : 'text-danger'
                  }`}>
                    {healthScore >= 80 ? 'EXCELLENT' : healthScore >= 60 ? 'GOOD' : 'WARNING'}
                  </span>
                </div>
              </div>
            </div>

            {/* Metrics */}
            <div className="lg:col-span-3 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
              {metrics.map((m) => (
                <div
                  key={m.label}
                  className="card group hover:border-success/50 transition-all cursor-pointer"
                >
                  <div className="flex justify-between items-start mb-4">
                    <div
                      className="p-3 rounded-2xl bg-opacity-10"
                      style={{ backgroundColor: m.color + '20', color: m.color }}
                    >
                      <m.icon className="w-6 h-6" />
                    </div>
                    <div className={`flex items-center text-xs font-bold ${
                      m.up ? 'text-success' : 'text-danger'
                    }`}>
                      {m.up ? <TrendingUp className="w-3 h-3 mr-1" /> : <TrendingDown className="w-3 h-3 mr-1" />}
                      {m.trend}
                    </div>
                  </div>
                  <div className="text-2xl font-bold text-white mb-1">{m.value}</div>
                  <div className="text-sm text-text-muted">{m.label}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Charts & Alerts */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* VM Status */}
            <div className="card">
              <h3 className="text-lg font-semibold text-white mb-6">{t('vmStatus')}</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={vmStatsData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={80}
                      paddingAngle={5}
                      dataKey="value"
                    >
                      {vmStatsData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1a1f2e',
                        border: '1px solid #2a3441',
                        borderRadius: '8px',
                        color: '#fff'
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="grid grid-cols-2 gap-4 mt-4">
                {vmStatsData.map((stat) => (
                  <div key={stat.name} className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: stat.color }} />
                    <span className="text-sm text-text-secondary">{stat.name}</span>
                    <span className="text-sm font-bold text-white ml-auto">{stat.value}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Performance Trend */}
            <div className="card">
              <h3 className="text-lg font-semibold text-white mb-6">{t('performanceTrend')}</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={performanceData}>
                    <defs>
                      <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#2196f3" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#2196f3" stopOpacity={0} />
                      </linearGradient>
                      <linearGradient id="colorMem" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#4caf50" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#4caf50" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <XAxis dataKey="time" stroke="#607d8b" fontSize={12} />
                    <YAxis stroke="#607d8b" fontSize={12} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1a1f2e',
                        border: '1px solid #2a3441',
                        borderRadius: '8px',
                        color: '#fff'
                      }}
                    />
                    <Area
                      type="monotone"
                      dataKey="cpu"
                      stroke="#2196f3"
                      fillOpacity={1}
                      fill="url(#colorCpu)"
                      strokeWidth={2}
                    />
                    <Area
                      type="monotone"
                      dataKey="mem"
                      stroke="#4caf50"
                      fillOpacity={1}
                      fill="url(#colorMem)"
                      strokeWidth={2}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Recent Alerts */}
            <div className="card">
              <div className="flex justify-between items-center mb-6">
                <h3 className="text-lg font-semibold text-white">{t('recentAlerts')}</h3>
                <button className="text-sm text-info hover:underline">{t('viewAll')}</button>
              </div>
              <div className="space-y-4">
                {recentAlerts.map((alert) => (
                  <div
                    key={alert.id}
                    className="flex items-center gap-4 p-4 bg-background rounded-xl border border-border hover:border-warning/50 transition-all"
                  >
                    <div className={`p-2 rounded-lg ${
                      alert.level === 'critical' ? 'bg-danger/20 text-danger' :
                      alert.level === 'high' ? 'bg-warning/20 text-warning' :
                      'bg-info/20 text-info'
                    }`}>
                      <AlertCircle className="w-5 h-5" />
                    </div>
                    <div className="flex-1">
                      <div className="text-sm font-medium text-white">{alert.vm}</div>
                      <div className="text-xs text-text-muted">{alert.type}</div>
                    </div>
                    <div className="text-xs text-text-muted">{alert.time}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* VM Summary */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            <div className="card text-center">
              <Server className="w-8 h-8 text-info mx-auto mb-2" />
              <div className="text-3xl font-bold text-white">{overview?.vmMonitoring?.totalVMs || 150}</div>
              <div className="text-sm text-text-muted">{t('totalVMs')}</div>
            </div>
            <div className="card text-center">
              <div className="w-8 h-8 bg-success/20 rounded-lg flex items-center justify-center mx-auto mb-2">
                <Server className="w-5 h-5 text-success" />
              </div>
              <div className="text-3xl font-bold text-success">{overview?.vmMonitoring?.onlineVMs || 140}</div>
              <div className="text-sm text-text-muted">{t('onlineVMs')}</div>
            </div>
            <div className="card text-center">
              <div className="w-8 h-8 bg-danger/20 rounded-lg flex items-center justify-center mx-auto mb-2">
                <Server className="w-5 h-5 text-danger" />
              </div>
              <div className="text-3xl font-bold text-danger">{overview?.vmMonitoring?.offlineVMs || 5}</div>
              <div className="text-sm text-text-muted">{t('offlineVMs')}</div>
            </div>
            <div className="card text-center">
              <div className="w-8 h-8 bg-warning/20 rounded-lg flex items-center justify-center mx-auto mb-2">
                <AlertCircle className="w-5 h-5 text-warning" />
              </div>
              <div className="text-3xl font-bold text-warning">
                {(overview?.alerts?.critical || 0) + (overview?.alerts?.high || 0)}
              </div>
              <div className="text-sm text-text-muted">{t('alertVMs')}</div>
            </div>
          </div>
        </>
      ) : (
        // Fault Mode
        <div className="space-y-6">
          <div className="bg-danger/10 border border-danger/20 rounded-2xl p-6">
            <div className="flex items-center gap-4">
              <AlertCircle className="w-8 h-8 text-danger" />
              <div>
                <h2 className="text-xl font-bold text-danger">检测到严重告警</h2>
                <p className="text-text-muted">当前有 {recentAlerts.length} 个VM存在问题需要处理</p>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {recentAlerts.map((alert) => (
              <div
                key={alert.id}
                className="bg-danger/5 border border-danger/20 rounded-2xl p-6 hover:border-danger/50 transition-all"
              >
                <div className="flex items-start gap-4">
                  <div className="p-3 bg-danger/20 rounded-xl">
                    <Server className="w-6 h-6 text-danger" />
                  </div>
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-white">{alert.vm}</h3>
                    <p className="text-danger font-medium mt-1">{alert.type}</p>
                    <p className="text-sm text-text-muted mt-2">{alert.time}</p>
                    <div className="flex gap-2 mt-4">
                      <button className="px-4 py-2 bg-danger text-white rounded-lg text-sm font-medium hover:bg-danger/90">
                        查看详情
                      </button>
                      <button className="px-4 py-2 bg-surface border border-border text-text-secondary rounded-lg text-sm font-medium hover:border-text-tertiary">
                        处理
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default Dashboard;