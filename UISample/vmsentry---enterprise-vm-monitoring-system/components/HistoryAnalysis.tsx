
import React, { useState } from 'react';
import { useTranslation } from './LanguageContext';
import { 
  History, 
  Lightbulb, 
  ArrowRight, 
  Calendar,
  Sparkles,
  Search
} from 'lucide-react';
import { analyzeAlerts } from '../services/geminiService';

const HistoryAnalysis: React.FC = () => {
  const { t, language } = useTranslation();
  const [view, setView] = useState<'troubleshoot' | 'planning'>('troubleshoot');
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [aiInsight, setAiInsight] = useState<string | null>(null);

  const mockAlerts = [
    "Critical: CPU load exceeded 95% on prod-web-01 at 08:32 AM",
    "Warning: Memory utilization at 88% on prod-web-01 at 08:35 AM",
    "System: Automatic scaling failed due to resource exhaustion at 08:36 AM"
  ];

  const handleRunAI = async () => {
    setIsAnalyzing(true);
    const result = await analyzeAlerts(mockAlerts, language);
    setAiInsight(result);
    setIsAnalyzing(false);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex bg-[#1a1f2e] p-1.5 rounded-2xl border border-[#2a3441]">
          <button 
            onClick={() => setView('troubleshoot')}
            className={`px-6 py-2 rounded-xl text-sm font-bold transition-all ${view === 'troubleshoot' ? 'bg-[#2196f3] text-white shadow-lg shadow-[#2196f3]/20' : 'text-[#b0b8c5] hover:text-white'}`}
          >
            {t('troubleshoot')}
          </button>
          <button 
            onClick={() => setView('planning')}
            className={`px-6 py-2 rounded-xl text-sm font-bold transition-all ${view === 'planning' ? 'bg-[#9c27b0] text-white shadow-lg shadow-[#9c27b0]/20' : 'text-[#b0b8c5] hover:text-white'}`}
          >
            {t('capacityPlanning')}
          </button>
        </div>

        <div className="flex gap-4">
           <div className="flex items-center gap-2 bg-[#1a1f2e] border border-[#2a3441] rounded-xl px-4 py-2 text-sm text-[#b0b8c5]">
              <Calendar size={16} /> Last 7 Days
           </div>
           <button className="bg-[#1a1f2e] border border-[#2a3441] rounded-xl px-4 py-2 text-sm font-bold hover:bg-[#2a3441] transition-all">
              Export PDF
           </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-[#1a1f2e] border border-[#2a3441] p-8 rounded-3xl min-h-[400px] flex flex-col items-center justify-center text-center">
             <div className="w-16 h-16 bg-[#2196f3]/10 text-[#2196f3] rounded-2xl flex items-center justify-center mb-4">
                <History size={32} />
             </div>
             <h3 className="text-xl font-bold mb-2">History Log Analysis</h3>
             <p className="text-[#b0b8c5] max-w-sm mb-6">Select a timeframe and specific VMs to view the historical performance correlation and event markers.</p>
             <button className="bg-[#2196f3] px-8 py-3 rounded-xl font-bold hover:shadow-lg hover:shadow-[#2196f3]/20 transition-all flex items-center gap-2">
                Launch Chart Viewer <ArrowRight size={18} />
             </button>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-[#1a1f2e] border border-[#2a3441] rounded-3xl p-6 relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-32 h-32 bg-[#4caf50]/5 rounded-full -mr-16 -mt-16 group-hover:bg-[#4caf50]/10 transition-all"></div>
            <div className="flex items-center gap-3 mb-6">
              <div className="p-2 bg-[#4caf50]/10 text-[#4caf50] rounded-lg">
                <Sparkles size={20} />
              </div>
              <h4 className="font-bold">AI Fault Insights</h4>
            </div>

            <div className="space-y-4 mb-6">
               {mockAlerts.map((a, i) => (
                 <div key={i} className="text-xs bg-[#0f1419] p-3 rounded-xl text-[#8090a0] font-mono border border-[#2a3441]">
                   {a}
                 </div>
               ))}
            </div>

            <button 
              onClick={handleRunAI}
              disabled={isAnalyzing}
              className={`w-full py-3 rounded-xl font-bold transition-all flex items-center justify-center gap-2 ${isAnalyzing ? 'bg-[#2a3441] text-[#607080]' : 'bg-[#4caf50] text-white hover:scale-[1.02] shadow-lg shadow-[#4caf50]/20'}`}
            >
              {isAnalyzing ? t('generating') : 'Analyze with Gemini'}
            </button>

            {aiInsight && (
              <div className="mt-6 p-4 bg-[#0f1419] border border-[#4caf50]/30 rounded-2xl text-sm text-[#b0b8c5] animate-fadeIn">
                <div className="flex items-center gap-2 mb-2 text-[#4caf50] font-bold text-xs uppercase">
                  <Lightbulb size={14} /> Root Cause Summary
                </div>
                {aiInsight}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default HistoryAnalysis;
