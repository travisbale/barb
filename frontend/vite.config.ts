import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const apiPort = process.env.VITE_API_PORT || '8443'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: '../cmd/barb/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': {
        target: `https://localhost:${apiPort}`,
        secure: false,
      },
    },
  },
})
