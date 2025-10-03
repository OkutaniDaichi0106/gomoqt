import { RoomElement } from "./room";
import { describe, test, expect, beforeEach, vi, it, beforeAll, afterEach } from 'vitest';

vi.mock("../room", () => ({
    Room: vi.fn().mockImplementation(() => ({
        join: vi.fn(),
        leave: vi.fn(),
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
        customElements.define("hang-room", RoomElement);
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
});
