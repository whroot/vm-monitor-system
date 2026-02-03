import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Shield, Globe, Eye, EyeOff, Lock, User, Loader2 } from 'lucide-react';
import { useAuthStore } from '@stores/authStore';

const Login: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { login, isLoading, error } = useAuthStore();
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showLangMenu, setShowLangMenu] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login({
        username,
        password,
        rememberMe,
        language: i18n.language
      });
      navigate('/dashboard');
    } catch (err) {
      // Error is handled in store
    }
  };

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
    localStorage.setItem('language', lang);
    setShowLangMenu(false);
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4 relative overflow-hidden">
      {/* Background Decor */}
      <div className="absolute top-0 right-0 w-96 h-96 bg-success/5 rounded-full blur-3xl -mr-48 -mt-48" />
      <div className="absolute bottom-0 left-0 w-96 h-96 bg-info/5 rounded-full blur-3xl -ml-48 -mb-48" />

      <div className="w-full max-w-md relative z-10">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-success rounded-xl flex items-center justify-center shadow-lg shadow-success/20">
              <Shield className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white">{t('appTitle')}</h1>
              <p className="text-xs text-text-muted">{t('appSubtitle')}</p>
            </div>
          </div>

          {/* Language Switcher */}
          <div className="relative">
            <button
              onClick={() => setShowLangMenu(!showLangMenu)}
              className="flex items-center gap-2 text-text-tertiary hover:text-white text-xs font-bold uppercase transition-all"
            >
              <Globe className="w-4 h-4" /> {i18n.language}
            </button>
            {showLangMenu && (
              <div className="absolute right-0 mt-2 w-40 bg-surface border border-border rounded-xl shadow-2xl overflow-hidden z-50">
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
        </div>

        {/* Login Card */}
        <div className="bg-surface border border-border p-8 rounded-3xl shadow-2xl">
          {error && (
            <div className="mb-6 p-4 bg-danger/10 border border-danger/20 rounded-xl">
              <p className="text-sm text-danger">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Username */}
            <div className="space-y-2">
              <label className="text-sm font-semibold text-text-muted ml-1">
                {t('username')}
              </label>
              <div className="relative">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                  placeholder="admin"
                  className="input pl-12"
                />
              </div>
            </div>

            {/* Password */}
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <label className="text-sm font-semibold text-text-muted ml-1">
                  {t('password')}
                </label>
                <button type="button" className="text-xs text-info hover:underline">
                  {t('forgotPassword')}
                </button>
              </div>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  placeholder="••••••••"
                  className="input pl-12 pr-12"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted hover:text-white"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            {/* Remember Me */}
            <div className="flex items-center gap-3 ml-1">
              <input
                type="checkbox"
                id="remember"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                className="w-4 h-4 rounded bg-background border-border text-success focus:ring-success"
              />
              <label htmlFor="remember" className="text-sm text-text-tertiary">
                {t('rememberMe')}
              </label>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-4 rounded-xl font-bold text-lg shadow-xl shadow-success/20 transition-all active:scale-95 flex items-center justify-center gap-2 bg-success text-white hover:bg-success/90 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  {t('signInLoading')}
                </>
              ) : (
                t('signIn')
              )}
            </button>
          </form>
        </div>

        {/* Footer */}
        <p className="text-center text-text-muted text-sm mt-8">
          © 2026 {t('appTitle')}. {t('allRightsReserved')}.
        </p>
      </div>
    </div>
  );
};

export default Login;