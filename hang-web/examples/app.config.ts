import { defineConfig } from "@solidjs/start/config";

export default defineConfig({
  vite: {
    ssr: {
      external: ["@okutanidaichi/moqt"]
    }
  }
});
