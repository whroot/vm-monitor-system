import React from 'react';
import { useTranslation } from 'react-i18next';
import { History, Calendar, Download } from 'lucide-react';

const HistoryData: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
          <History className="w-6 h-6" />
          {t('history')}
        </h1>
        <button className="btn-secondary flex items-center gap-2">
          <Download className="w-4 h-4" />
          {t('export')}
        </button>
      </div>

      {/* Filters */}
      <div className="card">
        <div className="flex gap-4 flex-wrap">
          <div className="flex items-center gap-2">
            <Calendar className="w-5 h-5 text-text-muted" />
            <span className="text-text-secondary">时间范围:</span>
            <button className="px-4 py-2 bg-background border border-border rounded-lg text-sm text-white">
              最近7天
            </button>
          </div>
        </div>
      </div>

      {/* Chart Placeholder */}
      <div className="card h-96 flex items-center justify-center text-text-muted">
        历史数据分析图表区域
      </div>
    </div>
  );
};

export default HistoryData;