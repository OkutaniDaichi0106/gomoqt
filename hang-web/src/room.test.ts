import { describe, it, expect, vi, beforeEach } from "vitest";
import { Room } from "./room";

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
    });

    describe("leave", () => {
        it("should leave the room", async () => {
            await expect(room.leave()).resolves.not.toThrow();
        });
    });
});
