import path from "path";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

const devServer = "http://localhost:8787";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss() as any],
  server: {
    proxy: {
      "/ddns": devServer,
      "/myip": devServer,
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
