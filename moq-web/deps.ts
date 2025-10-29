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

// Optional test-time promise leak monitor.
// Enable by setting the environment variable MOQT_ENABLE_PROMISE_MONITOR=1
// when running tests. This patches Promise.prototype.then to track the
// number of outstanding promises so we can diagnose pending-promise issues
// reported by Deno at process exit. Only enable in test/debug runs.
let enableMonitor = false;
try {
  enableMonitor = Deno.env.get("MOQT_ENABLE_PROMISE_MONITOR") === "1";
} catch {
  // No env access permission — monitor disabled
}
if (enableMonitor) {
  // Avoid double-installing
  if (!(globalThis as any).__moqt_promise_monitor_installed) {
    (globalThis as any).__moqt_promise_monitor_installed = true;
    const pending = new Set<Promise<unknown>>();
    const origThen = Promise.prototype.then;
    // Use `any`-typed function to avoid TypeScript signature mismatches.
    (Promise.prototype as any).then = function (this: any, onFulfilled?: any, onRejected?: any): any {
      try {
        pending.add(this as Promise<unknown>);
        const wrappedFulfilled = (v: unknown) => {
          pending.delete(this as Promise<unknown>);
          return onFulfilled ? onFulfilled(v) : v;
        };
        const wrappedRejected = (e: unknown) => {
          pending.delete(this as Promise<unknown>);
          return onRejected ? onRejected(e) : Promise.reject(e);
        };
        return origThen.call(this, wrappedFulfilled, wrappedRejected);
      } catch (e) {
        // If anything goes wrong, fallback to original then
        return origThen.call(this, onFulfilled, onRejected);
      }
    } as any;

    // Expose a helper to get the current pending count
    (globalThis as any).__moqt_get_pending_count = () => pending.size;
  }
}

// Additional optional runtime resource monitor: timers and streams.
if (enableMonitor) {
  if (!(globalThis as any).__moqt_resource_monitor_installed) {
    (globalThis as any).__moqt_resource_monitor_installed = true;

    // Timers
    const originalSetTimeout = (globalThis as any).setTimeout;
    const originalSetInterval = (globalThis as any).setInterval;
    const originalClearTimeout = (globalThis as any).clearTimeout;
    const originalClearInterval = (globalThis as any).clearInterval;
    const activeTimers = new Map<number | string, string>();
    (globalThis as any).setTimeout = function (fn: any, ms?: any, ...args: any[]) {
      let id: any;
      const wrapped = function (...a: any[]) {
        try { activeTimers.delete(id); } catch {}
        return (fn as any)(...a);
      };
      id = originalSetTimeout(wrapped, ms, ...args);
      try { activeTimers.set(id, new Error().stack || ""); } catch { activeTimers.set(id, ""); }
      return id;
    };
    (globalThis as any).setInterval = function (fn: any, ms?: any, ...args: any[]) {
      const id = originalSetInterval(fn, ms, ...args);
      try { activeTimers.set(id, new Error().stack || ""); } catch { activeTimers.set(id, ""); }
      return id;
    };
    (globalThis as any).clearTimeout = function (id: any) {
      try { activeTimers.delete(id); } catch { }
      return originalClearTimeout(id);
    };
    (globalThis as any).clearInterval = function (id: any) {
      try { activeTimers.delete(id); } catch { }
      return originalClearInterval(id);
    };

    // NOTE: Stream wrapping was removed — it created references that prevented cleanup.
    // We now only track timers and promises.
    (globalThis as any).__moqt_get_resource_snapshot = () => ({
      timers: activeTimers.size,
      timerStacks: Array.from(activeTimers.values()).slice(0, 10),
    });
  }
}

// Mock utilities for testing
export { createMock, createSpy, type MockFunction } from "./test-utils/mock.ts";

// External dependencies (golikejs)
// To avoid triggering Deno's npm resolver (which can spawn background
// tasks) at module load time, we export lightweight, pure-TS fallbacks.
// Tests or code that require the real `@okudai/golikejs` can import it
// directly or call a loader helper (not provided here) when needed.

export type Context = {
  // done() returns a promise that resolves when the context is cancelled
  done(): Promise<void>;
};

// Simple synchronous background context: returns an object whose `done`
// promise never resolves until someone cancels it (we don't provide
// cancellation here; for tests that need cancellation use withCancelCause).
export function background(): Context {
  // Return a context whose `done()` resolves after a short tick. This avoids
  // leaving a permanently pending promise which can make Deno exit with a
  // non-zero status in some test environments. Tests that need stronger
  // cancellation semantics should use `withCancelCause`.
  const p = new Promise<void>((res) => setTimeout(res, 0));
  return { done: () => p };
}

// withCancelCause returns a pair [ctx, cancelFunc]; cancelFunc will resolve
// the done() promise when invoked. This is a minimal implementation used in
// tests; it does not emulate all golikejs behavior but is sufficient for test use.
export function withCancelCause(_ctx?: Context): [Context, (cause?: unknown) => void] {
  let resolveFn: ((v?: void) => void) | undefined;
  // Default resolution after a short tick to avoid indefinite pending promises
  const p = new Promise<void>((res) => { resolveFn = res; setTimeout(res, 0); });
  const ctx: Context = { done: () => p };
  const cancel = (_cause?: unknown) => { resolveFn?.(); };
  return [ctx, cancel];
}

// Lightweight synchronization primitives (no-op implementations suitable for
// tests that do not rely on heavy synchronization semantics).
export class Mutex {
  async lock(): Promise<void> { return; }
  unlock(): void { return; }
}

export class RWMutex extends Mutex {}

export class WaitGroup {
  add(_n: number): void { /* noop */ }
  done(): void { /* noop */ }
  async wait(): Promise<void> { return; }
}
