import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Server, Cpu, HardDrive, Activity } from 'lucide-react';

const VMDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/vms')}
          className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-white"
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-3">
            <Server className="w-6 h-6" />
            VM详情
          </h1>
          <p className="text-text-muted">ID: {id}</p>
        </div>
      </div>

      {/* Content Placeholder */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="card lg:col-span-2">
          <h3 className="text-lg font-semibold text-white mb-4">实时监控</h3>
          <div className="h-64 flex items-center justify-center text-text-muted">
            性能图表区域
          </div>
        </div>

        <div className="space-y-6">
          <div className="card">
            <h3 className="text-lg font-semibold text-white mb-4">基本信息</h3>
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <Cpu className="w-5 h-5 text-info" />
                <div>
                  <div className="text-sm text-text-muted">CPU</div>
                  <div className="text-white">4 核心</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Activity className="w-5 h-5 text-success" />
                <div>
                  <div className="text-sm text-text-muted">内存</div>
                  <div className="text-white">8 GB</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <HardDrive className="w-5 h-5 text-warning" />
                <div>
                  <div className="text-sm text-text-muted">磁盘</div>
                  <div className="text-white">100 GB</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VMDetail;