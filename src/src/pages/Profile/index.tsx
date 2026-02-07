import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { User, Mail, Phone, Building, Shield, Save, Check, Camera, RefreshCw } from 'lucide-react';
import { useAuthStore } from '@stores/authStore';
import { authApi } from '../../api/auth';

const Profile: React.FC = () => {
  const { t } = useTranslation();
  const { user, fetchUser, logout } = useAuthStore();
  const [saved, setSaved] = useState(false);
  const [saving, setSaving] = useState(false);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const [profile, setProfile] = useState({
    name: '',
    email: '',
    phone: '',
    department: '',
    avatar: '',
  });

  const loadProfile = async () => {
    try {
      const userData = await authApi.getMe();
      setProfile({
        name: userData.name || '',
        email: userData.email || '',
        phone: userData.phone || '',
        department: userData.department || '',
        avatar: (userData as any).avatar || '',
      });
    } catch (error) {
      console.error('加载用户数据失败:', error);
    }
  };

  useEffect(() => {
    loadProfile();
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await authApi.updateProfile({
        name: profile.name,
        email: profile.email,
        phone: profile.phone,
        department: profile.department,
      });
      await loadProfile();
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (error) {
      console.error('保存失败:', error);
      alert('保存失败，请重试');
    } finally {
      setSaving(false);
    }
  };

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploadingAvatar(true);
    try {
      const formData = new FormData();
      formData.append('avatar', file);
      await authApi.uploadAvatar(formData);
      await loadProfile();
    } catch (error) {
      console.error('上传头像失败:', error);
      alert('上传头像失败，请重试');
    } finally {
      setUploadingAvatar(false);
    }
  };

  const handleRefresh = async () => {
    await fetchUser();
    await loadProfile();
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
          <User className="w-6 h-6" />
          个人资料
        </h1>
        <button
          onClick={handleRefresh}
          className="btn-secondary flex items-center gap-2"
        >
          <RefreshCw className="w-4 h-4" />
          刷新数据
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="card lg:col-span-1">
          <div className="flex flex-col items-center py-8">
            <div className="relative mb-4">
              {profile.avatar ? (
                <img
                  src={profile.avatar}
                  alt="头像"
                  className="w-24 h-24 rounded-full object-cover border-2 border-success"
                />
              ) : (
                <div className="w-24 h-24 bg-success/20 rounded-full flex items-center justify-center">
                  <User className="w-12 h-12 text-success" />
                </div>
              )}
              <label
                className={`absolute bottom-0 right-0 w-8 h-8 bg-primary rounded-full flex items-center justify-center cursor-pointer hover:bg-primary/80 transition-colors ${
                  uploadingAvatar ? 'opacity-50 cursor-not-allowed' : ''
                }`}
              >
                {uploadingAvatar ? (
                  <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                ) : (
                  <Camera className="w-4 h-4 text-white" />
                )}
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleAvatarChange}
                  className="hidden"
                  disabled={uploadingAvatar}
                />
              </label>
            </div>
            <h2 className="text-xl font-semibold text-white">{user?.name || user?.username}</h2>
            <p className="text-text-muted mt-1">{user?.username}</p>
            <div className="flex gap-2 mt-4">
              {user?.roles?.map((role: any) => (
                <span key={role.id} className="px-3 py-1 bg-primary/20 text-primary rounded-full text-sm">
                  {role.name}
                </span>
              ))}
            </div>
          </div>
        </div>

        <div className="card lg:col-span-2">
          <h3 className="text-lg font-semibold text-white mb-6">基本信息</h3>
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                姓名
              </label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
                <input
                  type="text"
                  value={profile.name}
                  onChange={(e) => setProfile({ ...profile, name: e.target.value })}
                  className="input w-full pl-10"
                  placeholder="输入姓名"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                邮箱
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
                <input
                  type="email"
                  value={profile.email}
                  onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                  className="input w-full pl-10"
                  placeholder="输入邮箱"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                电话
              </label>
              <div className="relative">
                <Phone className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
                <input
                  type="tel"
                  value={profile.phone}
                  onChange={(e) => setProfile({ ...profile, phone: e.target.value })}
                  className="input w-full pl-10"
                  placeholder="输入电话"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                部门
              </label>
              <div className="relative">
                <Building className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
                <input
                  type="text"
                  value={profile.department}
                  onChange={(e) => setProfile({ ...profile, department: e.target.value })}
                  className="input w-full pl-10"
                  placeholder="输入部门"
                />
              </div>
            </div>

            <div className="flex justify-end">
              <button
                onClick={handleSave}
                disabled={saving}
                className="btn-primary flex items-center gap-2 disabled:opacity-50"
              >
                {saving ? (
                  <>
                    <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    保存中...
                  </>
                ) : saved ? (
                  <>
                    <Check className="w-4 h-4" />
                    已保存
                  </>
                ) : (
                  <>
                    <Save className="w-4 h-4" />
                    保存更改
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="card">
        <h3 className="text-lg font-semibold text-white mb-6">账户信息</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 bg-surface rounded-lg flex items-center justify-center">
              <Shield className="w-5 h-5 text-text-muted" />
            </div>
            <div>
              <p className="text-sm text-text-muted">用户名</p>
              <p className="text-white font-medium">{user?.username}</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 bg-surface rounded-lg flex items-center justify-center">
              <User className="w-5 h-5 text-text-muted" />
            </div>
            <div>
              <p className="text-sm text-text-muted">用户ID</p>
              <p className="text-white font-medium text-sm">{user?.id}</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 bg-surface rounded-lg flex items-center justify-center">
              <Check className="w-5 h-5 text-success" />
            </div>
            <div>
              <p className="text-sm text-text-muted">账户状态</p>
              <p className="text-success font-medium">活跃</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Profile;
