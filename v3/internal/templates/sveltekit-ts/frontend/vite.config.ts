import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig, searchForWorkspaceRoot } from "vite";
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT, 10) || 9245,
    strictPort: true,
    fs: {
      allow: [
        // search up for workspace root
        searchForWorkspaceRoot(process.cwd()),
        // your custom rules
        "./bindings/*",
      ],
    },
  },
  plugins: [sveltekit(), wails("./bindings")],
});
