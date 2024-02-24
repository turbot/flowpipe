import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";
import { defineConfig } from "vite";

// https://vitejs.dev/config/
export default defineConfig({
  base: "/form",
  server: {
    proxy: {
      "/api": "http://localhost:7103",
    },
  },
  plugins: [svgr(), react()],
  resolve: {
    alias: {
      api: "/src/api",
      assets: "/src/assets",
      components: "/src/components",
      hooks: "/src/hooks",
      lib: "/src/lib",
      src: "/src",
      styles: "/src/styles",
      types: "/src/types",
      utils: "/src/utils",
    },
  },
});
