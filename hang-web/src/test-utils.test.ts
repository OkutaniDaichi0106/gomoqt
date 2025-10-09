// Common test utilities and mocks for hang-web tests

import { vi } from 'vitest';

// Mock canvas context
export const mockCanvasContext = {
    clearRect: vi.fn(),
    drawImage: vi.fn(),
    fillText: vi.fn(),
    fillStyle: '',
    font: '',
    textAlign: 'left' as CanvasTextAlign,
    textBaseline: 'top' as CanvasTextBaseline
};

// Mock canvas element
export const mockCanvas = {
    getContext: vi.fn((contextType: string) => {
        if (contextType === '2d') {
            return mockCanvasContext;
        }
        return null;
    }),
    width: 320,
    height: 240
};

// Mock video element
export const mockVideo = {
    readyState: 0,
    videoWidth: 640,
    videoHeight: 480,
    currentTime: 0,
    duration: 0,
    paused: true,
    play: vi.fn().mockResolvedValue(undefined),
    pause: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn()
};

// Mock audio context
export const mockAudioWorkletAddModule = vi.fn().mockResolvedValue(undefined);
export const mockAudioContextClose = vi.fn().mockResolvedValue(undefined);

export const mockAudioContext = {
    audioWorklet: {
        addModule: mockAudioWorkletAddModule
    },
    get currentTime() { return this._currentTime || 0; },
    set currentTime(value: number) { this._currentTime = value; },
    _currentTime: 0,
    sampleRate: 44100,
    destination: {},
    close: mockAudioContextClose
};

// Mock gain node - simplified to only provide GainNode interface
export const mockGainNodeConnect = vi.fn();
export const mockGainNodeDisconnect = vi.fn();

export class MockGainNode {
    connect = mockGainNodeConnect;
    disconnect = mockGainNodeDisconnect;
    gain: { value: number; cancelScheduledValues: any; setValueAtTime: any; exponentialRampToValueAtTime: any };
    context: any;

    constructor(audioContext?: any, options?: any) {
        this.gain = {
            value: options?.gain ?? 0.5,
            cancelScheduledValues: vi.fn(),
            setValueAtTime: vi.fn((value: number) => {
                // For testing, immediately set the gain value
                this.gain.value = value;
            }),
            exponentialRampToValueAtTime: vi.fn((value: number) => {
                // For testing, immediately set the gain value
                this.gain.value = value;
            })
        };
        this.context = audioContext || { currentTime: 0 };
    }
}

export const mockGainNode = new MockGainNode();

// Mock audio worklet node
export const mockWorkletConnect = vi.fn();
export const mockWorkletDisconnect = vi.fn();
export const mockWorkletPort = {
    postMessage: vi.fn()
};

export const mockAudioWorkletNode = {
    connect: mockWorkletConnect,
    disconnect: mockWorkletDisconnect,
    port: mockWorkletPort
};

// Global constructor mocks
export function setupGlobalMocks() {
    global.AudioContext = vi.fn(() => mockAudioContext) as any;
    global.GainNode = MockGainNode as any;
    global.AudioWorkletNode = vi.fn(() => mockAudioWorkletNode) as any;
    global.HTMLCanvasElement = vi.fn(() => mockCanvas) as any;
    global.HTMLVideoElement = vi.fn(() => mockVideo) as any;
}

export function resetGlobalMocks() {
    vi.clearAllMocks();
    (global.AudioContext as any).mockImplementation(() => mockAudioContext);
    global.GainNode = MockGainNode as any;
    (global.AudioWorkletNode as any).mockImplementation(() => mockAudioWorkletNode);
    (global.HTMLCanvasElement as any).mockImplementation(() => mockCanvas);
    (global.HTMLVideoElement as any).mockImplementation(() => mockVideo);
}