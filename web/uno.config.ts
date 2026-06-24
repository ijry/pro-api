import { defineConfig, presetUno, presetAttributify, presetIcons, transformerVariantGroup } from 'unocss'
import path from 'node:path'

/**
 * proapi 共享 UnoCSS 配置 — admin 与 user 各自 vite.config 引入并按需扩展。
 * 主题 token 通过 CSS 变量在运行时切换深浅色。
 */
export default defineConfig({
  presets: [
    presetUno({ dark: 'class' }),
    presetAttributify(),
    presetIcons({
      scale: 1.1,
      collectionsNodeResolvePath: path.resolve(__dirname),
    }),
  ],
  transformers: [transformerVariantGroup()],
  theme: {
    colors: {
      primary: 'rgb(var(--c-primary) / <alpha-value>)',
      'primary-hover': 'rgb(var(--c-primary-hover) / <alpha-value>)',
      bg: 'rgb(var(--c-bg) / <alpha-value>)',
      'bg-elevated': 'rgb(var(--c-bg-elevated) / <alpha-value>)',
      fg: 'rgb(var(--c-fg) / <alpha-value>)',
      'fg-muted': 'rgb(var(--c-fg-muted) / <alpha-value>)',
      border: 'rgb(var(--c-border) / <alpha-value>)',
    },
  },
  shortcuts: {
    'btn-primary': 'inline-flex items-center justify-center h-9 px-4 rounded-md bg-primary text-white font-medium hover:bg-primary-hover transition-colors',
    card: 'rounded-lg border border-border bg-bg-elevated p-4',
  },
})
