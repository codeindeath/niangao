import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  base: '/admin/',
  server: {
    proxy: {
      '/api': {
        target: 'http://115.190.177.146',
        changeOrigin: true,
      },
    },
  },
})
