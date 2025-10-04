import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  test: {
    environment: 'node',
    setupFiles: ['./vitest.setup.ts'],
    globals: true,
    server: {
      deps: {
        inline: ['golikejs'],
      },
    },
    
    // Performance optimization settings
    isolate: true,
    pool: 'threads',
    poolOptions: {
      threads: {
        singleThread: true, // Run tests serially for stability
      },
    },
    
    // Test timeout
    testTimeout: 10000,
    
    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'dist/',
        'coverage/',
        '**/*.test.ts',
        '**/*.spec.ts',
      ],
    },
    
    // File patterns
    include: ['src/**/*.{test,spec}.ts'],
    exclude: [
      'node_modules',
      'dist',
      'coverage',
      '.vitest-cache',
      '.git',
    ],
  },
  
  resolve: {
    conditions: ['node', 'import', 'module', 'browser', 'default'],
    extensions: ['.mjs', '.js', '.mts', '.ts', '.jsx', '.tsx', '.json'],
  },
});
