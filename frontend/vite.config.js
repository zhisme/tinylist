import { defineConfig } from 'vite'
import preact from '@preact/preset-vite'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [preact(), tailwindcss()],
  // Base path for assets - frontend is always served at /tinylist/
  base: '/tinylist/',
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
