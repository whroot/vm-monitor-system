import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Plus, Search, Filter, MoreVertical, Server, AlertCircle } from 'lucide-react';
import { useVMStore } from '@stores/vmStore';
import { VM } from '@types/api';

const VMList: React.FC = () => {
  const { t } = useTranslation();
  const { vms, total, isLoading, fetchVMs, queryParams, setQueryParams } = useVMStore();
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    fetchVMs();
  }, [fetchVMs]);

  const handleSearch = () => {
    setQueryParams({ keyword: searchTerm, page: 1 });
    fetchVMs();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return 'bg-success/20 text-success';
      case 'offline':
        return 'bg-danger/20 text-danger';
      case 'error':
        return 'bg-warning/20 text-warning';
      default:
        return 'bg-border text-text-muted';
    }
  };

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white">{t('vmList')}</h1>
        <button className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          {t('create')}
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
          <input
            type="text"
            placeholder={t('search')}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
            className="input pl-12"
          />
        </div>
        <button className="btn-secondary flex items-center gap-2">
          <Filter className="w-4 h-4" />
          {t('filter')}
        </button>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left p-4 text-sm font-semibold text-text-muted">名称</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">IP地址</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">操作系统</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">状态</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">资源</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">最后在线</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">操作</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={7} className="p-8 text-center text-text-muted">
                  {t('loading')}
                </td>
              </tr>
            ) : vms.length === 0 ? (
              <tr>
                <td colSpan={7} className="p-8 text-center text-text-muted">
                  {t('noData')}
                </td>
              </tr>
            ) : (
              vms.map((vm: VM) => (
                <tr key={vm.id} className="border-b border-border hover:bg-background/50">
                  <td className="p-4">
                    <div className="flex items-center gap-3">
                      <Server className="w-5 h-5 text-text-muted" />
                      <div>
                        <div className="font-medium text-white">{vm.name}</div>
                        <div className="text-xs text-text-muted">{vm.vmwareId || vm.id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="p-4 text-text-secondary">{vm.ip || '-'}</td>
                  <td className="p-4 text-text-secondary">
                    {vm.os} {vm.osVersion}
                  </td>
                  <td className="p-4">
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(vm.status)}`}>
                      {t(vm.status)}
                    </span>
                  </td>
                  <td className="p-4 text-text-secondary">
                    {vm.cpuCores || '-'} CPU / {vm.memoryGB || '-'} GB
                  </td>
                  <td className="p-4 text-text-secondary">
                    {vm.lastSeen ? new Date(vm.lastSeen).toLocaleString() : '-'}
                  </td>
                  <td className="p-4">
                    <button className="p-2 hover:bg-border rounded-lg text-text-muted hover:text-white">
                      <MoreVertical className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>

        {/* Pagination */}
        <div className="flex justify-between items-center p-4 border-t border-border">
          <div className="text-sm text-text-muted">
            共 {total} 条记录
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => {
                if (queryParams.page && queryParams.page > 1) {
                  setQueryParams({ page: queryParams.page - 1 });
                  fetchVMs();
                }
              }}
              disabled={!queryParams.page || queryParams.page <= 1}
              className="px-4 py-2 bg-surface border border-border rounded-lg text-sm text-text-secondary hover:border-text-tertiary disabled:opacity-50"
            >
              {t('previous')}
            </button>
            <button
              onClick={() => {
                setQueryParams({ page: (queryParams.page || 1) + 1 });
                fetchVMs();
              }}
              className="px-4 py-2 bg-surface border border-border rounded-lg text-sm text-text-secondary hover:border-text-tertiary"
            >
              {t('next')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VMList;