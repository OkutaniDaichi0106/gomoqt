/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.cjs'],
  extensionsToTreatAsEsm: ['.ts'],
  slowTestThreshold: 5000,
  
  // パフォーマンス最適化設定
  maxWorkers: '50%', // CPUコア数の50%を使用
  cache: true,
  cacheDirectory: '<rootDir>/.jest-cache',
  
  // 並列実行の最適化
  testTimeout: 10000, // 10秒でタイムアウト
  
  // TypeScript変換の最適化 - 最新の推奨設定
  transform: {
    '^.+\\.ts$': ['ts-jest', {
      useESM: true,
      tsconfig: 'tsconfig.test.json',
      // 型チェックを無効化して速度向上
      diagnostics: false,
    }],
  },
  
  moduleNameMapper: {
    '^(\\.{1,2}/.*)\\.js$': '$1',
  },
  
  // 不要なファイルをスキップ
  testPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/coverage/',
  ],
  
  // 監視モードの最適化
  watchPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/coverage/',
    '/.jest-cache/',
  ],
};
