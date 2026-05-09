import { defineConfig } from "vite";

export default defineConfig({
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT, 10) || 9245,
    strictPort: true,
  },
});
