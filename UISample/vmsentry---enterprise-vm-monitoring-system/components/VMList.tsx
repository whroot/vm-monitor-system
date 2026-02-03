
import React from 'react';
import { useTranslation } from './LanguageContext';
import { MoreVertical, Search, Filter, Monitor } from 'lucide-react';
import { VMStatus, VMData } from '../types';

const MOCK_VMS: VMData[] = [
  { id: '1', name: 'prod-web-lb', ip: '192.168.1.10', status: VMStatus.ONLINE, cpu: 12, memory: 45, disk: 30, network: 120, os: 'CentOS 7.9', uptime: '124d 14h' },
  { id: '2', name: 'stage-db-sql', ip: '192.168.1.22', status: VMStatus.WARNING, cpu: 85, memory: 92, disk: 15, network: 45, os: 'Windows Server 2022', uptime: '12d 6h' },
  { id: '3', name: 'dev-node-01', ip: '192.168.1.45', status: VMStatus.ONLINE, cpu: 5, memory: 12, disk: 8, network: 12, os: 'Ubuntu 22.04', uptime: '4d 2h' },
  { id: '4', name: 'backup-srv', ip: '192.168.1.100', status: VMStatus.CRITICAL, cpu: 98, memory: 95, disk: 99, network: 2, os: 'Debian 11', uptime: '1y 22d' },
  { id: '5', name: 'internal-wiki', ip: '10.0.4.15', status: VMStatus.OFFLINE, cpu: 0, memory: 0, disk: 45, network: 0, os: 'Ubuntu 20.04', uptime: '0s' },
];

const VMList: React.FC = () => {
  const { t } = useTranslation();

  const getStatusColor = (status: VMStatus) => {
    switch(status) {
      case VMStatus.ONLINE: return 'text-[#00d4aa] bg-[#00d4aa]/10';
      case VMStatus.WARNING: return 'text-[#ff9800] bg-[#ff9800]/10';
      case VMStatus.CRITICAL: return 'text-[#f44336] bg-[#f44336]/10';
      default: return 'text-[#b0b8c5] bg-[#b0b8c5]/10';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
        <div className="relative w-full md:w-96">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-[#607080]" size={18} />
          <input 
            type="text" 
            placeholder="Search VMs..." 
            className="w-full bg-[#1a1f2e] border border-[#2a3441] rounded-xl py-3 pl-12 pr-4 focus:border-[#4caf50] focus:outline-none transition-all"
          />
        </div>
        <div className="flex gap-2 w-full md:w-auto">
          <button className="flex-1 md:flex-none flex items-center justify-center gap-2 px-6 py-3 bg-[#1a1f2e] border border-[#2a3441] rounded-xl hover:bg-[#2a3441] transition-all">
            <Filter size={18} /> Filter
          </button>
          <button className="flex-1 md:flex-none px-6 py-3 bg-[#4caf50] rounded-xl font-bold hover:shadow-lg hover:shadow-[#4caf50]/20 transition-all">
            + New VM
          </button>
        </div>
      </div>

      <div className="bg-[#1a1f2e] border border-[#2a3441] rounded-3xl overflow-hidden overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-[#0f1419] border-b border-[#2a3441]">
              <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Name & IP</th>
              <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Status</th>
              <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">CPU / Memory</th>
              <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Disk</th>
              <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Uptime</th>
              <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">OS</th>
              <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[#2a3441]">
            {MOCK_VMS.map((vm) => (
              <tr key={vm.id} className="hover:bg-[#2a3441]/30 transition-colors group">
                <td className="px-6 py-4">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-[#2a3441] flex items-center justify-center text-[#4caf50]">
                      <Monitor size={18} />
                    </div>
                    <div>
                      <div className="font-bold text-white group-hover:text-[#4caf50] transition-colors">{vm.name}</div>
                      <div className="text-xs text-[#8090a0] font-mono">{vm.ip}</div>
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${getStatusColor(vm.status)}`}>
                    {vm.status}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <div className="w-32 space-y-2">
                    <div className="flex justify-between text-[10px] text-[#8090a0]">
                      <span>CPU: {vm.cpu}%</span>
                      <span>MEM: {vm.memory}%</span>
                    </div>
                    <div className="h-1.5 bg-[#2a3441] rounded-full overflow-hidden">
                       <div className="h-full bg-[#2196f3]" style={{ width: `${vm.cpu}%` }}></div>
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4">
                   <div className="flex items-center gap-2">
                      <div className="text-sm font-semibold">{vm.disk}%</div>
                      <div className="w-16 h-1 bg-[#2a3441] rounded-full overflow-hidden">
                        <div className={`h-full ${vm.disk > 90 ? 'bg-red-500' : 'bg-orange-400'}`} style={{ width: `${vm.disk}%` }}></div>
                      </div>
                   </div>
                </td>
                <td className="px-6 py-4 text-sm text-[#b0b8c5]">{vm.uptime}</td>
                <td className="px-6 py-4 text-xs text-[#8090a0] font-medium">{vm.os}</td>
                <td className="px-6 py-4">
                  <button className="p-2 hover:bg-[#2a3441] rounded-lg transition-colors">
                    <MoreVertical size={18} className="text-[#8090a0]" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default VMList;
