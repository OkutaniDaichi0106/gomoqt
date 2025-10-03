import { defineConfig } from 'vitest/config';
import { resolve } from 'path';
import type { Plugin } from 'vite';

// Workaround for golikejs missing .js extensions in ESM imports
const golikejsFixPlugin = (): Plugin => ({
  name: 'golikejs-fix',
  enforce: 'pre',
  resolveId(source, importer) {
    if (importer?.includes('golikejs') && source.startsWith('./')) {
      // Add .js extension if missing
      if (!source.endsWith('.js') && !source.endsWith('.ts') && !source.endsWith('.json')) {
        return this.resolve(source + '.js', importer, { skipSelf: true });
      }
    }
    return null;
  },
});

export default defineConfig({
  plugins: [golikejsFixPlugin()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    deps: {
      inline: ['golikejs'],
    },
    include: ['src/**/*.test.ts'],
    poolOptions: {
      threads: {
        maxThreads: 4,
        minThreads: 2,
      },
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
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
    },
    extensions: ['.mjs', '.js', '.mts', '.ts', '.jsx', '.tsx', '.json'],
    // Allow resolving modules without explicit .js extension
    mainFields: ['module', 'jsnext:main', 'jsnext', 'main'],
  },
});
