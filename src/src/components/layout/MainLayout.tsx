import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { 
  LayoutDashboard, 
  Server, 
  History, 
  AlertCircle, 
  Users, 
  Settings,
  Menu,
  Bell,
  User,
  LogOut,
  ChevronDown,
  Globe
} from 'lucide-react';
import { useAuthStore } from '@stores/authStore';

const MainLayout: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [showLangMenu, setShowLangMenu] = useState(false);

  const menuItems = [
    { key: '/dashboard', icon: LayoutDashboard, label: t('dashboard') },
    { key: '/vms', icon: Server, label: t('vmList') },
    { key: '/history', icon: History, label: t('history') },
    { key: '/alerts', icon: AlertCircle, label: t('alerts') },
    { key: '/users', icon: Users, label: t('users') },
    { key: '/system', icon: Settings, label: t('system') },
  ];

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
    localStorage.setItem('language', lang);
    setShowLangMenu(false);
  };

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside 
        className={`bg-surface border-r border-border transition-all duration-300 ${
          sidebarCollapsed ? 'w-20' : 'w-64'
        }`}
      >
        {/* Logo */}
        <div className="h-16 flex items-center justify-center border-b border-border">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-success rounded-xl flex items-center justify-center">
              <Server className="w-6 h-6 text-white" />
            </div>
            {!sidebarCollapsed && (
              <span className="text-lg font-bold text-white">{t('appTitle')}</span>
            )}
          </div>
        </div>

        {/* Navigation */}
        <nav className="p-4 space-y-2">
          {menuItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname.startsWith(item.key);
            return (
              <button
                key={item.key}
                onClick={() => navigate(item.key)}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                  isActive 
                    ? 'bg-success text-white' 
                    : 'text-text-tertiary hover:bg-surface hover:text-white'
                }`}
              >
                <Icon className="w-5 h-5" />
                {!sidebarCollapsed && <span className="font-medium">{item.label}</span>}
              </button>
            );
          })}
        </nav>

        {/* Collapse Button */}
        <div className="absolute bottom-4 left-4">
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="p-2 rounded-lg text-text-muted hover:text-white hover:bg-surface transition-all"
          >
            <Menu className="w-5 h-5" />
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-h-screen overflow-hidden">
        {/* Header */}
        <header className="h-16 bg-surface border-b border-border flex items-center justify-between px-6">
          {/* Left - Breadcrumb could go here */}
          <div />

          {/* Right */}
          <div className="flex items-center gap-4">
            {/* Language Switcher */}
            <div className="relative">
              <button
                onClick={() => setShowLangMenu(!showLangMenu)}
                className="flex items-center gap-2 text-text-tertiary hover:text-white transition-all"
              >
                <Globe className="w-5 h-5" />
                <span className="text-sm font-medium uppercase">{i18n.language}</span>
                <ChevronDown className="w-4 h-4" />
              </button>
              
              {showLangMenu && (
                <div className="absolute right-0 mt-2 w-40 bg-surface border border-border rounded-xl shadow-2xl z-50">
                  <button 
                    onClick={() => changeLanguage('zh-CN')}
                    className="w-full px-4 py-2 text-left hover:bg-border text-sm text-text-secondary"
                  >
                    简体中文
                  </button>
                  <button 
                    onClick={() => changeLanguage('en')}
                    className="w-full px-4 py-2 text-left hover:bg-border text-sm text-text-secondary"
                  >
                    English
                  </button>
                  <button 
                    onClick={() => changeLanguage('ja-JP')}
                    className="w-full px-4 py-2 text-left hover:bg-border text-sm text-text-secondary"
                  >
                    日本語
                  </button>
                </div>
              )}
            </div>

            {/* Notifications */}
            <button className="relative p-2 text-text-tertiary hover:text-white transition-all">
              <Bell className="w-5 h-5" />
              <span className="absolute top-0 right-0 w-4 h-4 bg-danger rounded-full text-xs flex items-center justify-center text-white">
                3
              </span>
            </button>

            {/* User Menu */}
            <div className="relative">
              <button
                onClick={() => setShowUserMenu(!showUserMenu)}
                className="flex items-center gap-3 text-text-tertiary hover:text-white transition-all"
              >
                <div className="w-8 h-8 bg-border rounded-full flex items-center justify-center">
                  <User className="w-4 h-4" />
                </div>
                <span className="text-sm font-medium">{user?.name || user?.username}</span>
                <ChevronDown className="w-4 h-4" />
              </button>

              {showUserMenu && (
                <div className="absolute right-0 mt-2 w-48 bg-surface border border-border rounded-xl shadow-2xl z-50">
                  <div className="px-4 py-3 border-b border-border">
                    <p className="text-sm font-medium text-white">{user?.name}</p>
                    <p className="text-xs text-text-muted">{user?.email}</p>
                  </div>
                  <button 
                    onClick={() => {
                      navigate('/profile');
                      setShowUserMenu(false);
                    }}
                    className="w-full px-4 py-2 text-left hover:bg-border text-sm text-text-secondary flex items-center gap-2"
                  >
                    <User className="w-4 h-4" />
                    {t('profile')}
                  </button>
                  <button 
                    onClick={() => {
                      navigate('/system');
                      setShowUserMenu(false);
                    }}
                    className="w-full px-4 py-2 text-left hover:bg-border text-sm text-text-secondary flex items-center gap-2"
                  >
                    <Settings className="w-4 h-4" />
                    {t('settings')}
                  </button>
                  <div className="border-t border-border">
                    <button 
                      onClick={handleLogout}
                      className="w-full px-4 py-2 text-left hover:bg-border text-sm text-danger flex items-center gap-2"
                    >
                      <LogOut className="w-4 h-4" />
                      {t('logout')}
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </header>

        {/* Content Area */}
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
};

export default MainLayout;