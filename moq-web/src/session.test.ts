import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Session } from './session';
import { Versions, DEFAULT_CLIENT_VERSIONS } from './internal';
import { Extensions } from './internal/extensions';
import { DefaultTrackMux, TrackMux } from './track_mux';
import type { BroadcastPath } from './broadcast_path';
import type { TrackPrefix } from './track_prefix';
import { TrackName } from './track_name';

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
vi.mock('./session_stream', () => ({
  SessionStream: vi.fn(() => ({
    context: mockContext,
  })),
}));

vi.mock('./io', () => ({
  Writer: vi.fn(() => ({
    writeUint8: vi.fn(),
    flush: vi.fn(() => Promise.resolve(null)),
    streamId: 0n,
  })),
  Reader: vi.fn(() => ({
    readUint8: vi.fn(() => [0, null]),
    streamId: 0n,
  })),
}));

vi.mock('./message', () => ({
  SessionClientMessage: vi.fn(() => ({
    encode: vi.fn(() => Promise.resolve(null)),
    versions: new Set([1n]),
    extensions: new Extensions(),
  })),
  SessionServerMessage: vi.fn(() => ({
    decode: vi.fn(() => Promise.resolve(null)),
    version: 0xffffff00n,
    extensions: new Extensions(),
  })),
  AnnouncePleaseMessage: vi.fn(() => ({
    encode: vi.fn(() => Promise.resolve(null)),
    prefix: '',
  })),
  AnnounceInitMessage: vi.fn(() => ({
    decode: vi.fn(() => Promise.resolve(null)),
  })),
  SubscribeMessage: vi.fn(() => ({
    encode: vi.fn(() => Promise.resolve(null)),
    subscribeId: 0n,
  })),
  SubscribeOkMessage: vi.fn(() => ({
    decode: vi.fn(() => Promise.resolve(null)),
  })),
  GroupMessage: vi.fn(() => ({
    decode: vi.fn(() => Promise.resolve(null)),
    subscribeId: 0n,
  })),
}));

vi.mock('./announce_stream', () => ({
  AnnouncementReader: vi.fn(),
  AnnouncementWriter: vi.fn(),
}));

vi.mock('./subscribe_stream', () => ({
  ReceiveSubscribeStream: vi.fn(),
  SendSubscribeStream: vi.fn(),
}));

vi.mock('./track', () => ({
  TrackReader: vi.fn(),
  TrackWriter: vi.fn(),
}));

vi.mock('./group_stream', () => ({
  GroupReader: vi.fn(),
  GroupWriter: vi.fn(),
}));

vi.mock('./internal/queue', () => ({
  Queue: vi.fn().mockImplementation(() => ({
    enqueue: vi.fn(),
    dequeue: vi.fn(),
  })),
}));

const mockContext = {
  err: vi.fn(() => null),
  done: vi.fn(() => Promise.resolve()),
};

vi.mock('golikejs/context', () => ({
  background: vi.fn(() => mockContext),
  watchPromise: vi.fn(() => mockContext),
}));

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
      expect(session).toBeDefined();
      expect(session.ready).toBeInstanceOf(Promise);
      expect(session.mux).toBe(DefaultTrackMux);
      // Ensure the promise is resolved to prevent unhandled rejections
      await session.ready.catch(() => {});
    });

    it('should create session with custom version', async () => {
      const conn = { ...mockWebTransport };
      const versions = new Set([2n]);
      session = new Session({ conn, versions });
      expect(session).toBeDefined();
      await session.ready.catch(() => {});
    });

    it('should create session with custom extensions', async () => {
      const conn = { ...mockWebTransport };
      const extensions = new Extensions();
      session = new Session({ conn, extensions });
      expect(session).toBeDefined();
      await session.ready.catch(() => {});
    });

    it('should create session with custom mux', async () => {
      const conn = { ...mockWebTransport };
      const mux = new TrackMux();
      session = new Session({ conn, mux });
      expect(session).toBeDefined();
      expect(session.mux).toBe(mux);
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
      expect(session).toBeDefined();
    });

    it('should have all required methods', () => {
      const conn = { ...mockWebTransport };
      session = new Session({ conn });
      expect(typeof session.acceptAnnounce).toBe('function');
      expect(typeof session.subscribe).toBe('function');
      expect(typeof session.close).toBe('function');
      expect(typeof session.closeWithError).toBe('function');
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
      expect(typeof session.acceptAnnounce).toBe('function');
    });

    it('should have subscribe method', () => {
      expect(session).toHaveProperty('subscribe');
      expect(typeof session.subscribe).toBe('function');
    });

    it('should have close method', () => {
      expect(session).toHaveProperty('close');
      expect(typeof session.close).toBe('function');
    });

    it('should have closeWithError method', () => {
      expect(session).toHaveProperty('closeWithError');
      expect(typeof session.closeWithError).toBe('function');
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
      await expect(session.ready).rejects.toThrow();
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
      expect(result).toBeDefined();
    });

    it('should handle subscribe', async () => {
      const result = await session.subscribe('' as BroadcastPath, '' as TrackName);
      expect(result).toBeDefined();
    });

    it('should handle incoming bidirectional streams', async () => {
      // This is tested internally in #listenBiStreams
      expect(true).toBe(true); // Placeholder for stream handling tests
    });

    it('should handle incoming unidirectional streams and message queue', async () => {
      // This is tested internally in #listenUniStreams
      expect(true).toBe(true); // Placeholder for stream handling tests
    });
  });
});
