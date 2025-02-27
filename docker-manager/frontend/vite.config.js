import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 3000,  // 指定前端端口
    proxy: {
      '/api': {
        target: 'http://0.0.0.0:8080',  // 修改为后端地址，使用 0.0.0.0
        changeOrigin: true,
        secure: false
      }
    }
  }
})
