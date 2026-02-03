
import React from 'react';
import { useTranslation } from './LanguageContext';
import { 
  Bell, 
  CheckCircle2, 
  Clock, 
  AlertTriangle, 
  XCircle,
  Search,
  Settings
} from 'lucide-react';
import { Alert } from '../types';

const MOCK_ALERTS: Alert[] = [
  { id: '1', time: '2024-02-02 08:32:10', vmName: 'prod-api-01', type: 'CPU High', severity: 'critical', message: 'CPU Usage reached 98% for over 5 minutes.', status: 'pending' },
  { id: '2', time: '2024-02-02 08:15:00', vmName: 'stage-db-01', type: 'Disk Warning', severity: 'medium', message: 'Available disk space below 10% (Mounted on /data).', status: 'resolved' },
  { id: '3', time: '2024-02-02 07:45:22', vmName: 'internal-git', type: 'Network Spike', severity: 'low', message: 'Unusual outbound traffic detected (500MB/s).', status: 'ignored' },
];

const AlertManagement: React.FC = () => {
  const { t } = useTranslation();

  const getSeverityStyles = (severity: string) => {
    switch(severity) {
      case 'critical': return 'text-red-500 bg-red-500/10 border-red-500/20';
      case 'high': return 'text-orange-500 bg-orange-500/10 border-orange-500/20';
      case 'medium': return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
      default: return 'text-blue-500 bg-blue-500/10 border-blue-500/20';
    }
  };

  const getStatusIcon = (status: string) => {
    switch(status) {
      case 'resolved': return <CheckCircle2 size={16} className="text-[#00d4aa]" />;
      case 'ignored': return <XCircle size={16} className="text-[#607080]" />;
      default: return <Clock size={16} className="text-orange-400" />;
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <h2 className="text-2xl font-bold flex items-center gap-3">
          <Bell className="text-orange-400" /> {t('alerts')}
        </h2>
        <div className="flex gap-2 w-full md:w-auto">
          <button className="flex-1 md:flex-none flex items-center justify-center gap-2 px-6 py-3 bg-[#1a1f2e] border border-[#2a3441] rounded-xl hover:bg-[#2a3441] transition-all">
            <Settings size={18} /> Config Rules
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-4 gap-6">
        <div className="xl:col-span-3 space-y-4">
          <div className="bg-[#1a1f2e] border border-[#2a3441] rounded-3xl overflow-hidden">
             {MOCK_ALERTS.map((alert) => (
               <div key={alert.id} className="p-6 border-b border-[#2a3441] hover:bg-[#2a3441]/20 transition-colors flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
                 <div className="flex items-start gap-4 flex-1">
                   <div className={`p-3 rounded-2xl border ${getSeverityStyles(alert.severity)}`}>
                     <AlertTriangle size={24} />
                   </div>
                   <div>
                     <div className="flex items-center gap-3 mb-1">
                       <span className="font-bold text-lg">{alert.vmName}</span>
                       <span className="text-xs font-mono text-[#8090a0]">{alert.time}</span>
                     </div>
                     <div className="text-white font-medium mb-1">{alert.type}</div>
                     <p className="text-[#b0b8c5] text-sm">{alert.message}</p>
                   </div>
                 </div>
                 <div className="flex items-center gap-3 w-full md:w-auto">
                   <div className="flex items-center gap-2 bg-[#0f1419] px-4 py-2 rounded-xl border border-[#2a3441] text-xs font-bold uppercase tracking-widest text-[#b0b8c5]">
                     {getStatusIcon(alert.status)} {alert.status}
                   </div>
                   <button className="px-4 py-2 bg-[#2a3441] hover:bg-[#3a4451] rounded-xl text-xs font-bold transition-all">
                     Acknowledge
                   </button>
                 </div>
               </div>
             ))}
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-[#1a1f2e] border border-[#2a3441] p-6 rounded-3xl">
            <h3 className="font-bold mb-4">Summary (Today)</h3>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-[#b0b8c5]">Critical</span>
                <span className="px-2 py-0.5 bg-red-500/10 text-red-500 rounded font-bold text-xs">1</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-[#b0b8c5]">High</span>
                <span className="px-2 py-0.5 bg-orange-500/10 text-orange-500 rounded font-bold text-xs">0</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-[#b0b8c5]">Warning</span>
                <span className="px-2 py-0.5 bg-yellow-500/10 text-yellow-500 rounded font-bold text-xs">2</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-[#b0b8c5]">Info</span>
                <span className="px-2 py-0.5 bg-blue-500/10 text-blue-500 rounded font-bold text-xs">12</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AlertManagement;
