import { defineConfig } from 'vite'
import preact from '@preact/preset-vite'
import tailwindcss from '@tailwindcss/vite'

// Base path configurable via VITE_BASE_PATH (default: /tinylist/)
const basePath = process.env.VITE_BASE_PATH || '/tinylist/'

// https://vite.dev/config/
export default defineConfig({
  plugins: [preact(), tailwindcss()],
  base: basePath,
  define: {
    __BASE_PATH__: JSON.stringify(basePath.replace(/\/$/, '')), // without trailing slash
  },
  server: {
    proxy: {
      [`${basePath}api`]: 'http://localhost:8080'
    }
  }
})
