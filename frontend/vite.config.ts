import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { codeInspectorPlugin } from 'code-inspector-plugin';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(),
    codeInspectorPlugin({
      bundler: 'vite',
    })],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    sourcemap: true
  },
  server: {
    proxy: {
      '^/api/*': {
        target: 'http://172.16.100.221:8081',
        changeOrigin: true,
        ws: true,
      }
    }
  },
})
