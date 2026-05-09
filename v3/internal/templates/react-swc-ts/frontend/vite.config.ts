import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), wails("./bindings")],
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT, 10) || 9245,
    strictPort: true,
  },
});
