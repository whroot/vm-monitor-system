/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // 主色调
        background: '#0f1419',
        surface: '#1a1f2e',
        border: '#2a3441',
        
        // 状态色
        success: '#00d4aa',
        warning: '#ff9800',
        danger: '#f44336',
        info: '#2196f3',
        
        // 文字色
        'text-primary': '#ffffff',
        'text-secondary': '#e0e0e0',
        'text-tertiary': '#b0b8c5',
        'text-muted': '#8090a0',
        
        // 数据可视化色
        'chart-cpu': '#2196f3',
        'chart-memory': '#4caf50',
        'chart-disk': '#ff9800',
        'chart-network': '#9c27b0',
      },
      fontFamily: {
        sans: ['Inter', 'PingFang SC', 'Hiragino Sans', 'Microsoft YaHei', 'Meiryo', 'sans-serif'],
        mono: ['JetBrains Mono', 'Consolas', 'Monaco', 'monospace'],
      },
      fontSize: {
        '2xs': '11px',
        'xs': '12px',
        'sm': '14px',
        'base': '16px',
        'lg': '18px',
        'xl': '20px',
        '2xl': '24px',
      },
      spacing: {
        '18': '4.5rem',
        '22': '5.5rem',
      },
      borderRadius: {
        '2xl': '16px',
        '3xl': '24px',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  plugins: [],
}
