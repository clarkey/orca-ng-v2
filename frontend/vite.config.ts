import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: false, // Important: don't change origin to preserve cookies
        secure: false,
        configure: (proxy, options) => {
          proxy.on('proxyReq', (proxyReq, req, res) => {
            // Forward cookies from the client to the backend
            const cookies = req.headers.cookie;
            if (cookies) {
              proxyReq.setHeader('Cookie', cookies);
            }
          });
          proxy.on('proxyRes', (proxyRes, req, res) => {
            // Ensure set-cookie headers are properly forwarded
            const setCookieHeaders = proxyRes.headers['set-cookie'];
            if (setCookieHeaders) {
              res.setHeader('Set-Cookie', setCookieHeaders);
            }
          });
        },
      },
    },
  },
  build: {
    outDir: '../backend/cmd/orca/dist',
    emptyOutDir: true,
  },
})