
import React, { useState } from 'react';
import { LanguageProvider } from './components/LanguageContext';
import Layout from './components/Layout';
import Dashboard from './components/Dashboard';
import VMList from './components/VMList';
import HistoryAnalysis from './components/HistoryAnalysis';
import AlertManagement from './components/AlertManagement';
import UserManagement from './components/UserManagement';
import ImageEditor from './components/ImageEditor';
import Login from './components/Login';

const App: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [activeTab, setActiveTab] = useState('dashboard');

  if (!isAuthenticated) {
    return (
      <LanguageProvider>
        <Login onLogin={() => setIsAuthenticated(true)} />
      </LanguageProvider>
    );
  }

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard': return <Dashboard />;
      case 'vmlist': return <VMList />;
      case 'history': return <HistoryAnalysis />;
      case 'alerts': return <AlertManagement />;
      case 'users': return <UserManagement />;
      case 'image-editor': return <ImageEditor />; // Optional feature tab
      default: return <Dashboard />;
    }
  };

  return (
    <LanguageProvider>
      <Layout 
        activeTab={activeTab} 
        setActiveTab={setActiveTab} 
        onLogout={() => setIsAuthenticated(false)}
      >
        {renderContent()}
        
        {/* Floating AI Tool Button */}
        <button 
          onClick={() => setActiveTab('image-editor')}
          className={`fixed bottom-6 right-6 w-14 h-14 rounded-full shadow-2xl flex items-center justify-center transition-all z-50
            ${activeTab === 'image-editor' ? 'bg-[#4caf50] rotate-45' : 'bg-purple-600 hover:scale-110'}
          `}
          title="AI Visual Assistant"
        >
          <span className="text-white text-2xl">✨</span>
        </button>
      </Layout>
    </LanguageProvider>
  );
};

export default App;
