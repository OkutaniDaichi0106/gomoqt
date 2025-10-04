import { defineConfig } from 'vitest/config';
import { resolve } from 'path';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    server: {
      deps: {
        inline: ['golikejs'],
      },
    },
    include: ['src/**/*.test.ts'],
    poolOptions: {
      threads: {
        maxThreads: 4,
        minThreads: 2,
      },
    },
    testTimeout: 10000, // 10 seconds timeout for tests
    hookTimeout: 5000,  // 5 seconds timeout for hooks
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['src/**/*.ts'],
      exclude: [
        'node_modules/',
        'dist/',
        '**/*.test.ts',
        '**/*.config.ts',
      ],
    },
    // TypeScript support
    typecheck: {
      enabled: false, // Enable if you want type checking in tests
    },
  },
  resolve: {
    alias: {
      '@okutanidaichi/moqt': resolve(__dirname, '../moq-web/src'),
      '@okutanidaichi/moqt/io': resolve(__dirname, '../moq-web/src/io'),
      'golikejs': resolve(__dirname, 'node_modules/golikejs/dist'),
    },
    extensions: ['.mjs', '.js', '.mts', '.ts', '.jsx', '.tsx', '.json'],
    // Allow resolving modules without explicit .js extension
    mainFields: ['module', 'jsnext:main', 'jsnext', 'main'],
  },
});
