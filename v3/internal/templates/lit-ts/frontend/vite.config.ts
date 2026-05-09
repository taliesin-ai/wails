import { defineConfig } from "vite";
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [wails("./bindings")],
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT, 10) || 9245,
    strictPort: true,
  },
});
