import { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from '@stores/authStore';

// Layouts
import MainLayout from '@components/layout/MainLayout';

// Pages
import Login from '@pages/Login';
import Dashboard from '@pages/Dashboard';
import VMList from '@pages/VMList';
import VMDetail from '@pages/VMDetail';
import HistoryData from '@pages/HistoryData';
import AlertManagement from '@pages/AlertManagement';
import UserManagement from '@pages/UserManagement';
import SystemSettings from '@pages/SystemSettings';
import Profile from '@pages/Profile';

// 路由守卫组件
const PrivateRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuthStore();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
};

function App() {
  const { isAuthenticated, fetchUser } = useAuthStore();

  // 初始化时检查登录状态
  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (token && !isAuthenticated) {
      fetchUser();
    }
  }, [fetchUser, isAuthenticated]);

  return (
    <Routes>
      {/* 登录页面 */}
      <Route path="/login" element={<Login />} />

      {/* 受保护的路由 */}
      <Route
        path="/"
        element={
          <PrivateRoute>
            <MainLayout />
          </PrivateRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="vms" element={<VMList />} />
        <Route path="vms/:id" element={<VMDetail />} />
        <Route path="history" element={<HistoryData />} />
        <Route path="alerts" element={<AlertManagement />} />
        <Route path="users" element={<UserManagement />} />
        <Route path="system" element={<SystemSettings />} />
        <Route path="profile" element={<Profile />} />
      </Route>

      {/* 404 */}
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

export default App;
