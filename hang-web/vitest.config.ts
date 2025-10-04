import { defineConfig } from 'vitest/config';
import type { Plugin } from 'vite';

// Workaround for ESM packages missing .js extensions in relative imports
const esmFixPlugin = (): Plugin => ({
  name: 'esm-fix',
  enforce: 'pre',
  resolveId(source, importer) {
    // Fix relative imports in golikejs and @okutanidaichi/moqt packages
    if (importer && (importer.includes('golikejs') || importer.includes('@okutanidaichi/moqt')) && source.startsWith('./')) {
      // Add .js extension if missing
      if (!source.endsWith('.js') && !source.endsWith('.ts') && !source.endsWith('.json')) {
        return this.resolve(source + '.js', importer, { skipSelf: true });
      }
    }
    return null;
  },
});

export default defineConfig({
  plugins: [esmFixPlugin()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    deps: {
      optimizer: {
        web: {
          include: ['@okutanidaichi/moqt', 'golikejs'],
        },
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
    extensions: ['.mjs', '.js', '.mts', '.ts', '.jsx', '.tsx', '.json'],
    // Allow resolving modules without explicit .js extension
    mainFields: ['module', 'jsnext:main', 'jsnext', 'main'],
  },
});
