import { defineConfig } from "vite";
import { qwikVite } from "@builder.io/qwik/optimizer";
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    qwikVite({
      csr: true,
    }),
    wails("./bindings"),
  ],
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT || "9245"),
    strictPort: true,
  },
});
