
import React, { useState } from 'react';
import { useTranslation } from './LanguageContext';
import { Shield, Globe, Eye, EyeOff, Lock, User } from 'lucide-react';
import { Language } from '../types';

interface LoginProps {
  onLogin: () => void;
}

const Login: React.FC<LoginProps> = ({ onLogin }) => {
  const { t, language, setLanguage } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setTimeout(() => {
      onLogin();
      setIsLoading(false);
    }, 1200);
  };

  return (
    <div className="min-h-screen bg-[#0f1419] flex items-center justify-center p-4 relative overflow-hidden">
      {/* Background Decor */}
      <div className="absolute top-0 right-0 w-96 h-96 bg-[#4caf50]/5 rounded-full blur-3xl -mr-48 -mt-48"></div>
      <div className="absolute bottom-0 left-0 w-96 h-96 bg-[#2196f3]/5 rounded-full blur-3xl -ml-48 -mb-48"></div>

      <div className="w-full max-w-md relative">
        <div className="flex justify-between items-center mb-8">
           <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-[#4caf50] rounded-xl flex items-center justify-center shadow-lg shadow-[#4caf50]/20">
                 <Shield size={24} className="text-white" />
              </div>
              <h1 className="text-2xl font-bold tracking-tight text-white">{t('appTitle')}</h1>
           </div>
           
           <div className="relative group">
              <button className="flex items-center gap-2 text-[#b0b8c5] hover:text-white text-xs font-bold uppercase transition-all">
                 <Globe size={16} /> {language}
              </button>
              <div className="absolute right-0 mt-2 w-40 bg-[#1a1f2e] border border-[#2a3441] rounded-xl shadow-2xl overflow-hidden hidden group-hover:block z-50">
                 <button onClick={() => setLanguage('en')} className="w-full px-4 py-2 text-left hover:bg-[#2a3441] text-sm">English</button>
                 <button onClick={() => setLanguage('zh')} className="w-full px-4 py-2 text-left hover:bg-[#2a3441] text-sm">简体中文</button>
                 <button onClick={() => setLanguage('jp')} className="w-full px-4 py-2 text-left hover:bg-[#2a3441] text-sm">日本語</button>
              </div>
           </div>
        </div>

        <div className="bg-[#1a1f2e] border border-[#2a3441] p-8 rounded-3xl shadow-2xl">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label className="text-sm font-semibold text-[#8090a0] ml-1">{t('username')}</label>
              <div className="relative">
                <User size={18} className="absolute left-4 top-1/2 -translate-y-1/2 text-[#607080]" />
                <input 
                  type="text" 
                  required
                  placeholder="admin@vmsentry.io" 
                  className="w-full bg-[#0f1419] border border-[#2a3441] rounded-xl py-3.5 pl-12 pr-4 focus:border-[#4caf50] focus:ring-2 focus:ring-[#4caf50]/20 outline-none transition-all text-white"
                />
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <label className="text-sm font-semibold text-[#8090a0] ml-1">{t('password')}</label>
                <button type="button" className="text-xs text-[#2196f3] hover:underline">{t('forgotPassword')}</button>
              </div>
              <div className="relative">
                <Lock size={18} className="absolute left-4 top-1/2 -translate-y-1/2 text-[#607080]" />
                <input 
                  type={showPassword ? 'text' : 'password'} 
                  required
                  placeholder="••••••••" 
                  className="w-full bg-[#0f1419] border border-[#2a3441] rounded-xl py-3.5 pl-12 pr-12 focus:border-[#4caf50] focus:ring-2 focus:ring-[#4caf50]/20 outline-none transition-all text-white"
                />
                <button 
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-[#607080] hover:text-[#b0b8c5]"
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </div>

            <div className="flex items-center gap-3 ml-1">
              <input type="checkbox" id="remember" className="w-4 h-4 rounded bg-[#0f1419] border-[#2a3441] text-[#4caf50] focus:ring-[#4caf50]" />
              <label htmlFor="remember" className="text-sm text-[#b0b8c5]">{t('rememberMe')}</label>
            </div>

            <button 
              type="submit"
              disabled={isLoading}
              className={`w-full py-4 rounded-xl font-bold text-lg shadow-xl shadow-[#4caf50]/20 transition-all active:scale-95 flex items-center justify-center
                ${isLoading ? 'bg-[#2a3441] text-[#607080]' : 'bg-[#4caf50] text-white hover:bg-[#43a047]'}
              `}
            >
              {isLoading ? (
                <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin"></div>
              ) : t('login')}
            </button>
          </form>
        </div>

        <div className="mt-8 text-center text-[#607080] text-xs">
          © 2026 VMSentry. Enterprise Edition v1.0.4. All rights reserved.
        </div>
      </div>
    </div>
  );
};

export default Login;
