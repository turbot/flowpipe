import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";
import { defineConfig } from "vite";

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    proxy: {
      "/api": "http://localhost:7103",
    },
  },
  plugins: [svgr(), react()],
  resolve: {
    alias: {
      assets: "/src/assets",
      components: "/src/components",
      hooks: "/src/hooks",
      lib: "/src/lib",
      src: "/src",
      utils: "/src/utils",
    },
  },
});
