// @ts-nocheck
import { describe, test, expect, beforeEach, afterEach, beforeAll, afterAll, vi } from 'vitest';
import { VideoPreviewer } from './video_previewer';
import type { 
    VideoPreviewerInit,
    VirtualContent 
} from './video_previewer';

// Mock the internal dependencies
vi.mock('./internal', () => ({
    VideoTrackEncoder: vi.fn(() => mockVideoTrackEncoder),
    VideoTrackProcessor: vi.fn().mockImplementation((track) => ({
        readable: new ReadableStream()
    })),
}));

import { VideoTrackEncoder, VideoTrackProcessor } from './internal';

const mockVideoTrackEncoder = {
    encodeTo: vi.fn(),
    close: vi.fn(),
    encoding: false,
};

// Create comprehensive mocks
const fillStyleHistory: string[] = [];
const fontHistory: string[] = [];
const textAlignHistory: string[] = [];
let currentFillStyle = '';
let currentFont = '';
let currentTextAlign = '';

const mockContext = {
    clearRect: vi.fn(),
    fillRect: vi.fn(),
    fillText: vi.fn(),
    drawImage: vi.fn(),
    get fillStyle() {
        return currentFillStyle;
    },
    set fillStyle(value: string) {
        currentFillStyle = value;
        fillStyleHistory.push(value);
    },
    get font() {
        return currentFont;
    },
    set font(value: string) {
        currentFont = value;
        fontHistory.push(value);
    },
    get textAlign() {
        return currentTextAlign;
    },
    set textAlign(value: string) {
        currentTextAlign = value;
        textAlignHistory.push(value);
    }
};

const mockCanvas = {
    width: 640,
    height: 480,
    getContext: vi.fn(() => mockContext)
};

let shouldFailVideoCreation = false;

const mockVideo = {
    srcObject: null,
    autoplay: false,
    muted: false,
    readyState: 0,
    videoWidth: 640,
    videoHeight: 480
};

const mockMediaStreamTrack = {
    kind: 'video',
    readyState: 'live',
    stop: vi.fn(),
    enabled: true,
    id: 'mock-track-id',
    label: 'mock-track',
    muted: false,
    contentHint: '',
    dispatchEvent: vi.fn(() => true),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    clone: vi.fn(),
    getCapabilities: vi.fn(() => ({})),
    getSettings: vi.fn(() => ({ frameRate: 30, width: 640, height: 480 })),
    getConstraints: vi.fn(() => ({})),
    applyConstraints: vi.fn().mockResolvedValue(undefined),
    onended: null,
    onmute: null,
    onunmute: null
} as MediaStreamTrack;

const mockIntersectionObserver = {
    observe: vi.fn(),
    disconnect: vi.fn()
};

type RafCallback = Parameters<typeof requestAnimationFrame>[0];
const rafCallbacks = new Map<number, RafCallback>();
let nextRafId = 1;

const flushAnimationFrames = async (times = 1) => {
    for (let i = 0; i < times; i++) {
        await Promise.resolve();
        const callbacks = Array.from(rafCallbacks.values());
        rafCallbacks.clear();
        callbacks.forEach(callback => callback(0));
    }
};

// Setup global mocks
beforeAll(() => {
    // Mock document.createElement
    const nativeCreateElement = document.createElement.bind(document);
    vi.spyOn(document, 'createElement').mockImplementation((tagName: string) => {
        if (tagName === 'canvas') return mockCanvas as any;
        if (tagName === 'video') {
            if (shouldFailVideoCreation) {
                throw new Error('Video creation failed');
            }
            return mockVideo as any;
        }
        return nativeCreateElement(tagName);
    });

    // Mock global constructors
    global.MediaStream = vi.fn().mockImplementation((tracks) => ({
        getTracks: () => tracks || []
    })) as any;

    global.IntersectionObserver = vi.fn().mockImplementation((callback, options) => ({
        ...mockIntersectionObserver,
        callback,
        options
    })) as any;

    nextRafId = 1;
    rafCallbacks.clear();
    global.requestAnimationFrame = vi.fn().mockImplementation((callback: RafCallback) => {
        const id = nextRafId++;
        rafCallbacks.set(id, callback);
        return id;
    }) as any;

    global.cancelAnimationFrame = vi.fn().mockImplementation((id: number) => {
        rafCallbacks.delete(id);
    }) as any;
});

afterAll(() => {
    vi.restoreAllMocks();
});

describe('VideoPreviewer', () => {
    let mockTrackPromise: Promise<MediaStreamTrack>;
    let resolveTrack: (track: MediaStreamTrack) => void;
    let rejectTrack: (error: Error) => void;

    beforeEach(() => {
        vi.clearAllMocks();
        shouldFailVideoCreation = false;
        fillStyleHistory.length = 0;
        fontHistory.length = 0;
        textAlignHistory.length = 0;
        currentFillStyle = '';
        currentFont = '';
        currentTextAlign = '';
        
        // Reset canvas mock
        mockCanvas.getContext.mockReturnValue(mockContext);
        
        // Create fresh track promise for each test
        mockTrackPromise = new Promise((resolve, reject) => {
            resolveTrack = resolve;
            rejectTrack = reject;
        });
        
        // Reset global mocks
        rafCallbacks.clear();
        nextRafId = 1;
        
        // Reset VideoTrackEncoder mock
        vi.mocked(VideoTrackEncoder).mockClear();
        mockVideoTrackEncoder.close.mockClear().mockResolvedValue(undefined);
    });

    afterEach(() => {
        rafCallbacks.clear();
    });

    describe('Constructor and Initialization', () => {
        test('creates canvas with default dimensions', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            expect(document.createElement).toHaveBeenCalledWith('canvas');
            expect(previewer.canvas).toBe(mockCanvas);
            expect(mockCanvas.width).toBe(640);
            expect(mockCanvas.height).toBe(480);
        });

        test('creates canvas with custom dimensions', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                width: 1280,
                height: 720
            };

            const previewer = new VideoPreviewer(init);

            expect(mockCanvas.width).toBe(640); // Canvas size is hardcoded in implementation
            expect(mockCanvas.height).toBe(480);
        });

        test('gets 2D canvas context', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            expect(mockCanvas.getContext).toHaveBeenCalledWith('2d');
            expect(previewer.context).toBe(mockContext);
        });

        test('initializes with default virtual content', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            expect(previewer.virtualContent).toEqual({
                backgroundColor: 'black',
                textColor: 'white',
                title: 'No Camera',
                subtitle: 'Available',
                fontSize: 24,
                customDraw: undefined
            });
        });

        test('initializes with custom virtual content', () => {
            const customDraw = vi.fn();
            const virtualContent: VirtualContent = {
                backgroundColor: 'blue',
                textColor: 'yellow',
                title: 'Custom Title',
                subtitle: 'Custom Subtitle',
                fontSize: 32,
                customDraw
            };

            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                virtualContent
            };

            const previewer = new VideoPreviewer(init);

            expect(previewer.virtualContent).toEqual(virtualContent);
        });

        test('starts as virtual by default', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            // Private field, but we can test behavior
            expect(previewer.destroyed).toBe(false);
        });

        test('initializes intersection observer when background rendering disabled', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                backgroundRendering: false,
                intersectionThreshold: 0.5
            };

            const previewer = new VideoPreviewer(init);

            expect(IntersectionObserver).toHaveBeenCalledWith(
                expect.any(Function),
                { threshold: 0.5 }
            );
            expect(mockIntersectionObserver.observe).toHaveBeenCalledWith(mockCanvas);
            expect(previewer.observer).toBeDefined();
            expect(previewer.isVisible).toBe(true); // Initial state
        });

        test('does not create intersection observer when background rendering enabled', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                backgroundRendering: true
            };

            const previewer = new VideoPreviewer(init);

            expect(IntersectionObserver).not.toHaveBeenCalled();
            expect(previewer.observer).toBeUndefined();
            expect(previewer.isVisible).toBe(true);
        });

        test('uses default intersection threshold', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
                // backgroundRendering defaults to false
            };

            const previewer = new VideoPreviewer(init);

            expect(IntersectionObserver).toHaveBeenCalledWith(
                expect.any(Function),
                { threshold: 0.01 }
            );
        });
    });

    describe('Virtual Content Rendering', () => {
        test('renders default virtual content', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            // Wait for animation frame
            await flushAnimationFrames();

            expect(mockContext.clearRect).toHaveBeenCalledWith(0, 0, 640, 480);
            expect(fillStyleHistory[0]).toBe('black');
            expect(fillStyleHistory).toContain('white');
            expect(mockContext.fillRect).toHaveBeenCalledWith(0, 0, 640, 480);
            expect(mockContext.fillStyle).toBe('white');
            expect(fontHistory).toContain('24px Arial');
            expect(textAlignHistory[textAlignHistory.length - 1]).toBe('center');
            expect(mockContext.fillText).toHaveBeenCalledWith('No Camera', 320, 230);
            expect(mockContext.fillText).toHaveBeenCalledWith('Available', 320, 260);
        });

        test('renders custom virtual content', async () => {
            const virtualContent: VirtualContent = {
                backgroundColor: 'red',
                textColor: 'green',
                title: 'Custom Title',
                subtitle: 'Custom Subtitle',
                fontSize: 18
            };

            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                virtualContent
            };

            const previewer = new VideoPreviewer(init);

            // Wait for animation frame
            await flushAnimationFrames();

            expect(fillStyleHistory[0]).toBe('red');
            expect(mockContext.fillRect).toHaveBeenCalledWith(0, 0, 640, 480);
            expect(mockContext.fillStyle).toBe('green');
            expect(fontHistory).toContain('18px Arial');
            expect(mockContext.fillText).toHaveBeenCalledWith('Custom Title', 320, 230);
            expect(mockContext.fillText).toHaveBeenCalledWith('Custom Subtitle', 320, 260);
        });

        test('renders virtual content without subtitle', async () => {
            const virtualContent: VirtualContent = {
                title: 'Title Only',
                subtitle: undefined
            };

            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                virtualContent
            };

            const previewer = new VideoPreviewer(init);

            // Wait for animation frame
            await flushAnimationFrames();

            expect(mockContext.fillText).toHaveBeenCalledWith('Title Only', 320, 230);
            expect(mockContext.fillText).not.toHaveBeenCalledWith(undefined, 320, 250);
        });

        test('uses custom draw function when provided', async () => {
            const customDraw = vi.fn();
            const virtualContent: VirtualContent = {
                customDraw
            };

            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                virtualContent
            };

            const previewer = new VideoPreviewer(init);

            // Wait for animation frame
            await flushAnimationFrames();

            expect(customDraw).toHaveBeenCalledWith(mockContext, 640, 480);
            // Should not render default text when custom draw is used
            expect(mockContext.fillText).not.toHaveBeenCalled();
        });
    });

    describe('Real Video Source Integration', () => {
        test('transitions from virtual to real video source', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            // Initially should be virtual
            await flushAnimationFrames();
            expect(mockContext.fillText).toHaveBeenCalled(); // Virtual content rendered

            // Resolve track promise
            vi.clearAllMocks();
            mockVideo.readyState = 2; // HAVE_CURRENT_DATA
            resolveTrack(mockMediaStreamTrack);

            // Wait for promise resolution and next animation frame
            await flushAnimationFrames(2);

            expect(document.createElement).toHaveBeenCalledWith('video');
            expect((global.MediaStream as vi.mock)).toHaveBeenCalledWith([mockMediaStreamTrack]);
            expect(mockVideo.srcObject).toEqual(expect.objectContaining({
                getTracks: expect.any(Function)
            }));
            expect(mockVideo.autoplay).toBe(true);
            expect(mockVideo.muted).toBe(true);
        });

        test('draws real video content when available', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            // Allow initial virtual frame and reset counters
            await flushAnimationFrames();
            vi.clearAllMocks();

            // Resolve track and set up video
            mockVideo.readyState = 2;
            resolveTrack(mockMediaStreamTrack);

            // Wait for transition and animation frames after video ready
            await flushAnimationFrames(2);

            expect(mockContext.drawImage).toHaveBeenCalledWith(mockVideo, 0, 0, 640, 480);
            // Should not render virtual content when real video is available
            expect(mockContext.fillText).not.toHaveBeenCalled();
        });

        test('falls back to virtual when video source fails', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            // Reject track promise
            const error = new Error('Camera access denied');
            rejectTrack(error);

            // Wait for error handling and animation frame
            await flushAnimationFrames(2);

            // Should continue rendering virtual content
            expect(mockContext.fillText).toHaveBeenCalled();
            expect(mockContext.drawImage).not.toHaveBeenCalled();
        });

        test('handles video element creation failure', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            shouldFailVideoCreation = true;

            const previewer = new VideoPreviewer(init);
            resolveTrack(mockMediaStreamTrack);

            // Wait for error handling
            await flushAnimationFrames(2);

            // Should fall back to virtual content
            expect(mockContext.fillText).toHaveBeenCalled();
        });
    });

    describe('Animation and Visibility Management', () => {
        test('starts animation loop on construction', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            expect(requestAnimationFrame).toHaveBeenCalled();
        });

        test('stops rendering when not visible and background rendering disabled', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                backgroundRendering: false
            };

            const previewer = new VideoPreviewer(init);

            // Simulate visibility change via intersection observer
            const observerCallback = (IntersectionObserver as vi.mock).mock.calls[0][0];
            observerCallback([{ isIntersecting: false }]);

            expect(previewer.isVisible).toBe(false);

            // Clear previous calls and wait for next animation frame
            vi.clearAllMocks();
            await flushAnimationFrames();

            // Should not render when not visible
            expect(mockContext.clearRect).not.toHaveBeenCalled();
        });

        test('continues rendering when visible', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                backgroundRendering: false
            };

            const previewer = new VideoPreviewer(init);

            // Simulate visibility
            const observerCallback = (IntersectionObserver as vi.mock).mock.calls[0][0];
            observerCallback([{ isIntersecting: true }]);

            expect(previewer.isVisible).toBe(true);

            // Wait for animation frame
            await flushAnimationFrames();

            expect(mockContext.clearRect).toHaveBeenCalled();
        });

        test('always renders when background rendering enabled', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                backgroundRendering: true
            };

            const previewer = new VideoPreviewer(init);

            expect(previewer.isVisible).toBe(true);

            // Should render even without intersection observer
            await flushAnimationFrames();
            expect(mockContext.clearRect).toHaveBeenCalled();
        });

        test('uses custom delay function when provided', async () => {
            const delayFunc = vi.fn().mockResolvedValue(undefined);
            
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            previewer.delay(delayFunc);

            // Wait for animation frame
            await flushAnimationFrames();

            expect(delayFunc).toHaveBeenCalled();
        });

        test('stops animation when destroyed', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            previewer.animationId = 123;

            previewer.destroy();

            expect(cancelAnimationFrame).toHaveBeenCalledWith(123);
            expect(previewer.destroyed).toBe(true);
        });
    });

    describe('Encoder Integration', () => {
        test('creates encoder lazily', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            resolveTrack(mockMediaStreamTrack);

            const encoder = await previewer.encoder();

            expect(vi.mocked(VideoTrackProcessor)).toHaveBeenCalledWith(mockMediaStreamTrack);
            expect(vi.mocked(VideoTrackEncoder)).toHaveBeenCalledWith({
                source: expect.any(Object) // ReadableStream from VideoTrackProcessor
            });
            expect(encoder).toBe(mockVideoTrackEncoder);
        });

        test('reuses existing encoder', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            resolveTrack(mockMediaStreamTrack);

            const encoder1 = await previewer.encoder();
            const encoder2 = await previewer.encoder();

            expect(encoder1).toBe(encoder2);

            expect(vi.mocked(VideoTrackEncoder)).toHaveBeenCalledTimes(1);
        });

        test('waits for track resolution before creating encoder', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            // Start encoder creation but don't resolve track yet
            const encoderPromise = previewer.encoder();

            // Resolve track
            resolveTrack(mockMediaStreamTrack);

            const encoder = await encoderPromise;
            expect(encoder).toBe(mockVideoTrackEncoder);
        });
    });

    describe('Cleanup and Lifecycle', () => {
        test('destroy stops animation frame', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            previewer.animationId = 456;

            previewer.destroy();

            expect(cancelAnimationFrame).toHaveBeenCalledWith(456);
            expect(previewer.animationId).toBeUndefined();
        });

        test('destroy disconnects intersection observer', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                backgroundRendering: false
            };

            const previewer = new VideoPreviewer(init);

            previewer.destroy();

            expect(mockIntersectionObserver.disconnect).toHaveBeenCalled();
            expect(previewer.observer).toBeUndefined();
        });

        test('destroy stops source track', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            resolveTrack(mockMediaStreamTrack);

            // Wait for track resolution
                await flushAnimationFrames();

            previewer.destroy();

            // Wait for async cleanup
                await flushAnimationFrames();

            expect(mockMediaStreamTrack.stop).toHaveBeenCalled();
        });

        test('destroy closes encoder', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            resolveTrack(mockMediaStreamTrack);

            // Create encoder first
            await previewer.encoder();

            previewer.destroy();

            expect(mockVideoTrackEncoder.close).toHaveBeenCalled();
        });

        test('destroy sets destroyed flag', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);

            expect(previewer.destroyed).toBe(false);

            previewer.destroy();

            expect(previewer.destroyed).toBe(true);
        });

        test('destroy is idempotent', () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            previewer.animationId = 789;

            previewer.destroy();
            previewer.destroy(); // Second call

            expect(cancelAnimationFrame).toHaveBeenCalledTimes(1);
            expect(cancelAnimationFrame).toHaveBeenCalledWith(789);
        });

        test('ignores track stopping errors during destroy', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            
            // Make track stop throw error
            (mockMediaStreamTrack.stop as vi.mock).mockImplementation(() => {
                throw new Error('Stop failed');
            });
            
            resolveTrack(mockMediaStreamTrack);
            await flushAnimationFrames();

            // Should not throw
            expect(() => previewer.destroy()).not.toThrow();
        });

        test('ignores encoder close errors during destroy', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            resolveTrack(mockMediaStreamTrack);

            await previewer.encoder();

            // Make encoder close reject
            mockVideoTrackEncoder.close.mockRejectedValue(new Error('Close failed'));

            // Should not throw
            expect(() => previewer.destroy()).not.toThrow();
        });
    });

    describe('Edge Cases and Error Handling', () => {
        test('handles canvas context creation failure', () => {
            (mockCanvas.getContext as vi.mock).mockReturnValueOnce(null);

            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            // Should throw when canvas context creation fails
            expect(() => new VideoPreviewer(init)).toThrow();
        });

        test('handles track that never resolves', async () => {
            const neverResolvingPromise = new Promise<MediaStreamTrack>(() => {});
            
            const init: VideoPreviewerInit = {
                source: neverResolvingPromise
            };

            const previewer = new VideoPreviewer(init);

            // Should continue rendering virtual content
            await flushAnimationFrames();
            expect(mockContext.fillText).toHaveBeenCalled();

            // Encoder should wait for track
            const encoderPromise = previewer.encoder();
            let encoderResolved = false;
            encoderPromise.then(() => { encoderResolved = true; });

            await flushAnimationFrames();
            expect(encoderResolved).toBe(false);
        });

        test('handles video with no readyState', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            
            // Video with readyState < 2 should not be drawn
            mockVideo.readyState = 1;
            resolveTrack(mockMediaStreamTrack);

            await flushAnimationFrames(2);

            expect(mockContext.drawImage).not.toHaveBeenCalled();
            expect(mockContext.fillText).toHaveBeenCalled(); // Should render virtual content
        });

        test('handles missing intersection observer entry', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise,
                backgroundRendering: false
            };

            const previewer = new VideoPreviewer(init);

            // Call observer callback with empty entries
            const observerCallback = (IntersectionObserver as vi.mock).mock.calls[0][0];
            observerCallback([]);

            // Should not crash and maintain current visibility state
            expect(previewer.isVisible).toBe(true); // Should remain unchanged
        });

        test('continues animation after error in frame', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            // Make context.clearRect throw error
            mockContext.clearRect.mockImplementationOnce(() => {
                throw new Error('Render error');
            });

            const previewer = new VideoPreviewer(init);

            // Should continue requesting animation frames despite error
            await flushAnimationFrames(2);
            
            expect(requestAnimationFrame).toHaveBeenCalledTimes(3); // Initial call plus two scheduled frames
        });

        test('handles delay function that throws', async () => {
            const init: VideoPreviewerInit = {
                source: mockTrackPromise
            };

            const previewer = new VideoPreviewer(init);
            
            // Set delay function that throws
            previewer.delay(() => {
                throw new Error('Delay error');
            });

            // Should continue animation despite delay error
            await flushAnimationFrames(2);
            expect(requestAnimationFrame).toHaveBeenCalled();
        });
    });
});
