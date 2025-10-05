import { RoomElement, defineRoom } from "./room";
import { Room } from "../room";
import { describe, test, expect, beforeEach, vi, it, beforeAll, afterEach } from 'vitest';

vi.mock("../room", () => ({
    Room: vi.fn().mockImplementation((config) => ({
        join: vi.fn().mockImplementation(async (session, local) => {
            // Simulate calling onJoin callback
            if (config.onmember?.onJoin) {
                config.onmember.onJoin({ name: "test-member", remote: true });
            }
        }),
        leave: vi.fn().mockImplementation(() => {
            // Simulate calling onLeave callback when leave is called
            if (config.onmember?.onLeave) {
                config.onmember.onLeave({ name: "test-member", remote: true });
            }
        }),
        roomID: "mock-room",
    })),
}));

vi.mock("../internal/audio_hijack_worklet", () => ({
    importWorkletUrl: vi.fn(() => "mock-url"),
}));

vi.mock("../internal/audio_offload_worklet", () => ({
    importUrl: vi.fn(() => "mock-url"),
}));

describe("RoomElement", () => {
    beforeAll(() => {
        defineRoom();
    });

    let element: RoomElement;

    beforeEach(() => {
        element = new RoomElement();
    });

    afterEach(() => {
        document.body.innerHTML = "";
    });

    describe("constructor", () => {
        it("should create an instance", () => {
            expect(element).toBeInstanceOf(RoomElement);
            expect(element).toBeInstanceOf(HTMLElement);
        });
    });

    describe("observedAttributes", () => {
        it("should return correct attributes", () => {
            expect(RoomElement.observedAttributes).toEqual(["room-id", "description"]);
        });
    });

    describe("connectedCallback", () => {
        it("should render the element", () => {
            document.body.appendChild(element);
            expect(element.innerHTML).toContain("room-status-display");
            expect(element.innerHTML).toContain("local-participant");
            expect(element.innerHTML).toContain("remote-participants");
        });
    });

    describe("render", () => {
        it("should render the DOM structure", () => {
            element.render();
            expect(element.innerHTML).toContain("room-status-display");
            expect(element.innerHTML).toContain("local-participant");
            expect(element.innerHTML).toContain("remote-participants");
        });
    });

    describe("attributeChangedCallback", () => {
        it("should handle room-id change", () => {
            const originalLeave = element.leave;
            element.leave = vi.fn();

            // Set mock room
            element.room = { roomID: "old-room" } as any;

            element.attributeChangedCallback('room-id', 'old-room', 'new-room');
            // Now leave should be called
            expect(element.leave).toHaveBeenCalled();

            // Restore
            element.leave = originalLeave;
        });

        it("should not leave room for description change", () => {
            const leaveSpy = vi.spyOn(element, "leave");

            element.room = { roomID: "room" } as any;

            element.attributeChangedCallback('description', 'old', 'new');
            expect(leaveSpy).not.toHaveBeenCalled();
        });
    });

    describe("join", () => {
        it("should join room successfully", async () => {
            const mockSession = {};
            const mockPublisher = { name: "test-publisher" };

            element.setAttribute('room-id', 'test-room');

            await element.join(mockSession as any, mockPublisher as any);

            expect(element.room).toBeDefined();
            expect(element.room?.roomID).toBe("mock-room");
        });

        it("should set error status when room-id is missing", async () => {
            const mockSession = {};
            const mockPublisher = { name: "test-publisher" };

            const statusSpy = vi.fn();
            element.onstatus = statusSpy;

            await element.join(mockSession as any, mockPublisher as any);

            expect(statusSpy).toHaveBeenCalledWith({ type: 'error', message: 'room-id is missing' });
        });

        it("should handle join error", async () => {
            const mockSession = {};
            const mockPublisher = { name: "test-publisher" };

            // Temporarily change the mock to throw
            const RoomMock = vi.mocked(Room);
            RoomMock.mockImplementationOnce(() => {
                throw new Error("Join failed");
            });

            element.setAttribute('room-id', 'test-room');

            const statusSpy = vi.fn();
            element.onstatus = statusSpy;

            await element.join(mockSession as any, mockPublisher as any);

            expect(statusSpy).toHaveBeenCalledWith({ type: 'error', message: 'Failed to join: Join failed' });
        });

        it("should call onjoin callback when member joins", async () => {
            const mockSession = {};
            const mockPublisher = { name: "test-publisher" };

            element.setAttribute('room-id', 'test-room');

            const onjoinSpy = vi.fn();
            element.onjoin = onjoinSpy;

            await element.join(mockSession as any, mockPublisher as any);

            expect(onjoinSpy).toHaveBeenCalledWith({ name: "test-member", remote: true });
        });

        it("should call onleave callback when member leaves", async () => {
            const mockSession = {};
            const mockPublisher = { name: "test-publisher" };

            element.setAttribute('room-id', 'test-room');

            const onleaveSpy = vi.fn();
            element.onleave = onleaveSpy;

            await element.join(mockSession as any, mockPublisher as any);

            // Simulate leave by calling room.leave
            element.room?.leave();

            expect(onleaveSpy).toHaveBeenCalledWith({ name: "test-member", remote: true });
        });
    });

    describe("leave", () => {
        it("should leave room and clear state", () => {
            element.room = { roomID: "test-room", leave: vi.fn() } as any;

            element.leave();

            expect(element.room).toBeUndefined();
        });

        it("should do nothing if no room", () => {
            element.room = undefined;

            expect(() => element.leave()).not.toThrow();
        });

        it("should dispatch statuschange event", () => {
            element.room = { roomID: "test-room", leave: vi.fn() } as any;

            const eventSpy = vi.fn();
            element.addEventListener('statuschange', eventSpy);

            element.leave();

            expect(eventSpy).toHaveBeenCalledWith(expect.objectContaining({
                detail: { type: 'left', message: 'Left room test-room' }
            }));
        });
    });

    describe("disconnectedCallback", () => {
        it("should leave room on disconnect", () => {
            element.room = { roomID: "test-room", leave: vi.fn() } as any;
            const leaveSpy = vi.spyOn(element.room as any, "leave");

            element.disconnectedCallback();

            expect(leaveSpy).toHaveBeenCalled();
        });
    });
});
