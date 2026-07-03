import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import path from 'node:path'
import unoConfig from '../uno.config'

export default defineConfig({
  base: process.env.PROAPI_USER_BASE || '/',
  plugins: [vue(), UnoCSS(unoConfig)],
  resolve: { alias: { '@': path.resolve(__dirname, 'src') } },
  server: {
    proxy: {
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: true },
      '/v1': { target: 'http://127.0.0.1:8080', changeOrigin: true },
    },
  },
  build: { outDir: 'dist', emptyOutDir: true },
})
