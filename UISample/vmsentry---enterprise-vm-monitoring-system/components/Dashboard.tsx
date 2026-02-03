
import React, { useState, useMemo } from 'react';
import { useTranslation } from './LanguageContext';
import { 
  PieChart, Pie, Cell, ResponsiveContainer, 
  LineChart, Line, XAxis, YAxis, Tooltip, AreaChart, Area 
} from 'recharts';
import { 
  TrendingUp, TrendingDown, 
  Activity, Cpu, HardDrive, Network, 
  AlertCircle, ShieldCheck
} from 'lucide-react';
import { DashboardMode, VMStatus } from '../types';

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const [mode, setMode] = useState<DashboardMode>('normal');

  const healthScore = 92;
  const metrics = [
    { label: t('cpu'), value: '42.5%', trend: '+2.4%', up: true, icon: Cpu, color: '#2196f3' },
    { label: t('memory'), value: '12.8 GB', trend: '-1.1%', up: false, icon: Activity, color: '#4caf50' },
    { label: t('disk'), value: '78.2%', trend: '+0.5%', up: true, icon: HardDrive, color: '#ff9800' },
    { label: t('network'), value: '1.2 Gbps', trend: '+12%', up: true, icon: Network, color: '#9c27b0' },
  ];

  const vmStatsData = [
    { name: 'Online', value: 85, color: '#00d4aa' },
    { name: 'Warning', value: 10, color: '#ff9800' },
    { name: 'Critical', value: 3, color: '#f44336' },
    { name: 'Offline', value: 2, color: '#607d8b' },
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

  const faults = [
    { id: '1', vm: 'prod-db-01', type: 'High CPU Load', level: 'critical', time: '2 mins ago' },
    { id: '2', vm: 'app-server-cluster', type: 'Disk Pressure', level: 'high', time: '15 mins ago' },
  ];

  return (
    <div className="space-y-6">
      {/* Mode Switcher */}
      <div className="flex items-center justify-between bg-[#1a1f2e] p-2 rounded-2xl border border-[#2a3441] w-fit">
        <button 
          onClick={() => setMode('normal')}
          className={`px-4 py-2 rounded-xl text-sm font-semibold transition-all ${mode === 'normal' ? 'bg-[#4caf50] text-white' : 'text-[#b0b8c5]'}`}
        >
          {t('normalMode')}
        </button>
        <button 
          onClick={() => setMode('fault')}
          className={`px-4 py-2 rounded-xl text-sm font-semibold transition-all ${mode === 'fault' ? 'bg-[#f44336] text-white' : 'text-[#b0b8c5]'}`}
        >
          {t('faultMode')}
        </button>
      </div>

      {mode === 'normal' ? (
        <>
          {/* Health and Metrics */}
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            <div className="lg:col-span-1 bg-[#1a1f2e] border border-[#2a3441] p-6 rounded-3xl flex flex-col items-center justify-center relative overflow-hidden">
              <div className="absolute top-0 right-0 p-4 opacity-10">
                <ShieldCheck size={120} />
              </div>
              <div className="text-[#8090a0] text-sm font-semibold mb-2 uppercase tracking-wider">{t('healthScore')}</div>
              <div className="relative w-40 h-40">
                <svg className="w-full h-full transform -rotate-90">
                  <circle cx="80" cy="80" r="70" fill="transparent" stroke="#2a3441" strokeWidth="12" />
                  <circle 
                    cx="80" cy="80" r="70" fill="transparent" stroke="#4caf50" strokeWidth="12" 
                    strokeDasharray={440} strokeDashoffset={440 - (440 * healthScore) / 100}
                    strokeLinecap="round"
                    className="transition-all duration-1000 ease-out"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center flex-col">
                  <span className="text-4xl font-bold">{healthScore}</span>
                  <span className="text-xs text-[#00d4aa]">EXCELLENT</span>
                </div>
              </div>
            </div>

            <div className="lg:col-span-3 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
              {metrics.map((m) => (
                <div key={m.label} className="bg-[#1a1f2e] border border-[#2a3441] p-6 rounded-3xl group hover:border-[#4caf50]/50 transition-all">
                  <div className="flex justify-between items-start mb-4">
                    <div className={`p-3 rounded-2xl bg-opacity-10`} style={{ backgroundColor: m.color, color: m.color }}>
                      <m.icon size={24} />
                    </div>
                    <div className={`flex items-center text-xs font-bold ${m.up ? 'text-green-400' : 'text-red-400'}`}>
                      {m.trend} {m.up ? <TrendingUp size={14} className="ml-1" /> : <TrendingDown size={14} className="ml-1" />}
                    </div>
                  </div>
                  <div className="text-[#8090a0] text-sm font-medium mb-1">{m.label}</div>
                  <div className="text-2xl font-bold tracking-tight">{m.value}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Charts Row */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 bg-[#1a1f2e] border border-[#2a3441] p-6 rounded-3xl">
              <h3 className="text-lg font-bold mb-6 flex items-center gap-2">
                <Activity size={20} className="text-[#4caf50]" />
                System Performance Trend
              </h3>
              <div className="h-[300px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={performanceData}>
                    <defs>
                      <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#2196f3" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#2196f3" stopOpacity={0}/>
                      </linearGradient>
                      <linearGradient id="colorMem" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#4caf50" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#4caf50" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <XAxis dataKey="time" stroke="#607080" fontSize={12} tickLine={false} axisLine={false} />
                    <YAxis stroke="#607080" fontSize={12} tickLine={false} axisLine={false} />
                    <Tooltip contentStyle={{ backgroundColor: '#1a1f2e', border: '1px solid #2a3441', borderRadius: '12px' }} />
                    <Area type="monotone" dataKey="cpu" stroke="#2196f3" fillOpacity={1} fill="url(#colorCpu)" strokeWidth={3} />
                    <Area type="monotone" dataKey="mem" stroke="#4caf50" fillOpacity={1} fill="url(#colorMem)" strokeWidth={3} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="bg-[#1a1f2e] border border-[#2a3441] p-6 rounded-3xl">
              <h3 className="text-lg font-bold mb-6">VM Status Distribution</h3>
              <div className="h-[250px]">
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
                    <Tooltip contentStyle={{ backgroundColor: '#1a1f2e', border: '1px solid #2a3441', borderRadius: '12px' }} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="mt-4 space-y-2">
                {vmStatsData.map((s) => (
                  <div key={s.name} className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                      <div className="w-2 h-2 rounded-full" style={{ backgroundColor: s.color }}></div>
                      <span className="text-[#b0b8c5]">{s.name}</span>
                    </div>
                    <span className="font-bold">{s.value}%</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </>
      ) : (
        /* Fault Mode View */
        <div className="space-y-6">
          <div className="bg-red-500/10 border border-red-500/50 p-4 rounded-2xl flex items-center gap-4 animate-pulse">
            <AlertCircle className="text-red-500" size={24} />
            <div>
              <div className="font-bold text-red-500">Critical Fault Mode Active</div>
              <div className="text-sm text-red-400">System detected 2 critical alerts and 5 high warnings.</div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {faults.map(f => (
              <div key={f.id} className="bg-[#1a1f2e] border-l-4 border-red-500 p-6 rounded-2xl shadow-xl flex justify-between items-center">
                <div>
                  <div className="text-[#8090a0] text-xs font-bold uppercase mb-1">{f.time}</div>
                  <div className="text-xl font-bold mb-1">{f.vm}</div>
                  <div className="text-red-400 font-medium">{f.type}</div>
                </div>
                <div className="flex flex-col gap-2">
                  <button className="px-4 py-2 bg-red-500 rounded-xl font-bold text-sm">Restart VM</button>
                  <button className="px-4 py-2 bg-[#2a3441] rounded-xl font-bold text-sm">Details</button>
                </div>
              </div>
            ))}
          </div>

          <div className="bg-[#1a1f2e] border border-[#2a3441] rounded-3xl p-8 text-center">
             <div className="max-w-md mx-auto">
                <AlertCircle size={48} className="mx-auto mb-4 text-[#8090a0]" />
                <h2 className="text-2xl font-bold mb-2">Everything looks under control</h2>
                <p className="text-[#b0b8c5]">No more critical faults detected for other machines. Switch back to normal mode for overall status.</p>
             </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Dashboard;
