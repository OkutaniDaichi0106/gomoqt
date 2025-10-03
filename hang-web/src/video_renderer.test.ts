import { describe, test, expect, beforeEach, afterEach, vi, beforeAll, afterAll } from 'vitest';
import { VideoRenderer, VideoRendererInit } from "./video_renderer";
import { VideoTrackDecoder } from "./internal";

// Mock external dependencies
vi.mock("./internal", () => ({
    VideoTrackDecoder: vi.fn()
}));

// Test constants
const DEFAULT_CANVAS_WIDTH = 320;
const DEFAULT_CANVAS_HEIGHT = 240;
const DEFAULT_INTERSECTION_THRESHOLD = 0.01;
const ASYNC_FRAME_DELAY = 5; // Reduced delay for animation frame processing
const RAPID_FRAME_DELAY = 10; // Reduced delay for multiple frame processing
const MAX_ANIMATION_FRAME_ID = 1000; // Upper bound for random animation frame IDs

// Mock canvas context
const mockCanvasContext = {
    clearRect: vi.fn(),
    drawImage: vi.fn()
} as any;

// Mock canvas element
const mockCanvas = {
    width: DEFAULT_CANVAS_WIDTH,
    height: DEFAULT_CANVAS_HEIGHT,
    getContext: vi.fn(() => mockCanvasContext)
} as any;

// Mock IntersectionObserver with proper typing
const mockIntersectionObserver = {
    observe: vi.fn(),
    disconnect: vi.fn(),
    unobserve: vi.fn(),
    takeRecords: vi.fn(() => []),
    root: null,
    rootMargin: '',
    thresholds: [],
    callback: undefined as IntersectionObserverCallback | undefined,
    options: undefined as IntersectionObserverInit | undefined
} as any;

let createElementSpy: any;
let intersectionObserverConstructor: any;

const originalIntersectionObserver = global.IntersectionObserver;
const originalRequestAnimationFrame = global.requestAnimationFrame;
const originalCancelAnimationFrame = global.cancelAnimationFrame;
const nativeCreateElement = document.createElement.bind(document);

// Mock console methods
const originalConsoleLog = console.log;
const originalConsoleWarn = console.warn;

beforeAll(() => {
    createElementSpy = vi.spyOn(document, 'createElement');
});

afterAll(() => {
    createElementSpy.mockRestore();

    if (originalIntersectionObserver) {
        global.IntersectionObserver = originalIntersectionObserver;
    } else {
        delete (global as any).IntersectionObserver;
    }

    if (originalRequestAnimationFrame) {
        global.requestAnimationFrame = originalRequestAnimationFrame;
    } else {
        delete (global as any).requestAnimationFrame;
    }

    if (originalCancelAnimationFrame) {
        global.cancelAnimationFrame = originalCancelAnimationFrame;
    } else {
        delete (global as any).cancelAnimationFrame;
    }
});

describe("VideoRenderer", () => {
    beforeEach(() => {
        // Reset all mocks to ensure test isolation
        vi.clearAllMocks();

        // Reset canvas mock to default state
        mockCanvas.width = DEFAULT_CANVAS_WIDTH;
        mockCanvas.height = DEFAULT_CANVAS_HEIGHT;
        mockCanvas.getContext.mockReset();
        mockCanvas.getContext.mockImplementation(() => mockCanvasContext);

        // Reset canvas context methods
        mockCanvasContext.clearRect.mockReset();
        mockCanvasContext.drawImage.mockReset();

        // Setup document.createElement spy to return mock canvas
        createElementSpy.mockReset();
        createElementSpy.mockImplementation((tagName: string, options?: ElementCreationOptions) => {
            if (tagName === 'canvas') {
                return mockCanvas;
            }
            return nativeCreateElement(tagName, options);
        });

        // Reset IntersectionObserver mock
        mockIntersectionObserver.observe.mockReset();
        mockIntersectionObserver.disconnect.mockReset();
        mockIntersectionObserver.takeRecords.mockReset();
        mockIntersectionObserver.callback = undefined;
        mockIntersectionObserver.options = undefined;

        // Create fresh IntersectionObserver constructor mock
        intersectionObserverConstructor = vi.fn((callback, options) => {
            mockIntersectionObserver.callback = callback as IntersectionObserverCallback;
            mockIntersectionObserver.options = options as IntersectionObserverInit | undefined;
            return mockIntersectionObserver;
        });
        global.IntersectionObserver = intersectionObserverConstructor as unknown as typeof IntersectionObserver;

        // Mock requestAnimationFrame to execute callbacks asynchronously
        global.requestAnimationFrame = vi.fn((callback: FrameRequestCallback) => {
            // Generate random ID to simulate browser behavior
            const id = Math.floor(Math.random() * MAX_ANIMATION_FRAME_ID);
            // Execute callback asynchronously
            setTimeout(() => callback(performance.now()), 0);
            return id;
        });

        global.cancelAnimationFrame = vi.fn();

        // Spy on console methods to verify logging behavior
        console.log = vi.fn();
        console.warn = vi.fn();
    });

    afterEach(() => {
        console.log = originalConsoleLog;
        console.warn = originalConsoleWarn;

        createElementSpy.mockReset();

        if (originalIntersectionObserver) {
            global.IntersectionObserver = originalIntersectionObserver;
        } else {
            delete (global as any).IntersectionObserver;
        }

        if (originalRequestAnimationFrame) {
            global.requestAnimationFrame = originalRequestAnimationFrame;
        } else {
            delete (global as any).requestAnimationFrame;
        }

        if (originalCancelAnimationFrame) {
            global.cancelAnimationFrame = originalCancelAnimationFrame;
        } else {
            delete (global as any).cancelAnimationFrame;
        }
    });

    describe("Constructor", () => {
        test("creates video renderer with default options", () => {
            const renderer = new VideoRenderer();

            expect(createElementSpy).toHaveBeenCalledWith('canvas');
            expect(mockCanvas.width).toBe(320);
            expect(mockCanvas.height).toBe(240);
            expect(mockCanvas.getContext).toHaveBeenCalledWith('2d');
            expect(renderer.canvas).toBe(mockCanvas);
            expect(renderer.context).toBe(mockCanvasContext);
            expect(renderer.isVisible).toBe(true);
        });

        test("creates video renderer with custom dimensions", () => {
            const init: VideoRendererInit = {
                width: 1920,
                height: 1080
            };

            const renderer = new VideoRenderer(init);

            expect(mockCanvas.width).toBe(1920);
            expect(mockCanvas.height).toBe(1080);
            expect(renderer.canvas).toBe(mockCanvas);
        });

        test("sets up intersection observer by default", () => {
            const renderer = new VideoRenderer();

            expect(intersectionObserverConstructor).toHaveBeenCalledWith(
                expect.any(Function),
                { threshold: 0.01 }
            );
            expect(mockIntersectionObserver.observe).toHaveBeenCalledWith(mockCanvas);
            expect(renderer.observer).toBeDefined();
        });

        test("uses custom intersection threshold", () => {
            const init: VideoRendererInit = {
                intersectionThreshold: 0.5
            };

            const renderer = new VideoRenderer(init);

            expect(intersectionObserverConstructor).toHaveBeenCalledWith(
                expect.any(Function),
                { threshold: 0.5 }
            );
        });

        test("skips intersection observer when background rendering enabled", () => {
            const init: VideoRendererInit = {
                backgroundRendering: true
            };

            const renderer = new VideoRenderer(init);

            expect(intersectionObserverConstructor).not.toHaveBeenCalled();
            expect(renderer.observer).toBeUndefined();
            expect(renderer.isVisible).toBe(true);
        });

        test("creates renderer with all custom options", () => {
            const init: VideoRendererInit = {
                width: 640,
                height: 480,
                intersectionThreshold: 0.25,
                backgroundRendering: false
            };

            const renderer = new VideoRenderer(init);

            expect(mockCanvas.width).toBe(640);
            expect(mockCanvas.height).toBe(480);
            expect(intersectionObserverConstructor).toHaveBeenCalledWith(
                expect.any(Function),
                { threshold: 0.25 }
            );
            expect(renderer.isVisible).toBe(true);
        });
    });

    describe("Intersection Observer", () => {
        test("updates visibility when intersection changes", () => {
            const renderer = new VideoRenderer();
            const callback = mockIntersectionObserver.callback as IntersectionObserverCallback;

            // Simulate intersection change - visible
            const visibleEntry = { isIntersecting: true } as unknown as IntersectionObserverEntry;
            callback([visibleEntry], mockIntersectionObserver);

            expect(renderer.isVisible).toBe(true);

            // Simulate intersection change - not visible
            const hiddenEntry = { isIntersecting: false } as unknown as IntersectionObserverEntry;
            callback([hiddenEntry], mockIntersectionObserver);

            expect(renderer.isVisible).toBe(false);
        });

        test("handles empty intersection entries", () => {
            const renderer = new VideoRenderer();
            const callback = mockIntersectionObserver.callback as IntersectionObserverCallback;

            // Simulate empty entries
            callback([] as unknown as IntersectionObserverEntry[], mockIntersectionObserver);

            // Should not crash and visibility should remain unchanged
            expect(renderer.isVisible).toBe(true);
        });

        test("handles undefined entry", () => {
            const renderer = new VideoRenderer();
            const callback = mockIntersectionObserver.callback as IntersectionObserverCallback;

            // Simulate undefined entry
            callback([undefined as unknown as IntersectionObserverEntry], mockIntersectionObserver);

            // Should not crash
            expect(renderer.isVisible).toBe(true);
        });
    });

    describe("decoder()", () => {
        test("creates and returns decoder on first call", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer();
            const decoder = await renderer.decoder();

            expect(decoder).toBe(mockDecoderInstance);
            expect(VideoTrackDecoder).toHaveBeenCalledWith({
                destination: expect.any(WritableStream)
            });
        });

        test("returns cached decoder on subsequent calls", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer();
            
            const decoder1 = await renderer.decoder();
            const decoder2 = await renderer.decoder();

            expect(decoder1).toBe(decoder2);
            expect(VideoTrackDecoder).toHaveBeenCalledTimes(1);
        });

        test("writable stream processes video frames", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer();
            await renderer.decoder();

            // Get the WritableStream from the decoder constructor
            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            // Mock VideoFrame
            const mockVideoFrame = {
                close: vi.fn()
            } as unknown as VideoFrame;

            // Write frame to the stream
            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            // Should schedule rendering
            expect(global.requestAnimationFrame).toHaveBeenCalled();
        });
    });

    describe("Frame Rendering", () => {
        beforeEach(() => {
            // Reset requestAnimationFrame mock to control execution
            (global.requestAnimationFrame as any).mockImplementation((callback) => {
                const id = Math.floor(Math.random() * MAX_ANIMATION_FRAME_ID);
                callback(); // Execute immediately for testing
                return id;
            });
        });

        test("renders video frame when visible", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer();
            renderer.isVisible = true;
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = {
                close: vi.fn()
            } as unknown as VideoFrame;

            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            // Wait for animation frame to process
            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            expect(mockCanvasContext.clearRect).toHaveBeenCalledWith(0, 0, 320, 240);
            expect(mockCanvasContext.drawImage).toHaveBeenCalledWith(mockVideoFrame, 0, 0, 320, 240);
            expect(mockVideoFrame.close).toHaveBeenCalledTimes(1);
        });

        test("skips rendering when not visible", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer();
            renderer.isVisible = false;
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = {
                close: vi.fn()
            } as unknown as VideoFrame;

            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            expect(mockCanvasContext.clearRect).not.toHaveBeenCalled();
            expect(mockCanvasContext.drawImage).not.toHaveBeenCalled();
            expect(mockVideoFrame.close).toHaveBeenCalledTimes(1); // Still closes frame
        });

        test("handles custom delay function", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const delayFunc = vi.fn(() => Promise.resolve());
            const renderer = new VideoRenderer();
            renderer.delay(delayFunc);
            renderer.isVisible = true;
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = {
                close: vi.fn()
            } as unknown as VideoFrame;

            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            expect(delayFunc).toHaveBeenCalledTimes(1);
            expect(console.log).toHaveBeenCalledWith('Rendering delayed');
            expect(mockCanvasContext.drawImage).toHaveBeenCalled();
        });

        test("cancels previous animation frame when new frame arrives", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const animationIds: number[] = [];
            (global.requestAnimationFrame as any).mockImplementation((callback) => {
                const id = Math.floor(Math.random() * MAX_ANIMATION_FRAME_ID);
                animationIds.push(id);
                setTimeout(() => callback(performance.now()), 0);
                return id;
            });

            const renderer = new VideoRenderer();
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame1 = { close: vi.fn() } as unknown as VideoFrame;
            const mockVideoFrame2 = { close: vi.fn() } as unknown as VideoFrame;

            const writer = writableStream.getWriter();
            
            // Write first frame
            await writer.write(mockVideoFrame1);
            
            // Write second frame quickly
            await writer.write(mockVideoFrame2);

            // Wait for async operations
            await new Promise(resolve => setTimeout(resolve, 0));

            expect(global.cancelAnimationFrame).toHaveBeenCalledWith(animationIds[0]);
        });

        test("handles undefined frame gracefully", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer();
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const writer = writableStream.getWriter();

            // Allow any pending animation frames from previous operations to flush
            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));
            mockCanvasContext.clearRect.mockClear();
            mockCanvasContext.drawImage.mockClear();

            await writer.write(undefined as any);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            // Should not perform any rendering for undefined frames
            expect(mockCanvasContext.clearRect).not.toHaveBeenCalled();
            expect(mockCanvasContext.drawImage).not.toHaveBeenCalled();
        });
    });

    describe("delay() method", () => {
        test("sets delay function", () => {
            const renderer = new VideoRenderer();
            const delayFunc = vi.fn();

            renderer.delay(delayFunc);

            expect(renderer.delayFunc).toBe(delayFunc);
        });

        test("clears delay function when called with undefined", () => {
            const renderer = new VideoRenderer();
            const delayFunc = vi.fn();

            renderer.delay(delayFunc);
            expect(renderer.delayFunc).toBe(delayFunc);

            renderer.delay(undefined);
            expect(renderer.delayFunc).toBeUndefined();
        });

        test("clears delay function when called with no arguments", () => {
            const renderer = new VideoRenderer();
            const delayFunc = vi.fn();

            renderer.delay(delayFunc);
            expect(renderer.delayFunc).toBe(delayFunc);

            renderer.delay();
            expect(renderer.delayFunc).toBeUndefined();
        });
    });

    describe("destroy()", () => {
        test("cancels animation frame and disconnects observer", () => {
            const renderer = new VideoRenderer();
            renderer.animateId = 12345;

            renderer.destroy();

            expect(global.cancelAnimationFrame).toHaveBeenCalledWith(12345);
            expect(mockIntersectionObserver.disconnect).toHaveBeenCalledTimes(1);
        });

        test("handles missing animation frame gracefully", () => {
            const renderer = new VideoRenderer();
            renderer.animateId = undefined;

            expect(() => renderer.destroy()).not.toThrow();
            expect(global.cancelAnimationFrame).not.toHaveBeenCalled();
        });

        test("closes decoder and handles errors", async () => {
            const mockDecoder = {
                close: vi.fn().mockImplementation(() => {
                    throw new Error("Decoder close failed");
                })
            };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoder as any);

            const renderer = new VideoRenderer();
            await renderer.decoder(); // Initialize decoder

            renderer.destroy();

            expect(mockDecoder.close).toHaveBeenCalledTimes(1);
            expect(console.warn).toHaveBeenCalledWith('Error closing decoder during destroy:', expect.any(Error));
        });

        test("handles decoder close success", async () => {
            const mockDecoder = {
                close: vi.fn()
            };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoder as any);

            const renderer = new VideoRenderer();
            await renderer.decoder();

            renderer.destroy();

            expect(mockDecoder.close).toHaveBeenCalledTimes(1);
            expect(console.warn).not.toHaveBeenCalled();
        });

        test("handles missing decoder gracefully", () => {
            const renderer = new VideoRenderer();

            expect(() => renderer.destroy()).not.toThrow();
        });

        test("handles missing observer gracefully", () => {
            const renderer = new VideoRenderer({ backgroundRendering: true });

            expect(() => renderer.destroy()).not.toThrow();
            expect(mockIntersectionObserver.disconnect).not.toHaveBeenCalled();
        });
    });

    describe("Integration Scenarios", () => {
        test("complete video rendering pipeline", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            // Create renderer with custom settings
            const init: VideoRendererInit = {
                width: 640,
                height: 480,
                intersectionThreshold: 0.1,
                backgroundRendering: false
            };

            const renderer = new VideoRenderer(init);

            // Verify initialization
            expect(mockCanvas.width).toBe(640);
            expect(mockCanvas.height).toBe(480);
            expect(intersectionObserverConstructor).toHaveBeenCalledWith(
                expect.any(Function),
                { threshold: 0.1 }
            );

            // Get decoder
            const decoder = await renderer.decoder();
            expect(decoder).toBe(mockDecoderInstance);

            // Test visibility change
            const callback = mockIntersectionObserver.callback as IntersectionObserverCallback;
            callback([{ isIntersecting: false } as any], mockIntersectionObserver);
            expect(renderer.isVisible).toBe(false);

            // Process video frame
            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = { close: vi.fn() } as unknown as VideoFrame;
            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            // Should skip rendering due to invisibility
            expect(mockCanvasContext.drawImage).not.toHaveBeenCalled();
            expect(mockVideoFrame.close).toHaveBeenCalledTimes(1);

            // Make visible and process another frame
            callback([{ isIntersecting: true } as any], mockIntersectionObserver);
            expect(renderer.isVisible).toBe(true);

            const mockVideoFrame2 = { close: vi.fn() } as unknown as VideoFrame;
            await writer.write(mockVideoFrame2);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            // Should render this time
            expect(mockCanvasContext.clearRect).toHaveBeenCalledWith(0, 0, 640, 480);
            expect(mockCanvasContext.drawImage).toHaveBeenCalledWith(mockVideoFrame2, 0, 0, 640, 480);

            // Cleanup
            renderer.destroy();
            expect(global.cancelAnimationFrame).toHaveBeenCalled();
            expect(mockIntersectionObserver.disconnect).toHaveBeenCalled();
        });

        test("background rendering workflow", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer({ backgroundRendering: true });

            // Should not create observer
            expect(intersectionObserverConstructor).not.toHaveBeenCalled();
            expect(renderer.isVisible).toBe(true);

            // Get decoder and process frame
            await renderer.decoder();
            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = { close: vi.fn() } as unknown as VideoFrame;
            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            // Should always render
            expect(mockCanvasContext.drawImage).toHaveBeenCalledWith(mockVideoFrame, 0, 0, 320, 240);
            expect(mockVideoFrame.close).toHaveBeenCalledTimes(1);
        });

        test("handles rapid frame processing", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer();
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;
            const writer = writableStream.getWriter();

            // Process multiple frames rapidly
            const frames = Array.from({ length: 5 }, (_, i) => ({
                close: vi.fn(),
                frameNumber: i
            } as unknown as VideoFrame));

            for (const frame of frames) {
                await writer.write(frame);
            }

            await new Promise(resolve => setTimeout(resolve, RAPID_FRAME_DELAY));

            // All frames should be closed
            frames.forEach(frame => {
                expect(frame.close).toHaveBeenCalledTimes(1);
            });

            // Should cancel previous animation frames
            expect(global.cancelAnimationFrame).toHaveBeenCalledTimes(4); // 5 frames - 1
        });

        test("delay function integration", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            let delayResolve: () => void;
            const delayPromise = new Promise<void>(resolve => {
                delayResolve = resolve;
            });

            const delayFunc = vi.fn().mockReturnValue(delayPromise);
            const renderer = new VideoRenderer();
            renderer.delay(delayFunc);
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = { close: vi.fn() } as unknown as VideoFrame;
            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            // Should call delay function but not render yet
            expect(delayFunc).toHaveBeenCalledTimes(1);
            expect(mockCanvasContext.drawImage).not.toHaveBeenCalled();

            // Resolve delay
            delayResolve!();
            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            // Now should render
            expect(mockCanvasContext.drawImage).toHaveBeenCalledWith(mockVideoFrame, 0, 0, 320, 240);
            expect(mockVideoFrame.close).toHaveBeenCalledTimes(1);
        });
    });

    describe("Edge Cases and Error Handling", () => {
        test("handles canvas context creation failure", () => {
            mockCanvas.getContext.mockReturnValueOnce(null);

            expect(() => new VideoRenderer()).toThrow('Failed to acquire 2D canvas context');
        });

        test("handles document.createElement failure", () => {
            createElementSpy.mockImplementationOnce(() => {
                throw new Error("Canvas creation failed");
            });

            expect(() => new VideoRenderer()).toThrow("Canvas creation failed");
        });

        test("handles IntersectionObserver creation failure", () => {
            intersectionObserverConstructor.mockImplementationOnce(() => {
                throw new Error("IntersectionObserver failed");
            });

            expect(() => new VideoRenderer()).toThrow("IntersectionObserver failed");
        });

        test("handles rendering with invalid canvas dimensions", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const renderer = new VideoRenderer({ width: 0, height: 0 });
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = { close: vi.fn() } as unknown as VideoFrame;
            const writer = writableStream.getWriter();
            await writer.write(mockVideoFrame);

            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));

            // Should still attempt to render
            expect(mockCanvasContext.clearRect).toHaveBeenCalledWith(0, 0, 0, 0);
            expect(mockCanvasContext.drawImage).toHaveBeenCalled();
        });

        test("handles delay function that throws error", async () => {
            const mockDecoderInstance = { decoder: 'mock' };
            (VideoTrackDecoder as any).mockImplementation(() => mockDecoderInstance as any);

            const delayFunc = vi.fn(() => Promise.reject(new Error("Delay failed")));
            const renderer = new VideoRenderer();
            renderer.delay(delayFunc);
            await renderer.decoder();

            const decoderCall = (VideoTrackDecoder as any).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockVideoFrame = { close: vi.fn() } as unknown as VideoFrame;
            const writer = writableStream.getWriter();
            
            // Should handle delay error gracefully
            await expect(writer.write(mockVideoFrame)).resolves.toBeUndefined();
            
            await new Promise(resolve => setTimeout(resolve, ASYNC_FRAME_DELAY));
            
            expect(delayFunc).toHaveBeenCalledTimes(1);
            expect(mockVideoFrame.close).toHaveBeenCalledTimes(1);
        });
    });
});
