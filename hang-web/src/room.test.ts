import { describe, it, expect, vi, beforeEach } from "vitest";
import { Room,participantName,broadcastPath } from "./room";

vi.mock("@okutanidaichi/moqt", () => ({
    validateBroadcastPath: vi.fn((path: string) => path),
    InternalAnnounceErrorCode: 1,
}));

vi.mock("golikejs/context", () => ({
    background: vi.fn(() => Promise.resolve()),
    withCancelCause: vi.fn(() => [{
        done: vi.fn(() => new Promise(() => {})),
        err: vi.fn(() => undefined),
    }, vi.fn()]),
    withPromise: vi.fn(),
}));

vi.mock("./broadcast", () => ({
    BroadcastPublisher: vi.fn().mockImplementation(() => ({
        name: "test-publisher",
    })),
    BroadcastSubscriber: vi.fn().mockImplementation(() => ({
        name: "test-subscriber",
        close: vi.fn(),
    })),
}));

vi.mock("./internal/audio_hijack_worklet", () => ({
    importWorkletUrl: vi.fn(() => "mock-url"),
}));

vi.mock("./internal/audio_offload_worklet", () => ({
    importUrl: vi.fn(() => "mock-url"),
}));

describe("Room", () => {
    let room: Room;
    const mockSession = {
        mux: {
            publish: vi.fn(),
        },
        acceptAnnounce: vi.fn(),
    };
    const mockLocal = {
        name: "local-user",
    };

    beforeEach(() => {
        room = new Room({
            roomID: "test-room",
            onmember: {
                onJoin: vi.fn(),
                onLeave: vi.fn(),
            },
        });
    });

    describe("constructor", () => {
        it("should create an instance with roomID", () => {
            expect(room.roomID).toBe("test-room");
        });
    });

    describe("join", () => {
        it("should join the room", async () => {
            const mockAnnouncementReader = {
                receive: vi.fn().mockResolvedValue([{
                    broadcastPath: "/test-room/local-user.hang",
                    ended: vi.fn().mockResolvedValue(undefined),
                }, null] as any),
                close: vi.fn(),
            };
            mockSession.acceptAnnounce.mockResolvedValue([mockAnnouncementReader, null] as any);

            await expect(room.join(mockSession as any, mockLocal as any)).resolves.not.toThrow();
        });

        it("should handle join errors gracefully", async () => {
            const mockAnnouncementReader = {
                receive: vi.fn().mockResolvedValue([null, new Error("Network error")]),
                close: vi.fn(),
            };
            mockSession.acceptAnnounce.mockResolvedValue([mockAnnouncementReader, null] as any);

            // Should not throw even if announcement reader fails
            await expect(room.join(mockSession as any, mockLocal as any)).resolves.not.toThrow();
        });
    });

    describe("leave", () => {
        it("should leave the room", async () => {
            await expect(room.leave()).resolves.not.toThrow();
        });

        it("should handle leave when not joined", async () => {
            // Leave without joining first
            await expect(room.leave()).resolves.not.toThrow();
        });
    });
});


vi.mock('@okutanidaichi/moqt', () => ({
  validateBroadcastPath: vi.fn((p: string) => p),
}));

import * as moqt from '@okutanidaichi/moqt';

describe('room utils', () => {
  it('broadcastPath calls validateBroadcastPath with constructed path', () => {
    const res = broadcastPath('myroom', 'alice');
    expect(res).toBe('/myroom/alice.hang');

    const mocked = vi.mocked(moqt);
    expect(mocked.validateBroadcastPath).toHaveBeenCalledWith('/myroom/alice.hang');
  });

  it('participantName extracts name from broadcast path', () => {
    expect(participantName('myroom', '/myroom/alice.hang')).toBe('alice');
    expect(participantName('r', '/r/bob.hang')).toBe('bob');
    // when name contains dots or dashes
    expect(participantName('room-x', '/room-x/john.doe.hang')).toBe('john.doe');
  });
});
