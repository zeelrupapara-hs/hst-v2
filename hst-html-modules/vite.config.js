import path from "node:path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  build: {
    target: "es2022",
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules/react-dom")) return "react-dom";
          if (id.includes("node_modules/react-router")) return "router";
          if (id.includes("node_modules/react/")) return "react";
          if (id.includes("/features/modules/definitions/")) return "modules";
          if (id.includes("/features/modules/columns/")) return "columns";
        },
      },
    },
  },
  server: {
    port: 5173,
    open: true,
  },
});
