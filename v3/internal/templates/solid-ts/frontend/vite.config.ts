import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
  plugins: [solid(), wails("./bindings")],
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT || "9245"),
    strictPort: true,
  },
});
