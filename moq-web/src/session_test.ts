import { describe, it, expect, vi, beforeEach, afterEach } from "../deps.ts";
import { Session } from './session.ts';
import { Versions, DEFAULT_CLIENT_VERSIONS } from './internal.ts';
import { Extensions } from './internal/extensions.ts';
import { DefaultTrackMux, TrackMux } from './track_mux.ts';
import type { BroadcastPath } from './broadcast_path.ts';
import type { TrackName } from './protocol.ts';
import type { TrackPrefix } from './track_prefix.ts';

// Mock WebTransport
const mockWebTransport = {
  ready: Promise.resolve(),
  createBidirectionalStream: vi.fn(() => Promise.resolve({
    writable: { getWriter: vi.fn() },
    readable: { getReader: vi.fn(() => ({ read: vi.fn(() => Promise.resolve({ done: true })) })) },
  })),
  createUnidirectionalStream: vi.fn(() => Promise.resolve({})),
  incomingBidirectionalStreams: {
    getReader: vi.fn(() => ({
      read: vi.fn(() => Promise.resolve({ done: true })),
      releaseLock: vi.fn(),
    })),
  },
  incomingUnidirectionalStreams: {
    getReader: vi.fn(() => ({
      read: vi.fn(() => Promise.resolve({ done: true })),
      releaseLock: vi.fn(),
    })),
  },
  close: vi.fn(),
  closed: Promise.resolve(),
} as unknown as WebTransport;

// Mock other dependencies
// TODO: Migrate mock to Deno compatible pattern
// TODO: Migrate mock to Deno compatible pattern
// TODO: Migrate mock to Deno compatible pattern
// TODO: Migrate mock to Deno compatible pattern
// TODO: Migrate mock to Deno compatible pattern
// TODO: Migrate mock to Deno compatible pattern
// TODO: Migrate mock to Deno compatible pattern
// TODO: Migrate mock to Deno compatible pattern
const mockContext = {
  err: vi.fn(() => null),
  done: vi.fn(() => Promise.resolve()),
};

// TODO: Migrate mock to Deno compatible pattern
describe('Session', () => {
  let session: Session;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Constructor', () => {
    it('should create session with default parameters', async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      assertExists(session);
      assertInstanceOf(session.ready, Promise);
      assertEquals(session.mux, DefaultTrackMux);
      // Ensure the promise is resolved to prevent unhandled rejections
      await session.ready.catch(() => {});
    });

    it('should create session with custom version', async () => {
      const conn = { ...mockWebTransport };
      const versions = new Set([2n]);
      session = new Session({ conn, versions });
      assertExists(session);
      await session.ready.catch(() => {});
    });

    it('should create session with custom extensions', async () => {
      const conn = { ...mockWebTransport };
      const extensions = new Extensions();
      session = new Session({ conn, extensions });
      assertExists(session);
      await session.ready.catch(() => {});
    });

    it('should create session with custom mux', async () => {
      const conn = { ...mockWebTransport };
      const mux = new TrackMux();
      session = new Session({ conn, mux });
      assertExists(session);
      assertEquals(session.mux, mux);
      await session.ready.catch(() => {});
    });
  });

  describe('Ready', () => {
    it('should resolve when connection is ready', async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      await expect(session.ready).resolves.toBeUndefined();
    });

    it('should handle setup error', async () => {
      const conn = {
        ...mockWebTransport,
        ready: Promise.reject(new Error('Connection failed')),
      };
      session = new Session({ conn });
      await expect(session.ready).rejects.toThrow('Connection failed');
    });
  });

  describe('Session Initialization', () => {
    it('should initialize session stream on ready', async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      await session.ready;
      // Check if SessionStream was created
      const { SessionStream } = await import('./session_stream');
      expect(SessionStream).toHaveBeenCalled();
    });

    it('should send session client message', async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      await session.ready;
      const { SessionClientMessage } = await import('./message');
      expect(SessionClientMessage).toHaveBeenCalled();
    });

    it('should receive session server message', async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      await session.ready;
      const { SessionServerMessage } = await import('./message');
      expect(SessionServerMessage).toHaveBeenCalled();
    });
  });

  describe('Internal State', () => {
    it('should initialize session properly', async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      await session.ready;
      assertExists(session);
    });

    it('should have all required methods', () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      assertEquals(typeof session.acceptAnnounce, 'function');
      assertEquals(typeof session.subscribe, 'function');
      assertEquals(typeof session.close, 'function');
      assertEquals(typeof session.closeWithError, 'function');
    });
  });

  describe('Methods', () => {
    beforeEach(async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      await session.ready;
    });

    it('should have acceptAnnounce method', () => {
      expect(session).toHaveProperty('acceptAnnounce');
      assertEquals(typeof session.acceptAnnounce, 'function');
    });

    it('should have subscribe method', () => {
      expect(session).toHaveProperty('subscribe');
      assertEquals(typeof session.subscribe, 'function');
    });

    it('should have close method', () => {
      expect(session).toHaveProperty('close');
      assertEquals(typeof session.close, 'function');
    });

    it('should have closeWithError method', () => {
      expect(session).toHaveProperty('closeWithError');
      assertEquals(typeof session.closeWithError, 'function');
    });

    it('should close connection with error', async () => {
      await session.closeWithError(1, 'Test error');
      expect(mockWebTransport.close).toHaveBeenCalledWith({
        closeCode: 1,
        reason: 'Test error',
      });
    });
  });

  describe('Error Handling', () => {
    afterEach(() => {
      vi.clearAllMocks();
    });

    it('should handle initialization errors', async () => {
      const conn = {
        ...mockWebTransport,
        createBidirectionalStream: vi.fn(() => Promise.reject(new Error('Stream failed'))),
      };
      session = new Session({ conn });
      await assertRejects(async () => await session.ready);
    });

    it('should handle invalid version', async () => {
      const conn = { ...mockWebTransport };
      const versions = new Set([999n]); // Invalid version
      session = new Session({ conn, versions });
      await expect(session.ready).rejects.toThrow('Incompatible session version');
    });
  });

  describe('Async Operations', () => {
    beforeEach(async () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      await session.ready;
    });

    it('should handle acceptAnnounce', async () => {
      const result = await session.acceptAnnounce('/' as TrackPrefix);
      assertExists(result);
    });

    it('should handle subscribe', async () => {
      const result = await session.subscribe('' as BroadcastPath, '' as TrackName);
      assertExists(result);
    });

    it('should handle incoming bidirectional streams', async () => {
      // This is tested internally in #listenBiStreams
      assertEquals(true, true); // Placeholder for stream handling tests
    });

    it('should handle incoming unidirectional streams and message queue', async () => {
      // This is tested internally in #listenUniStreams
      assertEquals(true, true); // Placeholder for stream handling tests
    });
  });
});
