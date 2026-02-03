
import React, { useState, useRef } from 'react';
import { useTranslation } from './LanguageContext';
import { Sparkles, Upload, Download, RefreshCw, Wand2 } from 'lucide-react';
import { editImageWithAI } from '../services/geminiService';

const ImageEditor: React.FC = () => {
  const { t } = useTranslation();
  const [image, setImage] = useState<string | null>(null);
  const [prompt, setPrompt] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (readerEvent) => {
        setImage(readerEvent.target?.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleEdit = async () => {
    if (!image || !prompt) return;
    setIsProcessing(true);
    try {
      const result = await editImageWithAI(image, prompt);
      if (result) setImage(result);
    } catch (err) {
      alert("Failed to edit image. Ensure API key is set.");
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold flex items-center gap-2">
          <Sparkles className="text-purple-400" /> {t('editImage')}
        </h2>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 bg-[#1a1f2e] border border-[#2a3441] rounded-3xl p-6 min-h-[400px] flex flex-col items-center justify-center relative overflow-hidden group">
          {image ? (
            <div className="w-full h-full flex flex-col items-center">
              <img src={image} className="max-h-[500px] rounded-xl shadow-2xl mb-4 transition-all" alt="Monitor Visual" />
              <div className="flex gap-4">
                <button 
                  onClick={() => setImage(null)}
                  className="px-4 py-2 bg-red-500/10 text-red-500 rounded-xl hover:bg-red-500/20 transition-all font-bold text-sm"
                >
                  Clear
                </button>
                <a href={image} download="vmsentry_report.png" className="px-4 py-2 bg-[#2a3441] rounded-xl hover:bg-[#3a4451] transition-all font-bold text-sm flex items-center gap-2">
                  <Download size={16} /> Download
                </a>
              </div>
            </div>
          ) : (
            <>
              <div className="w-20 h-20 bg-[#2a3441] rounded-full flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <Upload size={32} className="text-[#8090a0]" />
              </div>
              <p className="text-[#b0b8c5] mb-4">Upload a hardware photo or rack screenshot for AI labeling</p>
              <button 
                onClick={() => fileInputRef.current?.click()}
                className="px-8 py-3 bg-[#4caf50] rounded-xl font-bold text-white shadow-lg shadow-[#4caf50]/20"
              >
                Choose Photo
              </button>
              <input ref={fileInputRef} type="file" className="hidden" accept="image/*" onChange={handleUpload} />
            </>
          )}
        </div>

        <div className="bg-[#1a1f2e] border border-[#2a3441] rounded-3xl p-6 space-y-6">
          <div>
            <label className="block text-sm font-bold text-[#8090a0] mb-2 uppercase tracking-widest">AI Command</label>
            <textarea 
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder={t('promptPlaceholder')}
              className="w-full bg-[#0f1419] border border-[#2a3441] rounded-xl p-4 min-h-[120px] focus:border-purple-500 outline-none transition-all text-sm resize-none"
            />
          </div>

          <button 
            onClick={handleEdit}
            disabled={!image || !prompt || isProcessing}
            className={`w-full py-4 rounded-xl font-bold flex items-center justify-center gap-3 transition-all
              ${!image || !prompt || isProcessing ? 'bg-[#2a3441] text-[#607080] cursor-not-allowed' : 'bg-gradient-to-r from-purple-600 to-indigo-600 text-white shadow-lg shadow-purple-500/20 hover:scale-[1.02]'}
            `}
          >
            {isProcessing ? (
              <RefreshCw className="animate-spin" size={20} />
            ) : (
              <Wand2 size={20} />
            )}
            {isProcessing ? t('generating') : 'Run AI Processing'}
          </button>

          <div className="p-4 bg-purple-500/5 border border-purple-500/20 rounded-2xl">
             <h4 className="text-xs font-bold text-purple-400 uppercase mb-2">Example Prompts:</h4>
             <ul className="text-xs text-[#8090a0] space-y-2">
                <li>• "Add red arrows pointing to loose power cables"</li>
                <li>• "Highlight the overheated server rack with a glow"</li>
                <li>• "Annotate this screenshot with CPU usage stats"</li>
             </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ImageEditor;
