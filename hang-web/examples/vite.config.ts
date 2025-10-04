import { defineConfig } from 'vite';
import basicSsl from "@vitejs/plugin-basic-ssl";

export default defineConfig({
  plugins: [],
  server: {
    fs: {
      // Allow serving files from parent directory
      allow: ['..']
    }
  },
});
