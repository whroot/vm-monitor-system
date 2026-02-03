import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// 导入翻译文件
import zhCN from './locales/zh-CN.json';
import en from './locales/en.json';
import jaJP from './locales/ja-JP.json';

const resources = {
  'zh-CN': {
    translation: zhCN
  },
  'en': {
    translation: en
  },
  'ja-JP': {
    translation: jaJP
  }
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    lng: localStorage.getItem('language') || 'zh-CN',
    interpolation: {
      escapeValue: false,
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
  });

export default i18n;