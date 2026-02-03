
import React, { useState } from 'react';
import { useTranslation } from './LanguageContext';
import { 
  Users, 
  Shield, 
  MoreVertical, 
  Search,
  Check,
  Info
} from 'lucide-react';
import { User, Permission } from '../types';

const MOCK_USERS: User[] = [
  { id: '1', name: 'John Doe', email: 'john@vmsentry.io', role: 'Super Admin', status: 'active', lastLogin: '2 mins ago' },
  { id: '2', name: 'Jane Smith', email: 'jane@vmsentry.io', role: 'SRE Lead', status: 'active', lastLogin: '1h ago' },
  { id: '3', name: 'Bob Wilson', email: 'bob@vmsentry.io', role: 'Monitoring Op', status: 'pending', lastLogin: 'Never' },
];

const UserManagement: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'list' | 'roles' | 'matrix'>('list');

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex bg-[#1a1f2e] p-1.5 rounded-2xl border border-[#2a3441]">
          <button 
            onClick={() => setActiveTab('list')}
            className={`px-4 py-2 rounded-xl text-sm font-bold transition-all ${activeTab === 'list' ? 'bg-[#4caf50] text-white' : 'text-[#b0b8c5] hover:text-white'}`}
          >
            Users
          </button>
          <button 
            onClick={() => setActiveTab('roles')}
            className={`px-4 py-2 rounded-xl text-sm font-bold transition-all ${activeTab === 'roles' ? 'bg-[#4caf50] text-white' : 'text-[#b0b8c5] hover:text-white'}`}
          >
            Roles
          </button>
          <button 
            onClick={() => setActiveTab('matrix')}
            className={`px-4 py-2 rounded-xl text-sm font-bold transition-all ${activeTab === 'matrix' ? 'bg-[#4caf50] text-white' : 'text-[#b0b8c5] hover:text-white'}`}
          >
            Permissions Matrix
          </button>
        </div>
        <button className="bg-[#4caf50] px-6 py-2 rounded-xl font-bold hover:shadow-lg transition-all">+ Add User</button>
      </div>

      {activeTab === 'list' && (
        <div className="bg-[#1a1f2e] border border-[#2a3441] rounded-3xl overflow-hidden">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-[#0f1419] border-b border-[#2a3441]">
                <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">User</th>
                <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Role</th>
                <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Status</th>
                <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Last Login</th>
                <th className="px-6 py-4 text-xs font-bold uppercase text-[#8090a0]">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#2a3441]">
              {MOCK_USERS.map((user) => (
                <tr key={user.id} className="hover:bg-[#2a3441]/30 transition-all">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <img src={`https://picsum.photos/seed/${user.id}/32/32`} className="w-8 h-8 rounded-full border border-[#2a3441]" alt="" />
                      <div>
                        <div className="font-bold">{user.name}</div>
                        <div className="text-xs text-[#8090a0]">{user.email}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="flex items-center gap-2 text-sm text-[#b0b8c5]">
                      <Shield size={14} className="text-[#2196f3]" /> {user.role}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className={`text-[10px] font-bold uppercase px-2 py-0.5 rounded-full inline-block ${user.status === 'active' ? 'bg-[#00d4aa]/10 text-[#00d4aa]' : 'bg-[#ff9800]/10 text-[#ff9800]'}`}>
                      {user.status}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm text-[#8090a0]">{user.lastLogin}</td>
                  <td className="px-6 py-4">
                    <button className="p-2 hover:bg-[#2a3441] rounded-lg">
                      <MoreVertical size={18} className="text-[#8090a0]" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {activeTab === 'matrix' && (
        <div className="bg-[#1a1f2e] border border-[#2a3441] rounded-3xl p-6">
           <div className="flex items-center gap-2 mb-6 text-[#2196f3] bg-[#2196f3]/10 p-3 rounded-xl border border-[#2196f3]/20">
              <Info size={18} />
              <span className="text-sm font-medium">This matrix shows the effective permissions for each role. Inheritance is resolved automatically.</span>
           </div>
           
           <div className="overflow-x-auto">
             <table className="w-full text-center border-collapse">
                <thead>
                  <tr>
                    <th className="text-left py-4 text-[#8090a0]">Functional Module</th>
                    <th className="py-4 text-[#8090a0]">Super Admin</th>
                    <th className="py-4 text-[#8090a0]">SRE Lead</th>
                    <th className="py-4 text-[#8090a0]">Monitoring Op</th>
                    <th className="py-4 text-[#8090a0]">Viewer</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#2a3441]">
                  {['Monitoring Data', 'Alert Management', 'User Management', 'System Config'].map(module => (
                    <tr key={module} className="hover:bg-[#2a3441]/30 transition-all">
                      <td className="text-left py-6 font-medium">{module}</td>
                      <td><div className="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-[#4caf50]/20 text-[#4caf50]"><Check size={18} /></div></td>
                      <td><div className="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-[#4caf50]/20 text-[#4caf50]"><Check size={18} /></div></td>
                      <td>{module === 'User Management' ? <span className="text-[#607080]">-</span> : <div className="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-[#4caf50]/20 text-[#4caf50]"><Check size={18} /></div>}</td>
                      <td><div className="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-[#2a3441] text-[#607080] font-mono text-xs">Read</div></td>
                    </tr>
                  ))}
                </tbody>
             </table>
           </div>
        </div>
      )}
    </div>
  );
};

export default UserManagement;
