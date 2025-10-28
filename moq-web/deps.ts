// Centralized dependency management for the MoQT library
// This file exports all external dependencies used across the project

// Testing utilities from Deno standard library
export {
  assertEquals,
  assertExists,
  assertRejects,
  assertThrows,
  assertInstanceOf,
  assertStrictEquals,
  assertNotEquals,
  assertNotStrictEquals,
  assertArrayIncludes,
  fail,
  assertFalse,
} from "https://deno.land/std@0.224.0/assert/mod.ts";

export {
  describe,
  it,
  beforeEach,
  afterEach,
  beforeAll,
  afterAll,
} from "https://deno.land/std@0.224.0/testing/bdd.ts";

// Mock utilities for testing
export { createMock, createSpy, type MockFunction } from "./test-utils/mock.ts";

// External dependencies
export { background, withCancelCause, type Context } from "npm:golikejs@0.4.0/context";
export { Mutex, RWMutex, WaitGroup } from "npm:golikejs@0.4.0/sync";
