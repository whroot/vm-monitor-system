import React, { useState, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { History, Download, Server, Clock, RefreshCw, Calendar, X, Check, FileSpreadsheet } from 'lucide-react';
import { 
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer
} from 'recharts';
import { vmApi } from '../../api/vm';
import { VM, VMMetricsHistoryRecord } from '../../types/api';

type Period = '1h' | '6h' | '24h' | '7d' | '30d' | 'custom';

const HISTORY_PERIODS: { label: string; value: Period }[] = [
  { label: '最近1小时', value: '1h' },
  { label: '最近6小时', value: '6h' },
  { label: '最近24小时', value: '24h' },
  { label: '最近7天', value: '7d' },
  { label: '最近30天', value: '30d' },
  { label: '自定义', value: 'custom' },
];

const METRICS_CONFIG = [
  { key: 'cpuUsage', label: 'CPU使用率', unit: '%', color: '#3b82f6' },
  { key: 'memoryUsage', label: '内存使用率', unit: '%', color: '#10b981' },
  { key: 'diskUsage', label: '磁盘使用率', unit: '%', color: '#f59e0b' },
  { key: 'diskReadMbps', label: '磁盘读取', unit: 'MB/s', color: '#8b5cf6' },
  { key: 'diskWriteMbps', label: '磁盘写入', unit: 'MB/s', color: '#ec4899' },
  { key: 'networkInMbps', label: '网络入站', unit: 'MB/s', color: '#06b6d4' },
  { key: 'networkOutMbps', label: '网络出站', unit: 'MB/s', color: '#14b8a6' },
  { key: 'temperature', label: '温度', unit: '°C', color: '#ef4444' },
];

const VM_COLORS = [
  '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', 
  '#ec4899', '#06b6d4', '#14b8a6', '#f97316', '#84cc16'
];

interface VMHistoryData {
  vmId: string;
  vmName: string;
  metrics: VMMetricsHistoryRecord[];
  color: string;
}

type ExportFormat = 'csv' | 'json';

const HistoryData: React.FC = () => {
  const { t } = useTranslation();
  const [vms, setVMs] = useState<VM[]>([]);
  const [selectedVMIds, setSelectedVMIds] = useState<string[]>([]);
  const [showVMSelector, setShowVMSelector] = useState(false);
  const [selectedPeriod, setSelectedPeriod] = useState<Period>('24h');
  const [customStartTime, setCustomStartTime] = useState<string>('');
  const [customEndTime, setCustomEndTime] = useState<string>('');
  const [selectedMetrics, setSelectedMetrics] = useState<string[]>(['cpuUsage', 'memoryUsage']);
  const [vmHistoryData, setVMHistoryData] = useState<VMHistoryData[]>([]);
  const [loading, setLoading] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [exportFormat, setExportFormat] = useState<ExportFormat>('csv');
  const [exporting, setExporting] = useState(false);

  useEffect(() => {
    loadVMs();
  }, []);

  useEffect(() => {
    if (selectedVMIds.length > 0) {
      loadHistoryData();
    } else {
      setVMHistoryData([]);
    }
  }, [selectedVMIds, selectedPeriod]);

  const loadVMs = async () => {
    try {
      const response = await vmApi.list({ page: 1, pageSize: 100 });
      setVMs(response.list);
      if (response.list.length > 0 && selectedVMIds.length === 0) {
        setSelectedVMIds([response.list[0].id]);
      }
    } catch (error) {
      console.error('Failed to load VMs:', error);
    }
  };

  const loadHistoryData = async () => {
    if (selectedVMIds.length === 0) return;
    setLoading(true);
    try {
      let startTime: string | undefined;
      let endTime: string | undefined;
      
      if (selectedPeriod === 'custom') {
        startTime = customStartTime;
        endTime = customEndTime;
      }
      
      const promises = selectedVMIds.map(async (vmId, index) => {
        const data = await vmApi.getMetricsHistory(vmId, selectedPeriod !== 'custom' ? selectedPeriod : '24h', startTime, endTime);
        const vm = vms.find(v => v.id === vmId);
        return {
          vmId,
          vmName: vm?.name || vmId,
          metrics: data.metrics,
          color: VM_COLORS[index % VM_COLORS.length],
        };
      });
      
      const results = await Promise.all(promises);
      setVMHistoryData(results);
    } catch (error) {
      console.error('Failed to load history data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handlePeriodChange = (period: Period) => {
    setSelectedPeriod(period);
    if (period !== 'custom') {
      setCustomStartTime('');
      setCustomEndTime('');
    }
  };

  const handleCustomTimeApply = () => {
    if (customStartTime && customEndTime && new Date(customStartTime) < new Date(customEndTime)) {
      loadHistoryData();
    }
  };

  const toggleMetric = (key: string) => {
    setSelectedMetrics(prev => 
      prev.includes(key) 
        ? prev.filter(k => k !== key)
        : [...prev, key]
    );
  };

  const toggleVMSelection = (vmId: string) => {
    setSelectedVMIds(prev => 
      prev.includes(vmId)
        ? prev.filter(id => id !== vmId)
        : [...prev, vmId]
    );
  };

  const selectAllVMs = () => {
    setSelectedVMIds(vms.map(v => v.id));
  };

  const deselectAllVMs = () => {
    setSelectedVMIds([]);
  };

  const handleExport = async () => {
    if (vmHistoryData.length === 0) return;
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
    const headers = ['时间戳', '时间'];
    vmHistoryData.forEach(vm => {
      selectedMetrics.forEach(metric => {
        const metricConfig = METRICS_CONFIG.find(m => m.key === metric);
        if (metricConfig) {
          headers.push(`${vm.vmName}-${metricConfig.label}(${metricConfig.unit})`);
        }
      });
    });

    const rows: string[][] = [headers];

    vmHistoryData[0]?.metrics.forEach(metric => {
      const row = [
        metric.timestamp,
        formatTimestamp(metric.timestamp)
      ];
      vmHistoryData.forEach(vm => {
        const vmMetric = vm.metrics.find(m => m.timestamp === metric.timestamp);
        selectedMetrics.forEach(key => {
          row.push(vmMetric?.[key as keyof VMMetricsHistoryRecord]?.toString() || '');
        });
      });
      rows.push(row);
    });

    const csvContent = rows.map(row => row.map(cell => `"${cell}"`).join(',')).join('\n');
    downloadFile(csvContent, `vm-metrics-${timestamp}.csv`, 'text/csv');
  };

  const exportJSON = (timestamp: string) => {
    const exportData = {
      exportTime: new Date().toISOString(),
      timeRange: selectedPeriod,
      vms: vmHistoryData.map(vm => ({
        vmId: vm.vmId,
        vmName: vm.vmName,
        metrics: vm.metrics
      }))
    };

    const jsonContent = JSON.stringify(exportData, null, 2);
    downloadFile(jsonContent, `vm-metrics-${timestamp}.json`, 'application/json');
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

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    if (selectedPeriod === '1h' || selectedPeriod === '6h') {
      return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    }
    return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  const getCombinedData = () => {
    if (vmHistoryData.length === 0 || vmHistoryData[0].metrics.length === 0) return [];
    
    const timestamps = new Set<string>();
    vmHistoryData.forEach(data => {
      data.metrics.forEach(m => timestamps.add(m.timestamp));
    });
    
    const sortedTimestamps = Array.from(timestamps).sort((a, b) => 
      new Date(a).getTime() - new Date(b).getTime()
    );
    
    return sortedTimestamps.map(timestamp => {
      const record: any = { timestamp };
      vmHistoryData.forEach(data => {
        const metricData = data.metrics.find(m => m.timestamp === timestamp);
        selectedMetrics.forEach(key => {
          record[`${data.vmId}_${key}`] = metricData?.[key as keyof VMMetricsHistoryRecord] || null;
        });
      });
      return record;
    });
  };

  const combinedData = useMemo(() => getCombinedData(), [vmHistoryData, selectedMetrics]);

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
          <History className="w-6 h-6" />
          {t('history')}
        </h1>
        <button
          onClick={() => setShowExportModal(true)}
          disabled={vmHistoryData.length === 0}
          className="btn-secondary flex items-center gap-2 disabled:opacity-50"
        >
          <Download className="w-4 h-4" />
          {t('export')}
        </button>
      </div>

      <div className="card">
        <div className="flex gap-6 flex-wrap">
          <div className="relative">
            <div className="flex items-center gap-3">
              <Server className="w-5 h-5 text-text-muted" />
              <span className="text-text-secondary">虚拟机:</span>
              <button
                onClick={() => setShowVMSelector(!showVMSelector)}
                className="bg-background border border-border rounded-lg px-3 py-2 text-white text-sm min-w-48 focus:outline-none focus:border-primary flex items-center justify-between gap-2"
              >
                <span>{selectedVMIds.length === 0 ? '请选择虚拟机' : `已选择 ${selectedVMIds.length} 个`}</span>
                <span className="text-text-muted">▼</span>
              </button>
            </div>

            {showVMSelector && (
              <div className="absolute top-full left-8 mt-2 w-80 bg-gray-800 border border-border rounded-lg shadow-xl z-50 max-h-64 overflow-y-auto">
                <div className="flex justify-between items-center p-3 border-b border-border">
                  <span className="text-sm text-text-muted">选择虚拟机</span>
                  <div className="flex gap-2">
                    <button
                      onClick={selectAllVMs}
                      className="text-xs text-primary hover:text-primary/80"
                    >
                      全选
                    </button>
                    <span className="text-text-muted">|</span>
                    <button
                      onClick={deselectAllVMs}
                      className="text-xs text-text-muted hover:text-white"
                    >
                      取消全选
                    </button>
                  </div>
                </div>
                {vms.map(vm => (
                  <label
                    key={vm.id}
                    className="flex items-center gap-3 px-3 py-2 hover:bg-gray-700 cursor-pointer"
                  >
                    <div className={`w-4 h-4 rounded border flex items-center justify-center ${
                      selectedVMIds.includes(vm.id)
                        ? 'bg-primary border-primary'
                        : 'border-border'
                    }`}>
                      {selectedVMIds.includes(vm.id) && <Check className="w-3 h-3 text-white" />}
                    </div>
                    <input
                      type="checkbox"
                      checked={selectedVMIds.includes(vm.id)}
                      onChange={() => toggleVMSelection(vm.id)}
                      className="hidden"
                    />
                    <Server className="w-4 h-4 text-text-muted" />
                    <span className="text-white text-sm">{vm.name}</span>
                    <span className="text-xs text-text-muted ml-auto">{vm.ip || '-'}</span>
                  </label>
                ))}
              </div>
            )}
          </div>

          <div className="flex items-center gap-3">
            <Clock className="w-5 h-5 text-text-muted" />
            <span className="text-text-secondary">时间范围:</span>
            <div className="flex gap-1 flex-wrap">
              {HISTORY_PERIODS.map(period => (
                <button
                  key={period.value}
                  onClick={() => handlePeriodChange(period.value)}
                  className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                    selectedPeriod === period.value
                      ? 'bg-primary text-white'
                      : 'bg-background border border-border text-text-secondary hover:text-white'
                  }`}
                >
                  {period.label}
                </button>
              ))}
            </div>
          </div>

          {selectedPeriod === 'custom' && (
            <div className="flex items-center gap-3 ml-4">
              <Calendar className="w-5 h-5 text-text-muted" />
              <input
                type="datetime-local"
                value={customStartTime}
                onChange={(e) => setCustomStartTime(e.target.value)}
                className="bg-background border border-border rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary"
              />
              <span className="text-text-muted">至</span>
              <input
                type="datetime-local"
                value={customEndTime}
                onChange={(e) => setCustomEndTime(e.target.value)}
                className="bg-background border border-border rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary"
              />
              <button
                onClick={handleCustomTimeApply}
                className="btn-primary px-4 py-2 text-sm"
              >
                应用
              </button>
            </div>
          )}

          <button
            onClick={loadHistoryData}
            disabled={loading || selectedVMIds.length === 0}
            className="btn-secondary flex items-center gap-2 ml-auto disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </button>
        </div>
      </div>

      {selectedVMIds.length > 0 && (
        <div className="flex gap-2 flex-wrap">
          {vmHistoryData.map(data => (
            <div
              key={data.vmId}
              className="flex items-center gap-2 px-3 py-1.5 bg-surface border border-border rounded-lg text-sm"
            >
              <span
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: data.color }}
              />
              <span className="text-white">{data.vmName}</span>
              <button
                onClick={() => toggleVMSelection(data.vmId)}
                className="text-text-muted hover:text-white"
              >
                <X className="w-3 h-3" />
              </button>
            </div>
          ))}
        </div>
      )}

      <div className="card">
        <div className="flex items-center gap-3 mb-4">
          <span className="text-text-secondary">显示指标:</span>
          <div className="flex gap-2 flex-wrap">
            {METRICS_CONFIG.map(metric => (
              <button
                key={metric.key}
                onClick={() => toggleMetric(metric.key)}
                className={`px-3 py-1 rounded-lg text-sm transition-colors flex items-center gap-2 ${
                  selectedMetrics.includes(metric.key)
                    ? 'bg-opacity-20'
                    : 'opacity-50'
                }`}
                style={{
                  backgroundColor: selectedMetrics.includes(metric.key) ? `${metric.color}20` : 'transparent',
                  borderColor: selectedMetrics.includes(metric.key) ? metric.color : 'transparent',
                  color: selectedMetrics.includes(metric.key) ? metric.color : 'text-text-muted',
                }}
              >
                <span
                  className="w-2 h-2 rounded-full"
                  style={{ backgroundColor: metric.color }}
                />
                {metric.label}
              </button>
            ))}
          </div>
        </div>

        {loading ? (
          <div className="h-96 flex items-center justify-center text-text-muted">
            加载中...
          </div>
        ) : vmHistoryData.length === 0 ? (
          <div className="h-96 flex items-center justify-center text-text-muted">
            请选择虚拟机查看历史数据
          </div>
        ) : (
          <div className="space-y-6">
            <div className="card bg-surface">
              <h3 className="text-white font-medium mb-4">多VM综合趋势</h3>
              <div className="h-96">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={combinedData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                    <XAxis
                      dataKey="timestamp"
                      tickFormatter={formatTimestamp}
                      stroke="#9ca3af"
                      fontSize={12}
                    />
                    <YAxis stroke="#9ca3af" fontSize={12} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1f2937',
                        border: '1px solid #374151',
                        borderRadius: '8px',
                      }}
                      labelFormatter={formatTimestamp}
                    />
                    <Legend />
                    {vmHistoryData.flatMap(vmData =>
                      selectedMetrics.map(metricKey => {
                        const metricConfig = METRICS_CONFIG.find(m => m.key === metricKey);
                        if (!metricConfig) return null;
                        return (
                          <Line
                            key={`${vmData.vmId}_${metricKey}`}
                            type="monotone"
                            dataKey={`${vmData.vmId}_${metricKey}`}
                            name={`${vmData.vmName} - ${metricConfig.label}`}
                            stroke={vmData.color}
                            strokeWidth={2}
                            dot={false}
                            activeDot={{ r: 4 }}
                            connectNulls
                          />
                        );
                      }).filter(Boolean)
                    )}
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>

            {selectedMetrics.map(metricKey => {
              const config = METRICS_CONFIG.find(m => m.key === metricKey);
              if (!config) return null;

              return (
                <div key={metricKey} className="card bg-surface">
                  <h3 className="text-white font-medium mb-4 flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: config.color }}
                    />
                    {config.label} ({config.unit})
                  </h3>
                  <div className="h-64">
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart data={combinedData}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                        <XAxis
                          dataKey="timestamp"
                          tickFormatter={formatTimestamp}
                          stroke="#9ca3af"
                          fontSize={12}
                        />
                        <YAxis stroke="#9ca3af" fontSize={12} />
                        <Tooltip
                          contentStyle={{
                            backgroundColor: '#1f2937',
                            border: '1px solid #374151',
                            borderRadius: '8px',
                          }}
                          labelFormatter={formatTimestamp}
                          formatter={(value: number) => [`${value?.toFixed(2) || '-'} ${config.unit}`, config.label]}
                        />
                        <Legend />
                        {vmHistoryData.map(vmData => (
                          <Line
                            key={`${vmData.vmId}_${metricKey}`}
                            type="monotone"
                            dataKey={`${vmData.vmId}_${metricKey}`}
                            name={vmData.vmName}
                            stroke={vmData.color}
                            strokeWidth={2}
                            dot={false}
                            activeDot={{ r: 4 }}
                            connectNulls
                          />
                        ))}
                      </LineChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className="grid grid-cols-4 gap-4">
        {vmHistoryData.length > 0 && (
          <>
            <div className="card text-center">
              <div className="text-2xl font-bold text-primary">
                {vmHistoryData.length}
              </div>
              <div className="text-sm text-text-muted">监控VM数</div>
            </div>
            <div className="card text-center">
              <div className="text-2xl font-bold text-success">
                {vmHistoryData[0]?.metrics[vmHistoryData[0]?.metrics.length - 1]?.cpuUsage?.toFixed(1) || '-'}%
              </div>
              <div className="text-sm text-text-muted">{vmHistoryData[0]?.vmName} CPU</div>
            </div>
            <div className="card text-center">
              <div className="text-2xl font-bold text-warning">
                {vmHistoryData[0]?.metrics[vmHistoryData[0]?.metrics.length - 1]?.memoryUsage?.toFixed(1) || '-'}%
              </div>
              <div className="text-sm text-text-muted">{vmHistoryData[0]?.vmName} 内存</div>
            </div>
            <div className="card text-center">
              <div className="text-2xl font-bold text-info">
                {selectedPeriod}
              </div>
              <div className="text-sm text-text-muted">时间范围</div>
            </div>
          </>
        )}
      </div>

      {showExportModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md animate-fade-in">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-white flex items-center gap-2">
                <FileSpreadsheet className="w-5 h-5" />
                导出历史数据
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
                    <span className="text-white">{vmHistoryData.length}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-text-muted">时间范围:</span>
                    <span className="text-white">{selectedPeriod}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-text-muted">导出指标:</span>
                    <span className="text-white">{selectedMetrics.length}个</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-text-muted">数据点数:</span>
                    <span className="text-white">{vmHistoryData[0]?.metrics.length || 0}</span>
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

export default HistoryData;
