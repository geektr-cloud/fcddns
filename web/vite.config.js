import { defineConfig } from 'vite'

const devServer = "http://localhost:8787"

export default defineConfig({
  server: {
    proxy: {
      "/ddns": devServer,
      "/myip": devServer,
    }
  }
})
