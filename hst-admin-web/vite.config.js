import path from "node:path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { "@": path.resolve(import.meta.dirname, "src") },
  },
  server: {
    // 5173 and 5174 belong to the trader terminal and the old prototype
    port: 5175,
    strictPort: true,
  },
});
