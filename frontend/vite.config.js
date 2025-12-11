import { defineConfig } from 'vite'
import preact from '@preact/preset-vite'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [preact(), tailwindcss()],
  // Base path for assets - set via VITE_BASE_PATH env var or defaults to '/'
  // For deployment at /admin, build with: VITE_BASE_PATH=/admin/ npm run build
  base: process.env.VITE_BASE_PATH || '/',
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
