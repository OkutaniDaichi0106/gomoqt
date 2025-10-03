import { describe, expect, jest, it } from 'vitest';
import { TrackEncodeStream } from "./encoder_stream";

describe("TrackEncodeStream", () => {
    it("can be constructed with callbacks", () => {
        const output = vi.fn();
        const error = vi.fn();

        const stream = new TrackEncodeStream({
            output,
            error,
        });

        expect(stream).toBeInstanceOf(TrackEncodeStream);
    });
});
