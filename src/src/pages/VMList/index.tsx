import React, { useEffect, useState, useRef } from 'react';
import { Plus, Search, Filter, MoreVertical, Server, X, Check, Network, Download, FileSpreadsheet } from 'lucide-react';
import { useVMStore } from '../../stores/vmStore';
import { VM, VMMetrics } from '../../types/api';

type ExportFormat = 'csv' | 'json';

const VMList: React.FC = () => {
  const { vms, total, isLoading, fetchVMs, queryParams, setQueryParams, createVM, metrics, fetchAllMetrics } = useVMStore();
  const [searchTerm, setSearchTerm] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showFilterModal, setShowFilterModal] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [exportFormat, setExportFormat] = useState<ExportFormat>('csv');
  const [exporting, setExporting] = useState(false);
  const searchTimerRef = useRef<NodeJS.Timeout | null>(null);
  const lastSearchTermRef = useRef('');
  const [newVM, setNewVM] = useState<{
    name: string;
    ip: string;
    os: 'Linux' | 'Windows';
  }>({
    name: '',
    ip: '',
    os: 'Linux',
  });
  const [creating, setCreating] = useState(false);
  const [filters, setFilters] = useState({
    status: '',
    os: '',
  });

  // 初始加载
  useEffect(() => {
    fetchAllMetrics();
    fetchVMs();
  }, []);

  // 防抖搜索：输入停止300ms后自动搜索
  useEffect(() => {
    if (searchTerm === lastSearchTermRef.current) return;
    
    lastSearchTermRef.current = searchTerm;
    
    if (searchTimerRef.current) {
      clearTimeout(searchTimerRef.current);
    }
    
    searchTimerRef.current = setTimeout(() => {
      setQueryParams({ keyword: searchTerm, page: 1 });
      fetchVMs();
    }, 300);
    
    return () => {
      if (searchTimerRef.current) {
        clearTimeout(searchTimerRef.current);
      }
    };
  }, [searchTerm, fetchVMs, setQueryParams]);

  const handleApplyFilters = () => {
    setQueryParams({ 
      status: (filters.status as 'online' | 'offline' | 'error' | undefined) || undefined, 
      os: (filters.os as 'Linux' | 'Windows' | undefined) || undefined,
      page: 1 
    });
    fetchVMs();
    setShowFilterModal(false);
  };

  const handleClearFilters = () => {
    setFilters({ status: '', os: '' });
    setQueryParams({ status: undefined, os: undefined, page: 1 });
    fetchVMs();
    setShowFilterModal(false);
  };

  const handleCreateVM = async () => {
    if (!newVM.name) {
      alert('请输入VM名称');
      return;
    }
    setCreating(true);
    try {
      await createVM({
        name: newVM.name,
        ip: newVM.ip,
        os: newVM.os,
      });
      setShowCreateModal(false);
      setNewVM({ name: '', ip: '', os: 'Linux' });
    } catch (error) {
      console.error('创建VM失败:', error);
      alert('创建VM失败');
    } finally {
      setCreating(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
      case 'online':
        return 'bg-success/20 text-success';
      case 'poweredOff':
      case 'offline':
        return 'bg-danger/20 text-danger';
      case 'warning':
        return 'bg-warning/20 text-warning';
      default:
        return 'bg-border text-text-muted';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'running': return '运行中';
      case 'poweredOff': return '已停止';
      case 'warning': return '警告';
      case 'unknown': return '未知';
      default: return status;
    }
  };

  const getVMMetrics = (vmId: string): VMMetrics | undefined => {
    return metrics.find(m => m.vmId === vmId);
  };

  const getUsageColor = (usage: number) => {
    if (usage >= 80) return 'text-danger';
    if (usage >= 60) return 'text-warning';
    return 'text-success';
  };

  const handleExport = async () => {
    if (vms.length === 0) return;
    setExporting(true);
    try {
      const timestamp = new Date().toISOString().slice(0, 10);
      
      if (exportFormat === 'csv') {
        exportCSV(timestamp);
      } else {
        exportJSON(timestamp);
      }
    } finally {
      setExporting(false);
      setShowExportModal(false);
    }
  };

  const exportCSV = (timestamp: string) => {
    const headers = ['名称', 'IP地址', '操作系统', '状态', 'CPU(%)', '内存(%)', '磁盘(%)', '网络入站(MB/s)', '网络出站(MB/s)', '创建时间', '宿主机', '数据中心', '集群'];
    const rows: string[][] = [headers];

    vms.forEach(vm => {
      const vmMetrics = getVMMetrics(vm.id);
      const row = [
        vm.name,
        vm.ip || '-',
        vm.os || '-',
        getStatusText(vm.status),
        vmMetrics?.cpuUsage.toFixed(1) || '-',
        vmMetrics?.memoryUsage.toFixed(1) || '-',
        vmMetrics?.diskUsage.toFixed(1) || '-',
        vmMetrics?.networkInMbps.toFixed(1) || '-',
        vmMetrics?.networkOutMbps.toFixed(1) || '-',
        vm.createdAt ? new Date(vm.createdAt).toLocaleString() : '-',
        vm.hostName || '-',
        vm.datacenterName || '-',
        vm.clusterName || '-',
      ];
      rows.push(row);
    });

    const csvContent = rows.map(row => row.map(cell => `"${cell}"`).join(',')).join('\n');
    downloadFile(csvContent, `vm-list-${timestamp}.csv`, 'text/csv');
  };

  const exportJSON = (timestamp: string) => {
    const exportData = {
      exportTime: new Date().toISOString(),
      totalCount: vms.length,
      vms: vms.map(vm => {
        const vmMetrics = getVMMetrics(vm.id);
        return {
          id: vm.id,
          name: vm.name,
          ip: vm.ip || null,
          os: vm.os,
          status: vm.status,
          cpuCores: vm.cpuCores,
          memoryGB: vm.memoryGB,
          diskGB: vm.diskGB,
          metrics: {
            cpuUsage: vmMetrics?.cpuUsage || null,
            memoryUsage: vmMetrics?.memoryUsage || null,
            diskUsage: vmMetrics?.diskUsage || null,
            networkInMbps: vmMetrics?.networkInMbps || null,
            networkOutMbps: vmMetrics?.networkOutMbps || null,
          },
          host: {
            id: vm.hostId,
            name: vm.hostName,
          },
          location: {
            datacenter: vm.datacenterName,
            cluster: vm.clusterName,
          },
          createdAt: vm.createdAt,
        };
      }),
    };

    const jsonContent = JSON.stringify(exportData, null, 2);
    downloadFile(jsonContent, `vm-list-${timestamp}.json`, 'application/json');
  };

  const downloadFile = (content: string, filename: string, mimeType: string) => {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white">虚拟机管理</h1>
        <div className="flex gap-2">
          <button
            onClick={() => setShowExportModal(true)}
            disabled={vms.length === 0}
            className="btn-secondary flex items-center gap-2 disabled:opacity-50"
          >
            <Download className="w-4 h-4" />
            导出
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="btn-primary flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            新建VM
          </button>
        </div>
      </div>

      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
          <input
            type="text"
            placeholder="搜索虚拟机...（实时搜索）"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="input pl-12"
          />
          {isLoading && (
            <div className="absolute right-4 top-1/2 -translate-y-1/2">
              <div className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
            </div>
          )}
        </div>
        <button 
          onClick={() => setShowFilterModal(true)}
          className="btn-secondary flex items-center gap-2"
        >
          <Filter className="w-4 h-4" />
          筛选
          {(filters.status || filters.os) && (
            <span className="w-2 h-2 bg-primary rounded-full" />
          )}
        </button>
      </div>

      <div className="card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left p-4 text-sm font-semibold text-text-muted">名称</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">IP地址</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">操作系统</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">状态</th>
              <th className="text-center p-4 text-sm font-semibold text-text-muted">CPU</th>
              <th className="text-center p-4 text-sm font-semibold text-text-muted">内存</th>
              <th className="text-center p-4 text-sm font-semibold text-text-muted">磁盘</th>
              <th className="text-center p-4 text-sm font-semibold text-text-muted">网络</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">创建时间</th>
              <th className="text-left p-4 text-sm font-semibold text-text-muted">操作</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={11} className="p-8 text-center text-text-muted">
                  加载中...
                </td>
              </tr>
            ) : vms.length === 0 ? (
              <tr>
                <td colSpan={11} className="p-8 text-center text-text-muted">
                  暂无数据
                </td>
              </tr>
            ) : (
              vms.map((vm: VM) => {
                const vmMetrics = getVMMetrics(vm.id);
                return (
                <tr key={vm.id} className="border-b border-border hover:bg-background/50">
                  <td className="p-4">
                    <div className="flex items-center gap-3">
                      <Server className="w-5 h-5 text-text-muted" />
                      <div>
                        <div className="font-medium text-white">{vm.name}</div>
                        <div className="text-xs text-text-muted">{vm.id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="p-4 text-text-secondary">{vm.ip || '-'}</td>
                  <td className="p-4 text-text-secondary">{vm.os || '-'}</td>
                  <td className="p-4">
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(vm.status)}`}>
                      {getStatusText(vm.status)}
                    </span>
                  </td>
                  <td className="p-4 text-center">
                    {vmMetrics ? (
                      <div className={`text-sm font-medium ${getUsageColor(vmMetrics.cpuUsage)}`}>
                        {vmMetrics.cpuUsage.toFixed(1)}%
                      </div>
                    ) : (
                      <span className="text-text-muted text-sm">-</span>
                    )}
                  </td>
                  <td className="p-4 text-center">
                    {vmMetrics ? (
                      <div className={`text-sm font-medium ${getUsageColor(vmMetrics.memoryUsage)}`}>
                        {vmMetrics.memoryUsage.toFixed(1)}%
                      </div>
                    ) : (
                      <span className="text-text-muted text-sm">-</span>
                    )}
                  </td>
                  <td className="p-4 text-center">
                    {vmMetrics ? (
                      <div className={`text-sm font-medium ${getUsageColor(vmMetrics.diskUsage)}`}>
                        {vmMetrics.diskUsage.toFixed(1)}%
                      </div>
                    ) : (
                      <span className="text-text-muted text-sm">-</span>
                    )}
                  </td>
                  <td className="p-4 text-center">
                    {vmMetrics ? (
                      <div className="text-xs text-text-secondary">
                        <div className="flex items-center justify-center gap-1">
                          <Network className="w-3 h-3 text-success" />
                          <span>{vmMetrics.networkInMbps.toFixed(1)}</span>
                        </div>
                        <div className="flex items-center justify-center gap-1">
                          <Network className="w-3 h-3 text-primary" />
                          <span>{vmMetrics.networkOutMbps.toFixed(1)}</span>
                        </div>
                      </div>
                    ) : (
                      <span className="text-text-muted text-sm">-</span>
                    )}
                  </td>
                  <td className="p-4 text-text-secondary">
                    {vm.createdAt ? new Date(vm.createdAt).toLocaleString() : '-'}
                  </td>
                  <td className="p-4">
                    <button className="p-2 hover:bg-border rounded-lg text-text-muted hover:text-white">
                      <MoreVertical className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              )})
            )}
          </tbody>
        </table>

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
              上一页
            </button>
            <button
              onClick={() => {
                setQueryParams({ page: (queryParams.page || 1) + 1 });
                fetchVMs();
              }}
              className="px-4 py-2 bg-surface border border-border rounded-lg text-sm text-text-secondary hover:border-text-tertiary"
            >
              下一页
            </button>
          </div>
        </div>
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white">新建虚拟机</h3>
              <button
                onClick={() => setShowCreateModal(false)}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-text-muted mb-1">VM名称 *</label>
                <input
                  type="text"
                  value={newVM.name}
                  onChange={(e) => setNewVM({ ...newVM, name: e.target.value })}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="输入VM名称"
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">IP地址</label>
                <input
                  type="text"
                  value={newVM.ip}
                  onChange={(e) => setNewVM({ ...newVM, ip: e.target.value })}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                  placeholder="例如: 192.168.1.100"
                />
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">操作系统</label>
                <select
                  value={newVM.os}
                  onChange={(e) => setNewVM({ ...newVM, os: e.target.value as 'Linux' | 'Windows' })}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                >
                  <option value="Linux">Linux</option>
                  <option value="Windows">Windows</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 text-text-muted hover:text-white transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleCreateVM}
                disabled={creating}
                className="btn-primary flex items-center gap-2"
              >
                {creating ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    创建中...
                  </>
                ) : (
                  <>
                    <Check className="w-4 h-4" />
                    创建
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {showFilterModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white">筛选虚拟机</h3>
              <button
                onClick={() => setShowFilterModal(false)}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-text-muted mb-1">状态</label>
                <select
                  value={filters.status}
                  onChange={(e) => setFilters({ ...filters, status: e.target.value })}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                >
                  <option value="">全部</option>
                  <option value="running">运行中</option>
                  <option value="poweredOff">已停止</option>
                  <option value="warning">警告</option>
                  <option value="unknown">未知</option>
                </select>
              </div>
              <div>
                <label className="block text-sm text-text-muted mb-1">操作系统</label>
                <select
                  value={filters.os}
                  onChange={(e) => setFilters({ ...filters, os: e.target.value })}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-primary"
                >
                  <option value="">全部</option>
                  <option value="Linux">Linux</option>
                  <option value="Windows">Windows</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={handleClearFilters}
                className="px-4 py-2 text-text-muted hover:text-white transition-colors"
              >
                清除筛选
              </button>
              <button
                onClick={handleApplyFilters}
                className="btn-primary flex items-center gap-2"
              >
                <Check className="w-4 h-4" />
                应用筛选
              </button>
            </div>
          </div>
        </div>
      )}

      {showExportModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white flex items-center gap-2">
                <FileSpreadsheet className="w-5 h-5" />
                导出虚拟机列表
              </h3>
              <button
                onClick={() => setShowExportModal(false)}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X className="w-5 h-5 text-text-muted" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm text-text-muted mb-2">导出格式</label>
                <div className="flex gap-3">
                  <button
                    onClick={() => setExportFormat('csv')}
                    className={`flex-1 py-3 rounded-lg border transition-colors flex items-center justify-center gap-2 ${
                      exportFormat === 'csv'
                        ? 'bg-primary/20 border-primary text-primary'
                        : 'border-border text-text-secondary hover:text-white'
                    }`}
                  >
                    <FileSpreadsheet className="w-4 h-4" />
                    CSV
                  </button>
                  <button
                    onClick={() => setExportFormat('json')}
                    className={`flex-1 py-3 rounded-lg border transition-colors flex items-center justify-center gap-2 ${
                      exportFormat === 'json'
                        ? 'bg-primary/20 border-primary text-primary'
                        : 'border-border text-text-secondary hover:text-white'
                    }`}
                  >
                    <FileSpreadsheet className="w-4 h-4" />
                    JSON
                  </button>
                </div>
              </div>

              <div className="bg-gray-700/50 rounded-lg p-4">
                <div className="text-sm text-text-muted mb-2">导出信息</div>
                <div className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-text-muted">虚拟机数量:</span>
                    <span className="text-white">{vms.length}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-text-muted">包含实时指标:</span>
                    <span className="text-white">{metrics.length > 0 ? '是' : '否'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-text-muted">包含位置信息:</span>
                    <span className="text-white">是</span>
                  </div>
                </div>
              </div>

              <div className="flex justify-end gap-3 mt-6">
                <button
                  onClick={() => setShowExportModal(false)}
                  className="px-4 py-2 text-text-muted hover:text-white transition-colors"
                >
                  取消
                </button>
                <button
                  onClick={handleExport}
                  disabled={exporting}
                  className="btn-primary flex items-center gap-2"
                >
                  {exporting ? (
                    <>
                      <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                      导出中...
                    </>
                  ) : (
                    <>
                      <Download className="w-4 h-4" />
                      导出
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default VMList;
