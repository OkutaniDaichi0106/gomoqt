// Global Jest types for testing environment
// This file provides Jest globals when @types/jest is included but tests are excluded from main tsconfig

declare global {
  var describe: jest.Describe;
  var it: jest.It;
  var test: jest.It;
  var expect: jest.Expect;
  var beforeEach: jest.Lifecycle;
  var afterEach: jest.Lifecycle;
  var beforeAll: jest.Lifecycle;
  var afterAll: jest.Lifecycle;
  var jest: typeof jest;
}

export {};
