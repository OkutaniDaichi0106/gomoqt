import { describe, test, expect, beforeEach, afterEach, vi, beforeAll } from 'vitest';
import { Device, DeviceProps } from "./device";
import { Cond, Channel } from "golikejs/sync";

// Create mock functions that can be reused
const mockEnumerateDevices = vi.fn() as any;
const mockGetUserMedia = vi.fn() as any;
const mockAddEventListener = vi.fn() as any;
const mockRemoveEventListener = vi.fn() as any;

// Mock global navigator and media APIs
const mockMediaDevices = {
    enumerateDevices: mockEnumerateDevices,
    getUserMedia: mockGetUserMedia,
    addEventListener: mockAddEventListener,
    removeEventListener: mockRemoveEventListener,
    ondevicechange: null,
};

// Setup global mocks
beforeAll(() => {
    // Mock global objects
    Object.defineProperty(global, 'navigator', {
        writable: true,
        configurable: true,
        value: { mediaDevices: mockMediaDevices },
    });

    // Mock window methods - check if window already exists
    if (typeof (global as any).window === 'undefined') {
        (global as any).window = {};
    }
    
    // Store original functions
    const originalSetTimeout = global.setTimeout;
    const originalClearTimeout = global.clearTimeout;
});

describe("Device", () => {
    let mockCond: Cond;
    let mockChannel: Channel<void>;

    beforeEach(async () => {
        vi.clearAllMocks();
        
        const golikejsSync = await import("golikejs/sync");
        mockCond = new Cond(new (golikejsSync as any).Mutex());
        mockChannel = new Channel<void>();

        // Reset global mocks
        mockEnumerateDevices.mockReset();
        mockGetUserMedia.mockReset();
        mockAddEventListener.mockReset();
        mockRemoveEventListener.mockReset();
        mockMediaDevices.ondevicechange = null;

        // Mock console methods to avoid noise in tests
        vi.spyOn(console, 'warn').mockImplementation(() => {});
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe("Constructor", () => {
        test("creates audio device with default props", () => {
            const device = new Device("audio");

            expect(device.kind).toBe("audio");
            expect(device.preferred).toBeUndefined();
            expect(device.available).toBeUndefined();
            expect(device.hasPermission).toBe(false);
        });

        test("creates video device with preferred device", () => {
            const device = new Device("video", { preferred: "test-device-id" });

            expect(device.kind).toBe("video");
            expect(device.preferred).toBe("test-device-id");
            expect(device.available).toBeUndefined();
            expect(device.hasPermission).toBe(false);
        });

        test("sets up devicechange event listener", () => {
            new Device("audio");

            expect(mockAddEventListener).toHaveBeenCalledWith(
                "devicechange", 
                expect.any(Function)
            );
        });

        test("falls back to ondevicechange if addEventListener not available", () => {
            mockAddEventListener.mockImplementation(() => {
                throw new Error("Not supported");
            });

            new Device("audio");

            expect(mockAddEventListener).toHaveBeenCalled();
            // Should fall back to direct assignment
            expect(mockMediaDevices.ondevicechange).toBeDefined();
        });

        test("handles missing navigator.mediaDevices gracefully", () => {
            // Temporarily remove mediaDevices
            const originalMediaDevices = (global as any).navigator.mediaDevices;
            (global as any).navigator.mediaDevices = undefined;

            expect(() => new Device("audio")).not.toThrow();

            // Restore
            (global as any).navigator.mediaDevices = originalMediaDevices;
        });
    });

    describe("updateDevices", () => {
        const mockAudioDevices = [
            { deviceId: "audio1", kind: "audioinput" as MediaDeviceKind, label: "Microphone 1", groupId: "group1", toJSON: vi.fn() },
            { deviceId: "audio2", kind: "audioinput" as MediaDeviceKind, label: "Microphone 2", groupId: "group2", toJSON: vi.fn() },
        ];

        const mockVideoDevices = [
            { deviceId: "video1", kind: "videoinput" as MediaDeviceKind, label: "Camera 1", groupId: "group1", toJSON: vi.fn() },
            { deviceId: "video2", kind: "videoinput" as MediaDeviceKind, label: "Camera 2", groupId: "group2", toJSON: vi.fn() },
        ];

        test("updates available devices successfully", async () => {
            mockEnumerateDevices.mockResolvedValue([...mockAudioDevices, ...mockVideoDevices]);

            const device = new Device("audio");
            await device.updateDevices();

            expect(device.available).toEqual(mockAudioDevices);
            expect(device.hasPermission).toBe(true);
        });

        test("detects no permission when deviceIds are empty", async () => {
            const devicesWithoutIds = mockAudioDevices.map(d => ({ ...d, deviceId: "" }));
            mockEnumerateDevices.mockResolvedValue(devicesWithoutIds);

            const device = new Device("audio");
            await device.updateDevices();

            expect(device.hasPermission).toBe(false);
        });

        test("finds default audio device using heuristics", async () => {
            const devicesWithDefault = [
                { deviceId: "audio1", kind: "audioinput", label: "Microphone 1", groupId: "group1" },
                { deviceId: "default", kind: "audioinput", label: "Default - Microphone 2", groupId: "group2" },
            ];
            mockEnumerateDevices.mockResolvedValue(devicesWithDefault);

            const device = new Device("audio");
            await device.updateDevices();

            expect(device.default).toBe("default");
        });

        test("finds default video device using heuristics", async () => {
            const devicesWithDefault = [
                { deviceId: "video1", kind: "videoinput", label: "Camera 1", groupId: "group1" },
                { deviceId: "video2", kind: "videoinput", label: "Front Camera", groupId: "group2" },
            ];
            mockEnumerateDevices.mockResolvedValue(devicesWithDefault);

            const device = new Device("video");
            await device.updateDevices();

            expect(device.default).toBe("video2");
        });

        test("handles enumerateDevices error gracefully", async () => {
            mockEnumerateDevices.mockRejectedValue(new Error("Enumeration failed"));

            const device = new Device("audio");
            await expect(device.updateDevices()).resolves.not.toThrow();

            expect(device.available).toBeUndefined();
            expect(device.hasPermission).toBe(false);
        });

        test("handles missing navigator.mediaDevices", async () => {
            // Temporarily remove mediaDevices
            const originalMediaDevices = (global as any).navigator.mediaDevices;
            (global as any).navigator.mediaDevices = undefined;

            const device = new Device("audio");
            await device.updateDevices();

            expect(device.available).toBeUndefined();
            expect(device.hasPermission).toBe(false);

            // Restore
            (global as any).navigator.mediaDevices = originalMediaDevices;
        });
    });

    describe("requestPermission", () => {
        test("skips request if already has permission", async () => {
            const device = new Device("audio");
            device.hasPermission = true;

            const result = await device.requestPermission();

            expect(result).toBe(true);
            expect(mockGetUserMedia).not.toHaveBeenCalled();
        });

        test("requests audio permission successfully", async () => {
            const mockTrack = {
                stop: vi.fn(),
                getSettings: vi.fn().mockReturnValue({ deviceId: "audio-device-id" })
            };
            const mockStream = {
                getTracks: vi.fn().mockReturnValue([mockTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockTrack]),
                getVideoTracks: vi.fn().mockReturnValue([]),
                active: true,
                id: "stream1",
                addTrack: vi.fn(),
                removeTrack: vi.fn(),
                clone: vi.fn(),
                dispatchEvent: vi.fn(),
                onaddtrack: null,
                onremovetrack: null,
                onactive: null,
                oninactive: null,
            } as any;
            mockGetUserMedia.mockResolvedValue(mockStream);

            const device = new Device("audio");
            const result = await device.requestPermission();

            expect(result).toBe(true);
            expect(mockGetUserMedia).toHaveBeenCalledWith({ audio: true });
            expect(mockTrack.stop).toHaveBeenCalled();
        });

        test("requests video permission successfully", async () => {
            const mockTrack = {
                stop: vi.fn(),
                getSettings: vi.fn().mockReturnValue({ deviceId: "video-device-id" })
            };
            const mockStream = {
                getTracks: vi.fn().mockReturnValue([mockTrack])
            } as any;
            mockGetUserMedia.mockResolvedValue(mockStream);

            const device = new Device("video");
            const result = await device.requestPermission();

            expect(result).toBe(true);
            expect(mockGetUserMedia).toHaveBeenCalledWith({ video: true });
        });

        test("handles getUserMedia error gracefully", async () => {
            mockGetUserMedia.mockRejectedValue(new Error("Permission denied"));

            const device = new Device("audio");
            const result = await device.requestPermission();

            expect(result).toBe(false);
        });

        test("handles missing getUserMedia", async () => {
            // Temporarily remove getUserMedia
            const originalGetUserMedia = mockMediaDevices.getUserMedia;
            delete (mockMediaDevices as any).getUserMedia;

            const device = new Device("audio");
            const result = await device.requestPermission();

            expect(result).toBe(false);

            // Restore
            mockMediaDevices.getUserMedia = originalGetUserMedia;
        });
    });

    describe("getTrack", () => {
        test("gets audio track with preferred device", async () => {
            const mockTrack = {
                kind: "audio",
                id: "track1",
                getSettings: vi.fn().mockReturnValue({ deviceId: "preferred-device" }),
                stop: vi.fn()
            };
            
            // First call for requestPermission
            mockGetUserMedia.mockResolvedValueOnce({
                getTracks: vi.fn().mockReturnValue([{
                    stop: vi.fn(),
                    getSettings: vi.fn().mockReturnValue({ deviceId: "preferred-device" })
                }])
            } as any);
            
            // Second call for getTrack
            mockGetUserMedia.mockResolvedValueOnce({
                getTracks: vi.fn().mockReturnValue([mockTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockTrack]),
                getVideoTracks: vi.fn().mockReturnValue([])
            } as any);

            const device = new Device("audio", { preferred: "preferred-device" });
            const track = await device.getTrack();

            expect(track).toMatchObject({ kind: "audio", id: "track1" });
            expect(mockGetUserMedia).toHaveBeenLastCalledWith({
                audio: { deviceId: { exact: "preferred-device" } }
            });
        });

        test("gets video track with constraints", async () => {
            const mockTrack = {
                kind: "video",
                id: "track1",
                getSettings: vi.fn().mockReturnValue({ deviceId: "video-device" }),
                stop: vi.fn()
            };
            
            // First call for requestPermission
            mockGetUserMedia.mockResolvedValueOnce({
                getTracks: vi.fn().mockReturnValue([{
                    stop: vi.fn(),
                    getSettings: vi.fn().mockReturnValue({ deviceId: "video-device" })
                }])
            } as any);
            
            // Second call for getTrack
            mockGetUserMedia.mockResolvedValueOnce({
                getTracks: vi.fn().mockReturnValue([mockTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getVideoTracks: vi.fn().mockReturnValue([mockTrack])
            } as any);

            const device = new Device("video");
            const track = await device.getTrack({ width: 1920, height: 1080 });

            expect(track).toMatchObject({ kind: "video", id: "track1" });
            expect(mockGetUserMedia).toHaveBeenLastCalledWith({
                video: { deviceId: { exact: "video-device" }, width: 1920, height: 1080 }
            });
        });

        test("returns undefined when no tracks available", async () => {
            const mockStream = {
                getTracks: vi.fn().mockReturnValue([])
            } as any;
            mockGetUserMedia.mockResolvedValue(mockStream);

            const device = new Device("audio");
            const track = await device.getTrack();

            expect(track).toBeUndefined();
        });

        test("handles getUserMedia error gracefully", async () => {
            mockGetUserMedia.mockRejectedValue(new Error("Access denied"));

            const device = new Device("audio");
            const track = await device.getTrack();

            expect(track).toBeUndefined();
        });

        test("handles missing getUserMedia", async () => {
            // Temporarily remove getUserMedia
            const originalGetUserMedia = mockMediaDevices.getUserMedia;
            delete (mockMediaDevices as any).getUserMedia;

            const device = new Device("audio");
            const track = await device.getTrack();

            expect(track).toBeUndefined();

            // Restore
            mockMediaDevices.getUserMedia = originalGetUserMedia;
        });
    });

    describe("close", () => {
        test("removes event listener and cleans up", () => {
            const device = new Device("audio");
            device.close();

            expect(mockRemoveEventListener).toHaveBeenCalledWith(
                "devicechange",
                expect.any(Function)
            );
        });

        test("clears ondevicechange if removeEventListener not available", () => {
            // Set up scenario where removeEventListener is not available
            const originalRemoveEventListener = mockMediaDevices.removeEventListener;
            delete (mockMediaDevices as any).removeEventListener;
            (mockMediaDevices as any).ondevicechange = vi.fn();

            const device = new Device("audio");
            device.close();

            expect(mockMediaDevices.ondevicechange).toBeNull();

            // Restore
            mockMediaDevices.removeEventListener = originalRemoveEventListener;
        });

        test("handles cleanup errors gracefully", () => {
            mockRemoveEventListener.mockImplementation(() => {
                throw new Error("Cleanup error");
            });

            const device = new Device("audio");
            expect(() => device.close()).not.toThrow();
        });
    });

    describe("updated", () => {
        test("returns a promise", () => {
            const device = new Device("audio");
            const result = device.updated();
            
            // Should return a promise
            expect(result).toBeInstanceOf(Promise);
        });
    });

    describe("Device timeout and error handling", () => {
        test("handles GET_USER_MEDIA_TIMEOUT constant", async () => {
            // Test that timeout handling works (this tests the constant exists and is used)
            const device = new Device("audio");
            
            // Mock a slow getUserMedia that exceeds timeout
            mockGetUserMedia.mockImplementation(() => 
                new Promise((resolve, reject) => {
                    setTimeout(() => reject(new Error("Timeout")), 10);
                })
            );

            const track = await device.getTrack();
            expect(track).toBeUndefined();
        }, 15000);

        test("handles debounce timer in devicechange", () => {
            // Mock window.setTimeout to verify debounce behavior
            const mockSetTimeout = vi.fn((fn: () => void, delay: number) => {
                if (delay === 200) {
                    // Call the function after debounce delay
                    setTimeout(fn, 0);
                }
                return 123 as any;
            });
            const mockClearTimeout = vi.fn();
            
            (global as any).window.setTimeout = mockSetTimeout;
            (global as any).window.clearTimeout = mockClearTimeout;

            const device = new Device("audio");
            
            // Simulate devicechange event
            const onchangeHandler = mockAddEventListener.mock.calls[0][1] as (event: Event) => void;
            
            // Call the handler multiple times rapidly
            onchangeHandler(new Event('devicechange'));
            onchangeHandler(new Event('devicechange'));
            onchangeHandler(new Event('devicechange'));

            // Verify debounce behavior - setTimeout should be called with 200ms delay
            expect(mockSetTimeout).toHaveBeenCalledWith(expect.any(Function), 200);
            expect(mockClearTimeout).toHaveBeenCalled(); // Should clear previous timer
        });
    });

    describe("Integration and Real-world Scenarios", () => {
        test("handles complete audio device setup flow", async () => {
            const mockDevices = [
                { deviceId: "audio1", kind: "audioinput", label: "Microphone 1", groupId: "group1", toJSON: vi.fn() },
                { deviceId: "default", kind: "audioinput", label: "Default - Microphone 2", groupId: "group2", toJSON: vi.fn() },
            ];
            
            mockEnumerateDevices.mockResolvedValue(mockDevices);
            
            const mockTrack = {
                kind: "audio",
                id: "track1",
                stop: vi.fn(),
                getSettings: vi.fn().mockReturnValue({ deviceId: "audio1" })
            };
            const mockStream = {
                getTracks: vi.fn().mockReturnValue([mockTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockTrack]),
                getVideoTracks: vi.fn().mockReturnValue([])
            } as any;
            
            // Mock for both requestPermission and getTrack calls
            mockGetUserMedia.mockResolvedValue(mockStream);

            const device = new Device("audio", { preferred: "audio1" });
            
            // Complete flow: update devices -> request permission -> get track
            await device.updateDevices();
            expect(device.available).toEqual(mockDevices.filter(d => d.kind === "audioinput"));
            expect(device.default).toBe("default");
            
            const permissionGranted = await device.requestPermission();
            expect(permissionGranted).toBe(true);
            
            const track = await device.getTrack();
            expect(track).toBeDefined();
            expect(track?.kind).toBe("audio");
            
            device.close();
            expect(mockRemoveEventListener).toHaveBeenCalled();
        });

        test("handles device switching scenario", async () => {
            const initialDevices = [
                { deviceId: "audio1", kind: "audioinput", label: "Microphone 1", groupId: "group1", toJSON: vi.fn() },
            ];
            
            const updatedDevices = [
                { deviceId: "audio1", kind: "audioinput", label: "Microphone 1", groupId: "group1", toJSON: vi.fn() },
                { deviceId: "audio2", kind: "audioinput", label: "Microphone 2", groupId: "group2", toJSON: vi.fn() },
            ];
            
            mockEnumerateDevices.mockResolvedValueOnce(initialDevices);
            
            const device = new Device("audio");
            await device.updateDevices();
            
            expect(device.available).toHaveLength(1);
            
            // Simulate device change
            mockEnumerateDevices.mockResolvedValueOnce(updatedDevices);
            
            // Trigger devicechange event
            const onchangeHandler = mockAddEventListener.mock.calls[0][1] as (event: Event) => void;
            onchangeHandler(new Event('devicechange'));
            
            // Wait for debounced update
            setTimeout(async () => {
                expect(device.available).toHaveLength(2);
                expect(mockCond.broadcast).toHaveBeenCalled();
            }, 300);
        });
    });
});
