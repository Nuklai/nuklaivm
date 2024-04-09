// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/nuklaiapi': {
        target: 'http://127.0.0.1:38071', // Your Go backend server address
        changeOrigin: true,
        secure: false
      }
    }
  }
})
