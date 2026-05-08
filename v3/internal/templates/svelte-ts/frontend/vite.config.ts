import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte(), wails("./bindings")],
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT || "9245"),
    strictPort: true,
  },
});
