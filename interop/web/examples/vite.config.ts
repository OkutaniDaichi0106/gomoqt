import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';

export default defineConfig({
  plugins: [solidPlugin()],
  server: {
    port: 9000,
    host: 'moqt.example.com',
    https: {
      key: '../../server/moqt.example.com-key.pem',
      cert: '../../server/moqt.example.com.pem',
    },
  },
  build: {
    target: 'esnext',
  },
});
