import path from "path"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      // API requests go to backend (Gin server)
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // 移除 /api 前缀，直接透传到后端
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
      // 前端不直接连接 VictoriaLogs，所有请求都通过后端
      // 后端会在 /select 和 /metrics 路径上代理到 VictoriaLogs
    },
  },
})