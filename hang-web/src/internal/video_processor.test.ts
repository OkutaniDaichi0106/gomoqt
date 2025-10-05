import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { VideoTrackProcessor } from './video_processor';

describe('VideoTrackProcessor', () => {
    const originalSelf = globalThis.self;
    const originalDocument = globalThis.document;
    const originalPerformance = globalThis.performance;
    const originalRequestAnimationFrame = globalThis.requestAnimationFrame;
    const originalConsoleWarn = console.warn;

    const originalVideoFrame = globalThis.VideoFrame;
    const originalMediaStream = globalThis.MediaStream;
    const originalReadableStream = globalThis.ReadableStream;

    beforeEach(() => {
        (globalThis as any).self = {};
        (globalThis as any).document = {
            createElement: vi.fn(() => ({
                play: vi.fn(async () => undefined),
                srcObject: null,
                onloadedmetadata: null,
                videoWidth: 0,
                videoHeight: 0,
            })),
        };
        (globalThis as any).performance = {
            now: vi.fn(() => 0),
        };
        (globalThis as any).requestAnimationFrame = vi.fn();
        (globalThis as any).VideoFrame = vi.fn();
        (globalThis as any).MediaStream = vi.fn();
        console.warn = vi.fn();
    });

    afterEach(() => {
        if (originalSelf === undefined) {
            delete (globalThis as any).self;
        } else {
            (globalThis as any).self = originalSelf;
        }

        if (originalDocument === undefined) {
            delete (globalThis as any).document;
        } else {
            (globalThis as any).document = originalDocument;
        }

        if (originalPerformance === undefined) {
            delete (globalThis as any).performance;
        } else {
            (globalThis as any).performance = originalPerformance;
        }

        if (originalRequestAnimationFrame === undefined) {
            delete (globalThis as any).requestAnimationFrame;
        } else {
            (globalThis as any).requestAnimationFrame = originalRequestAnimationFrame;
        }

        if (originalVideoFrame === undefined) {
            delete (globalThis as any).VideoFrame;
        } else {
            (globalThis as any).VideoFrame = originalVideoFrame;
        }

        if (originalMediaStream === undefined) {
            delete (globalThis as any).MediaStream;
        } else {
            (globalThis as any).MediaStream = originalMediaStream;
        }

        if (originalReadableStream === undefined) {
            delete (globalThis as any).ReadableStream;
        } else {
            (globalThis as any).ReadableStream = originalReadableStream;
        }

        console.warn = originalConsoleWarn;
    });

    test('uses MediaStreamTrackProcessor when available', () => {
        const mockReadable = {};
        const mockProcessor = {
            readable: mockReadable,
        };
        (globalThis as any).self.MediaStreamTrackProcessor = vi.fn(() => mockProcessor);

        const track = {
            getSettings: vi.fn(() => ({ frameRate: 30 })),
        } as unknown as MediaStreamTrack;

        const processor = new VideoTrackProcessor(track);
        expect(processor.readable).toBe(mockReadable);
        expect((globalThis as any).self.MediaStreamTrackProcessor).toHaveBeenCalledWith({ track });
        expect(track.getSettings).not.toHaveBeenCalled();
        expect(console.warn).not.toHaveBeenCalled();
    });

    test('uses polyfill when MediaStreamTrackProcessor is unavailable', () => {
        const track = {
            getSettings: vi.fn(() => ({ frameRate: 30 })),
        } as unknown as MediaStreamTrack;

        const processor = new VideoTrackProcessor(track);
        expect(processor.readable).toBeInstanceOf(ReadableStream);
        expect(track.getSettings).toHaveBeenCalled();
        expect(console.warn).toHaveBeenCalledWith('Using MediaStreamTrackProcessor polyfill; performance might suffer.');

        // Test that start method creates video and canvas elements
        const reader = processor.readable.getReader();
        // Note: In a real test, we would need to mock the async start method
        // For now, just ensure the reader is created
        expect(reader).toBeDefined();
    });

    test('polyfill start method initializes video and canvas', async () => {
        // Temporarily remove MediaStreamTrackProcessor to force polyfill usage
        const originalProcessor = (self as any).MediaStreamTrackProcessor;
        delete (self as any).MediaStreamTrackProcessor;

        // Mock the start method's async operations
        let loadedMetadataResolve: () => void;
        const loadedMetadataPromise = new Promise<void>((resolve) => {
            loadedMetadataResolve = resolve;
        });

        const mockVideo = {
            play: vi.fn().mockResolvedValue(undefined),
            srcObject: null,
            set onloadedmetadata(callback: () => void) {
                // Simulate calling the callback immediately after setting
                Promise.resolve().then(() => {
                    callback();
                    loadedMetadataResolve();
                });
            },
            videoWidth: 640,
            videoHeight: 480,
        };
        const mockCanvas = {
            width: 0,
            height: 0,
            getContext: vi.fn(() => ({
                drawImage: vi.fn(),
            })),
        };

        (globalThis as any).document.createElement.mockImplementation((tagName: string) => {
            if (tagName === 'video') return mockVideo;
            if (tagName === 'canvas') return mockCanvas;
            return {};
        });

        let startPromise: Promise<void>;
        const OriginalReadableStream = (globalThis as any).ReadableStream;
        (globalThis as any).ReadableStream = class MockReadableStream {
            constructor(underlyingSource: any) {
                startPromise = underlyingSource.start();
            }
            getReader() {
                return {
                    read: vi.fn().mockResolvedValue({ done: true }),
                };
            }
        };

        const track = {
            getSettings: vi.fn(() => ({ frameRate: 30 })),
        } as unknown as MediaStreamTrack;

        const processor = new VideoTrackProcessor(track);

        // Wait for start to complete
        await startPromise!;

        expect(mockVideo.play).toHaveBeenCalled();
        expect(mockCanvas.width).toBe(640);
        expect(mockCanvas.height).toBe(480);
        expect(mockCanvas.getContext).toHaveBeenCalledWith('2d', { desynchronized: true });

        // Restore original MediaStreamTrackProcessor and ReadableStream
        if (originalProcessor) {
            (self as any).MediaStreamTrackProcessor = originalProcessor;
        }
        (globalThis as any).ReadableStream = OriginalReadableStream;
    }, 30000);

    test('throws when track settings are unavailable', () => {
        const track = {
            getSettings: vi.fn(() => undefined),
        } as unknown as MediaStreamTrack;

        expect(() => new VideoTrackProcessor(track)).toThrow('track has no settings');
        expect(track.getSettings).toHaveBeenCalled();
    });

    test('throws when canvas context creation fails', async () => {
        // Temporarily remove MediaStreamTrackProcessor to force polyfill usage
        const originalProcessor = (self as any).MediaStreamTrackProcessor;
        delete (self as any).MediaStreamTrackProcessor;

        const mockVideo = {
            play: vi.fn().mockResolvedValue(undefined),
            srcObject: null,
            set onloadedmetadata(callback: () => void) {
                Promise.resolve().then(callback);
            },
            videoWidth: 640,
            videoHeight: 480,
        };
        const mockCanvas = {
            width: 0,
            height: 0,
            getContext: vi.fn(() => null), // Context creation fails
        };

        (globalThis as any).document.createElement.mockImplementation((tagName: string) => {
            if (tagName === 'video') return mockVideo;
            if (tagName === 'canvas') return mockCanvas;
            return {};
        });

        let startPromise: Promise<void>;
        const OriginalReadableStream = (globalThis as any).ReadableStream;
        (globalThis as any).ReadableStream = class MockReadableStream {
            constructor(underlyingSource: any) {
                startPromise = underlyingSource.start();
            }
            getReader() {
                return {
                    read: vi.fn().mockResolvedValue({ done: true }),
                };
            }
        };

        const track = {
            getSettings: vi.fn(() => ({ frameRate: 30 })),
        } as unknown as MediaStreamTrack;

        const processor = new VideoTrackProcessor(track);

        await expect(startPromise!).rejects.toThrow('failed to create canvas context');

        // Restore
        if (originalProcessor) {
            (self as any).MediaStreamTrackProcessor = originalProcessor;
        }
        (globalThis as any).ReadableStream = OriginalReadableStream;
    });

    test('handles missing frameRate in settings', () => {
        // Temporarily remove MediaStreamTrackProcessor to force polyfill usage
        const originalProcessor = (self as any).MediaStreamTrackProcessor;
        delete (self as any).MediaStreamTrackProcessor;

        const track = {
            getSettings: vi.fn(() => ({ /* no frameRate */ })),
        } as unknown as MediaStreamTrack;

        const processor = new VideoTrackProcessor(track);
        expect(processor.readable).toBeInstanceOf(ReadableStream);
        expect(track.getSettings).toHaveBeenCalled();
        expect(console.warn).toHaveBeenCalledWith('Using MediaStreamTrackProcessor polyfill; performance might suffer.');

        // Restore
        if (originalProcessor) {
            (self as any).MediaStreamTrackProcessor = originalProcessor;
        }
    });

    test('polyfill handles video play failure', async () => {
        // Temporarily remove MediaStreamTrackProcessor to force polyfill usage
        const originalProcessor = (self as any).MediaStreamTrackProcessor;
        delete (self as any).MediaStreamTrackProcessor;

        const mockVideo = {
            play: vi.fn().mockRejectedValue(new Error('Video play failed')),
            srcObject: null,
            onloadedmetadata: null,
            videoWidth: 640,
            videoHeight: 480,
        };

        (globalThis as any).document.createElement.mockImplementation((tagName: string) => {
            if (tagName === 'video') return mockVideo;
            return {};
        });

        let startPromise: Promise<void>;
        const OriginalReadableStream = (globalThis as any).ReadableStream;
        (globalThis as any).ReadableStream = class MockReadableStream {
            constructor(underlyingSource: any) {
                startPromise = underlyingSource.start();
            }
            getReader() {
                return {
                    read: vi.fn().mockResolvedValue({ done: true }),
                };
            }
        };

        const track = {
            getSettings: vi.fn(() => ({ frameRate: 30 })),
        } as unknown as MediaStreamTrack;

        const processor = new VideoTrackProcessor(track);

        await expect(startPromise!).rejects.toThrow('Video play failed');

        // Restore
        if (originalProcessor) {
            (self as any).MediaStreamTrackProcessor = originalProcessor;
        }
        (globalThis as any).ReadableStream = OriginalReadableStream;
    });
});
