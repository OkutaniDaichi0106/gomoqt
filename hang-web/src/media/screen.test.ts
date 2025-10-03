import { describe, test, expect, beforeEach, vi } from 'vitest';
import { Screen, ScreenProps } from "./screen";

// Mock navigator.mediaDevices.getDisplayMedia
const mockGetDisplayMedia = vi.fn<(constraints?: DisplayMediaStreamOptions) => Promise<MediaStream>>();

Object.defineProperty(global, 'navigator', {
    writable: true,
    configurable: true,
    value: {
        mediaDevices: {
            getDisplayMedia: mockGetDisplayMedia
        }
    },
});

// Helper to create properly typed mock MediaStream
function createMockStream(
    videoTracks: MediaStreamTrack[],
    audioTracks: MediaStreamTrack[] = []
): MediaStream {
    const allTracks = [...videoTracks, ...audioTracks];
    return {
        getVideoTracks: vi.fn().mockReturnValue(videoTracks),
        getAudioTracks: vi.fn().mockReturnValue(audioTracks),
        getTracks: vi.fn().mockReturnValue(allTracks)
    } as unknown as MediaStream;
}

describe("Screen", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe("Constructor", () => {
        test("creates screen with default props", () => {
            const screen = new Screen();

            expect(screen.enabled).toBe(false);
            expect(screen.constraints).toBeUndefined();
        });

        test("creates screen with enabled=true", () => {
            const screen = new Screen({ enabled: true });

            expect(screen.enabled).toBe(true);
            expect(screen.constraints).toBeUndefined();
        });

        test("creates screen with constraints", () => {
            const constraints = { 
                video: { width: 1920, height: 1080 }, 
                audio: true 
            };
            const screen = new Screen({ constraints });

            expect(screen.enabled).toBe(false);
            expect(screen.constraints).toEqual(constraints);
        });

        test("creates screen with all props", () => {
            const constraints = { 
                video: { frameRate: 30 }, 
                audio: { echoCancellation: false } 
            };
            const props: ScreenProps = {
                enabled: true,
                constraints
            };

            const screen = new Screen(props);

            expect(screen.enabled).toBe(true);
            expect(screen.constraints).toEqual(constraints);
        });
    });

    describe("getVideoTrack", () => {
        test("gets video track when enabled", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio", 
                id: "screen-audio-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            const track = await screen.getVideoTrack();

            expect(track).toBe(mockVideoTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(undefined);
        });

        test("gets video track with constraints", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1", 
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack])
            } as unknown as MediaStream;

            const constraints = { video: { width: 1280, height: 720 }, audio: false };
            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true, constraints });
            const track = await screen.getVideoTrack();

            expect(track).toBe(mockVideoTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(constraints);
        });

        test("returns cached video track on subsequent calls", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // First call
            const track1 = await screen.getVideoTrack();
            expect(track1).toBe(mockVideoTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1);

            // Second call should return cached track
            const track2 = await screen.getVideoTrack();
            expect(track2).toBe(mockVideoTrack);
            expect(track2).toBe(track1);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1); // Not called again
        });

        test("throws error when screen is not enabled", async () => {
            const screen = new Screen({ enabled: false });

            await expect(screen.getVideoTrack()).rejects.toThrow("Screen capture is not enabled");
            expect(mockGetDisplayMedia).not.toHaveBeenCalled();
        });

        test("throws error when screen is disabled by default", async () => {
            const screen = new Screen(); // enabled defaults to false

            await expect(screen.getVideoTrack()).rejects.toThrow("Screen capture is not enabled");
            expect(mockGetDisplayMedia).not.toHaveBeenCalled();
        });

        test("throws error when getDisplayMedia rejects", async () => {
            const displayError = new Error("Screen capture permission denied");
            mockGetDisplayMedia.mockRejectedValue(displayError);

            const screen = new Screen({ enabled: true });

            await expect(screen.getVideoTrack()).rejects.toThrow("Screen capture permission denied");
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(undefined);
        });

        test("throws error when getDisplayMedia returns stream without video track", async () => {
            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([]), // No video tracks
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });

            await expect(screen.getVideoTrack()).rejects.toThrow("Failed to obtain display video track");
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(undefined);
            // All tracks should be stopped since capture failed
            expect(mockStream.getTracks).toHaveBeenCalled();
        });

        test("stops extra tracks from stream", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1", 
                stop: vi.fn()
            };

            const mockExtraTrack = {
                kind: "unknown",
                id: "extra-track",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack, mockExtraTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            await screen.getVideoTrack();

            // Extra track should be stopped
            expect(mockExtraTrack.stop).toHaveBeenCalledTimes(1);
            // Video and audio tracks should not be stopped
            expect(mockVideoTrack.stop).not.toHaveBeenCalled();
            expect(mockAudioTrack.stop).not.toHaveBeenCalled();
        });
    });

    describe("getAudioTrack", () => {
        test("gets audio track when available", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            const track = await screen.getAudioTrack();

            expect(track).toBe(mockAudioTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(undefined);
        });

        test("returns undefined when no audio track available", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            const track = await screen.getAudioTrack();

            expect(track).toBeUndefined();
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(undefined);
        });

        test("returns cached audio track on subsequent calls", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // First call
            const track1 = await screen.getAudioTrack();
            expect(track1).toBe(mockAudioTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1);

            // Second call should return cached track
            const track2 = await screen.getAudioTrack();
            expect(track2).toBe(mockAudioTrack);
            expect(track2).toBe(track1);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1); // Not called again
        });

        test("returns cached undefined on subsequent calls when no audio", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // First call
            const track1 = await screen.getAudioTrack();
            expect(track1).toBeUndefined();
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1);

            // Second call should return cached undefined
            const track2 = await screen.getAudioTrack();
            expect(track2).toBeUndefined();
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1); // Not called again
        });

        test("throws error when screen is not enabled", async () => {
            const screen = new Screen({ enabled: false });

            await expect(screen.getAudioTrack()).rejects.toThrow("Screen capture is not enabled");
            expect(mockGetDisplayMedia).not.toHaveBeenCalled();
        });

        test("uses same stream as video track when both called", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1", 
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // Get video track first
            const videoTrack = await screen.getVideoTrack();
            expect(videoTrack).toBe(mockVideoTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1);

            // Get audio track - should use same stream
            const audioTrack = await screen.getAudioTrack();
            expect(audioTrack).toBe(mockAudioTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1); // Not called again
        });
    });

    describe("close", () => {
        test("stops all tracks when stream exists", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // Get tracks first
            await screen.getVideoTrack();
            expect(mockVideoTrack.stop).not.toHaveBeenCalled();
            expect(mockAudioTrack.stop).not.toHaveBeenCalled();

            // Close the screen
            await screen.close();

            expect(mockVideoTrack.stop).toHaveBeenCalledTimes(1);
            expect(mockAudioTrack.stop).toHaveBeenCalledTimes(1);
        });

        test("stops only video track when no audio", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // Get video track first
            await screen.getVideoTrack();

            // Close the screen
            await screen.close();

            expect(mockVideoTrack.stop).toHaveBeenCalledTimes(1);
        });

        test("does nothing when no stream exists", async () => {
            const screen = new Screen();

            // Should not throw
            await expect(screen.close()).resolves.toBeUndefined();
        });

        test("clears stream reference after closing", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // Get track first
            const track1 = await screen.getVideoTrack();
            expect(track1).toBe(mockVideoTrack);

            // Close the screen
            await screen.close();

            // Verify stream is cleared - next call should get new track
            const mockVideoTrack2 = {
                kind: "video",
                id: "screen-video-track-2",
                stop: vi.fn()
            };

            const mockStream2 = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack2]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack2])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream2);

            const track2 = await screen.getVideoTrack();
            expect(track2).toBe(mockVideoTrack2);
            expect(track2).not.toBe(track1);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(2);
        });

        test("handles video track.stop() throwing error gracefully", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn().mockImplementation(() => {
                    throw new Error("Video stop failed");
                })
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // Get tracks first
            await screen.getVideoTrack();

            // Close should not throw even if video stop() fails
            await expect(screen.close()).resolves.toBeUndefined();
            expect(mockVideoTrack.stop).toHaveBeenCalledTimes(1);
            expect(mockAudioTrack.stop).toHaveBeenCalledTimes(1);
        });

        test("handles audio track.stop() throwing error gracefully", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1",
                stop: vi.fn().mockImplementation(() => {
                    throw new Error("Audio stop failed");
                })
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true });
            
            // Get tracks first
            await screen.getVideoTrack();

            // Close should not throw even if audio stop() fails
            await expect(screen.close()).resolves.toBeUndefined();
            expect(mockVideoTrack.stop).toHaveBeenCalledTimes(1);
            expect(mockAudioTrack.stop).toHaveBeenCalledTimes(1);
        });
    });

    describe("Integration and Real-world Scenarios", () => {
        test("handles complete screen capture lifecycle", async () => {
            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockAudioTrack = {
                kind: "audio",
                id: "screen-audio-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack, mockAudioTrack])
            } as unknown as MediaStream;

            const constraints = { video: { width: 1920, height: 1080 }, audio: true };
            mockGetDisplayMedia.mockResolvedValue(mockStream);

            const screen = new Screen({ enabled: true, constraints });

            // Verify initial state
            expect(screen.enabled).toBe(true);
            expect(screen.constraints).toEqual(constraints);

            // Get video track
            const videoTrack = await screen.getVideoTrack();
            expect(videoTrack).toBe(mockVideoTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(constraints);

            // Get audio track - should use same stream
            const audioTrack = await screen.getAudioTrack();
            expect(audioTrack).toBe(mockAudioTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1);

            // Verify cached behavior
            const videoTrack2 = await screen.getVideoTrack();
            const audioTrack2 = await screen.getAudioTrack();
            expect(videoTrack2).toBe(videoTrack);
            expect(audioTrack2).toBe(audioTrack);
            expect(mockGetDisplayMedia).toHaveBeenCalledTimes(1);

            // Close and cleanup
            await screen.close();
            expect(mockVideoTrack.stop).toHaveBeenCalledTimes(1);
            expect(mockAudioTrack.stop).toHaveBeenCalledTimes(1);
        });

        test("handles screen enable/disable workflow", async () => {
            const screen = new Screen({ enabled: false });

            // Should throw when disabled
            await expect(screen.getVideoTrack()).rejects.toThrow("Screen capture is not enabled");
            await expect(screen.getAudioTrack()).rejects.toThrow("Screen capture is not enabled");

            // Enable screen
            screen.enabled = true;

            const mockVideoTrack = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockStream = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValue(mockStream);

            // Should work when enabled
            const videoTrack = await screen.getVideoTrack();
            expect(videoTrack).toBe(mockVideoTrack);

            const audioTrack = await screen.getAudioTrack();
            expect(audioTrack).toBeUndefined();

            // Disable again
            screen.enabled = false;

            // Should throw again when disabled
            await expect(screen.getVideoTrack()).rejects.toThrow("Screen capture is not enabled");
            await expect(screen.getAudioTrack()).rejects.toThrow("Screen capture is not enabled");

            await screen.close();
        });

        test("handles constraints updates", async () => {
            const screen = new Screen({ enabled: true });

            const mockVideoTrack1 = {
                kind: "video",
                id: "screen-video-track-1",
                stop: vi.fn()
            };

            const mockStream1 = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack1]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack1])
            } as unknown as MediaStream;

            // First call with no constraints
            mockGetDisplayMedia.mockResolvedValueOnce(mockStream1);
            const track1 = await screen.getVideoTrack();
            expect(track1).toBe(mockVideoTrack1);
            expect(mockGetDisplayMedia).toHaveBeenCalledWith(undefined);

            // Close current stream
            await screen.close();

            // Update constraints
            screen.constraints = { video: { width: 1280, height: 720 }, audio: true };

            const mockVideoTrack2 = {
                kind: "video",
                id: "screen-video-track-2",
                stop: vi.fn()
            };

            const mockAudioTrack2 = {
                kind: "audio",
                id: "screen-audio-track-2",
                stop: vi.fn()
            };

            const mockStream2 = {
                getVideoTracks: vi.fn().mockReturnValue([mockVideoTrack2]),
                getAudioTracks: vi.fn().mockReturnValue([mockAudioTrack2]),
                getTracks: vi.fn().mockReturnValue([mockVideoTrack2, mockAudioTrack2])
            } as unknown as MediaStream;

            // Second call with new constraints
            mockGetDisplayMedia.mockResolvedValueOnce(mockStream2);
            const track2 = await screen.getVideoTrack();
            expect(track2).toBe(mockVideoTrack2);
            expect(mockGetDisplayMedia).toHaveBeenLastCalledWith({ video: { width: 1280, height: 720 }, audio: true });

            // Verify audio is now available
            const audioTrack = await screen.getAudioTrack();
            expect(audioTrack).toBe(mockAudioTrack2);

            await screen.close();
        });

        test("handles different screen capture scenarios", async () => {
            const screen = new Screen({ enabled: true });

            // Scenario 1: Video only
            const mockVideoOnlyStream = {
                getVideoTracks: vi.fn().mockReturnValue([{ kind: "video", id: "video-1", stop: vi.fn() }]),
                getAudioTracks: vi.fn().mockReturnValue([]),
                getTracks: vi.fn().mockReturnValue([{ kind: "video", id: "video-1", stop: vi.fn() }])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValueOnce(mockVideoOnlyStream);
            
            const videoTrack = await screen.getVideoTrack();
            expect(videoTrack.kind).toBe("video");
            
            const audioTrack = await screen.getAudioTrack();
            expect(audioTrack).toBeUndefined();

            await screen.close();

            // Scenario 2: Video + Audio
            const mockFullStream = {
                getVideoTracks: vi.fn().mockReturnValue([{ kind: "video", id: "video-2", stop: vi.fn() }]),
                getAudioTracks: vi.fn().mockReturnValue([{ kind: "audio", id: "audio-2", stop: vi.fn() }]),
                getTracks: vi.fn().mockReturnValue([
                    { kind: "video", id: "video-2", stop: vi.fn() },
                    { kind: "audio", id: "audio-2", stop: vi.fn() }
                ])
            } as unknown as MediaStream;

            mockGetDisplayMedia.mockResolvedValueOnce(mockFullStream);
            
            const videoTrack2 = await screen.getVideoTrack();
            expect(videoTrack2.kind).toBe("video");
            
            const audioTrack2 = await screen.getAudioTrack();
            expect(audioTrack2?.kind).toBe("audio");

            await screen.close();
        });
    });
});
