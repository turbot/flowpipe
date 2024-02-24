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
      "@flowpipe": "/src",
    },
  },
});
