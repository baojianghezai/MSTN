// 前台 UnoCSS 配置（与后台同款原子类引擎，presetWind3 = Tailwind v3 风格）
// 原子类（flex / grid / text-* / max-w-* 等）由 UnoCSS 在构建时扫描 .vue 生成真实 CSS
import { defineConfig } from 'unocss'
import presetWind3 from '@unocss/preset-wind3'
import presetIcons from '@unocss/preset-icons'
import transformerDirectives from '@unocss/transformer-directives'

// 设计 tokens（B1 美化 P0 2026-08-20 / P1 补 danger 色阶与图标预设）：
// 三页散落品牌蓝（#2B6BF0/#4A8BFF/#3B82F6）收敛为单一 primary 色板（500=#3B82F6 锚点）；
// Element Plus 组件色在 styles/index.css 用 --el-color-primary* 同步覆盖，原子类与 EP 组件共用同一品牌色。
// 使用规范：品牌色一律 primary-*；薪资/数字强调用 accent-*（橙）；危险操作/删除/错误态用 danger-*（#ef4444，
// 与 accent 橙区分：accent 是「强调」不是「警告」）；卡片投影用 shadow-card / shadow-card-hover；
// 图标用 presetIcons（i-lucide-* 原子类，lucide 线框风格，与 GVA 图标规范一致）；
// 间距沿用 Tailwind 默认刻度（模块间距 mt-6/py-8，卡片内 gap-4）；
// 字体层级：页面标题 text-2xl > 区块标题 text-lg > 正文 text-sm > 辅助说明 text-xs。
export default defineConfig({
  presets: [
    presetWind3({ dark: 'class' }),
    presetIcons({
      scale: 1.2,
      extraProperties: {
        display: 'inline-block',
        'vertical-align': 'middle'
      }
    })
  ],
  transformers: [transformerDirectives()],
  theme: {
    colors: {
      primary: {
        50: '#eff6ff',
        100: '#dbeafe',
        200: '#bfdbfe',
        300: '#93c5fd',
        400: '#60a5fa',
        500: '#3b82f6',
        600: '#2563eb',
        700: '#1d4ed8',
        800: '#1e40af',
        900: '#1e3a8a'
      },
      accent: {
        50: '#fff7ed',
        100: '#ffedd5',
        200: '#fed7aa',
        300: '#fdba74',
        400: '#fb923c',
        500: '#f97316',
        600: '#ea580c',
        700: '#c2410c'
      },
      danger: {
        50: '#fef2f2',
        100: '#fee2e2',
        200: '#fecaca',
        300: '#fca5a5',
        400: '#f87171',
        500: '#ef4444',
        600: '#dc2626',
        700: '#b91c1c'
      }
    },
    boxShadow: {
      card: '0 1px 3px rgba(15, 23, 42, 0.06), 0 1px 2px rgba(15, 23, 42, 0.04)',
      'card-hover': '0 4px 12px rgba(15, 23, 42, 0.10)'
    }
  }
})
