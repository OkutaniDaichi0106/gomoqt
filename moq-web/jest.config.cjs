/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node', // Changed from jsdom to node for better stability
  setupFilesAfterEnv: ['<rootDir>/jest.setup.cjs'],
  extensionsToTreatAsEsm: ['.ts'],
  slowTestThreshold: 5000,
  
  // Suppress log output for clean test results
  silent: false, // Keep false to enable log control in setup.cjs
  verbose: false, // Disable verbose test output
  
  // Performance optimization settings
  maxWorkers: 1, // Use single worker for stability
  // runInBand: true, // Replaced by maxWorkers: 1 to avoid deprecation warning
  cache: true,
  cacheDirectory: '<rootDir>/.jest-cache',
  
  // Test execution optimization
  testTimeout: 10000, // 10 second timeout
  
  // TypeScript transformation optimization - latest recommended settings
  transform: {
    '^.+\\.ts$': ['ts-jest', {
      useESM: true,
      tsconfig: 'tsconfig.test.json',
      // Disable type checking for faster execution
      diagnostics: false,
    }],
  },
  
  moduleNameMapper: {
    '^(\\.{1,2}/.*)\\.js$': '$1',
  },
  
  // Skip unnecessary files
  testPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/coverage/',
    '/.git/',
  ],
  
  // Watch mode optimization
  watchPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/coverage/',
    '/.jest-cache/',
    '/.git/',
  ],
  
  // Collect coverage from source files only
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.test.{ts,tsx}',
    '!src/**/index.ts', // Usually just exports
  ],
  
  // Error reporting configuration
  errorOnDeprecated: true,
  bail: false, // Continue running tests after failures
};
