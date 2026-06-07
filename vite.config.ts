import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// ChampIQ web — Vite + React. Port 5174 (the Fieldstone admin uses 5173).
export default defineConfig({
  plugins: [react()],
  server: { port: 5174, host: true },
  preview: { port: 5174 },
})
