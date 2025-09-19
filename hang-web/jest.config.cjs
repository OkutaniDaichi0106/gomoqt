/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  extensionsToTreatAsEsm: ['.ts'],
  slowTestThreshold: 5000,
  
  // Performance optimization settings
  maxWorkers: '50%', // Use 50% of CPU cores
  cache: true,
  cacheDirectory: '<rootDir>/.jest-cache',
  
  // Parallel execution optimization
  testTimeout: 10000, // 10 seconds timeout
  
  // TypeScript transformation optimization - latest recommended settings
  transform: {
    '^.+\\.ts$': ['ts-jest', {
      useESM: true,
      tsconfig: 'tsconfig.test.json',
      // Disable type checking for better performance
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
  ],
  
  // Watch mode optimization
  watchPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/coverage/',
    '/.jest-cache/',
  ],
};
