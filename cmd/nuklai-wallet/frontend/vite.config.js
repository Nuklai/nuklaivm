// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Assuming '/api' is the base path for your backend API
      // Adjust '/api' if your backend API uses a different base path
      '/api': {
        target: 'http://localhost:34115', // Your Go backend server address
        changeOrigin: true,
        secure: false
        // If your backend API does not use '/api' as the base path, rewrite the path:
        // rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
