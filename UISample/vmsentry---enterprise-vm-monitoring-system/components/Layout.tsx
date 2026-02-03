
import React, { useState } from 'react';
import { useTranslation } from './LanguageContext';
import { 
  LayoutDashboard, 
  Server, 
  History, 
  AlertTriangle, 
  Users, 
  Menu, 
  X, 
  Globe,
  LogOut,
  Bell
} from 'lucide-react';
import { Language } from '../types';

interface LayoutProps {
  children: React.ReactNode;
  activeTab: string;
  setActiveTab: (tab: string) => void;
  onLogout: () => void;
}

const Layout: React.FC<LayoutProps> = ({ children, activeTab, setActiveTab, onLogout }) => {
  const { t, setLanguage, language } = useTranslation();
  const [isSidebarOpen, setSidebarOpen] = useState(false);
  const [showLangMenu, setShowLangMenu] = useState(false);

  const navItems = [
    { id: 'dashboard', label: t('dashboard'), icon: LayoutDashboard },
    { id: 'vmlist', label: t('vmList'), icon: Server },
    { id: 'history', label: t('history'), icon: History },
    { id: 'alerts', label: t('alerts'), icon: AlertTriangle },
    { id: 'users', label: t('users'), icon: Users },
  ];

  return (
    <div className="min-h-screen bg-[#0f1419] text-white flex flex-col md:flex-row">
      {/* Mobile Header */}
      <div className="md:hidden bg-[#1a1f2e] border-b border-[#2a3441] p-4 flex justify-between items-center sticky top-0 z-50">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-[#4caf50] rounded-lg flex items-center justify-center font-bold">V</div>
          <span className="font-bold tracking-tight">{t('appTitle')}</span>
        </div>
        <button onClick={() => setSidebarOpen(!isSidebarOpen)}>
          {isSidebarOpen ? <X /> : <Menu />}
        </button>
      </div>

      {/* Sidebar */}
      <aside className={`
        fixed inset-y-0 left-0 z-40 w-64 bg-[#1a1f2e] border-r border-[#2a3441] transform transition-transform duration-300 md:relative md:translate-x-0
        ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        <div className="p-6 hidden md:flex items-center gap-3">
          <div className="w-10 h-10 bg-[#4caf50] rounded-lg flex items-center justify-center font-bold text-xl">V</div>
          <span className="font-bold text-lg">{t('appTitle')}</span>
        </div>

        <nav className="mt-4 px-4 space-y-2">
          {navItems.map((item) => (
            <button
              key={item.id}
              onClick={() => {
                setActiveTab(item.id);
                setSidebarOpen(false);
              }}
              className={`
                w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all
                ${activeTab === item.id ? 'bg-[#4caf50] text-white shadow-lg shadow-[#4caf50]/20' : 'text-[#b0b8c5] hover:bg-[#2a3441] hover:text-white'}
              `}
            >
              <item.icon size={20} />
              <span className="font-medium">{item.label}</span>
            </button>
          ))}
        </nav>

        <div className="absolute bottom-4 left-0 w-full px-4 space-y-2">
          <button 
            onClick={onLogout}
            className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-[#f44336] hover:bg-[#f44336]/10 transition-all"
          >
            <LogOut size={20} />
            <span className="font-medium">Logout</span>
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-h-0">
        {/* Top Header */}
        <header className="hidden md:flex bg-[#0f1419] border-b border-[#2a3441] h-16 items-center justify-between px-8 sticky top-0 z-30">
          <div className="text-[#b0b8c5] font-medium">
            {navItems.find(i => i.id === activeTab)?.label}
          </div>
          
          <div className="flex items-center gap-6">
            <div className="relative">
              <button className="text-[#b0b8c5] hover:text-white transition-colors relative">
                <Bell size={20} />
                <span className="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full"></span>
              </button>
            </div>

            <div className="relative">
              <button 
                onClick={() => setShowLangMenu(!showLangMenu)}
                className="flex items-center gap-2 text-[#b0b8c5] hover:text-white transition-colors"
              >
                <Globe size={18} />
                <span className="uppercase text-sm font-semibold">{language}</span>
              </button>
              
              {showLangMenu && (
                <div className="absolute right-0 mt-2 w-40 bg-[#1a1f2e] border border-[#2a3441] rounded-xl shadow-2xl py-2 z-50">
                  <button onClick={() => { setLanguage('en'); setShowLangMenu(false); }} className="w-full px-4 py-2 text-left hover:bg-[#2a3441] flex items-center gap-2">🇺🇸 English</button>
                  <button onClick={() => { setLanguage('zh'); setShowLangMenu(false); }} className="w-full px-4 py-2 text-left hover:bg-[#2a3441] flex items-center gap-2">🇨🇳 简体中文</button>
                  <button onClick={() => { setLanguage('jp'); setShowLangMenu(false); }} className="w-full px-4 py-2 text-left hover:bg-[#2a3441] flex items-center gap-2">🇯🇵 日本語</button>
                </div>
              )}
            </div>

            <div className="flex items-center gap-3 border-l border-[#2a3441] pl-6">
              <img src="https://picsum.photos/seed/admin/40/40" className="w-9 h-9 rounded-full border border-[#2a3441]" alt="User" />
              <div className="hidden lg:block text-sm">
                <div className="font-semibold">Admin User</div>
                <div className="text-[#8090a0] text-xs">Super Admin</div>
              </div>
            </div>
          </div>
        </header>

        {/* Dynamic Content */}
        <div className="flex-1 overflow-y-auto p-4 md:p-8">
          {children}
        </div>
      </main>
    </div>
  );
};

export default Layout;
