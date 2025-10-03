import { describe, expect, test } from "vitest";
import * as Descriptors from "./index";
import { VideoConfigSchema } from "./video";
import { AudioConfigSchema } from "./audio";
import { ProfileTrackSchema } from "./profile";
import { CaptionsTrackSchema } from "./captions";
import { TimeseriesTrackSchema } from "./timeseries";

describe("catalog descriptors index", () => {
    test("re-exports individual descriptor schemas", () => {
        expect(Descriptors.VideoConfigSchema).toBe(VideoConfigSchema);
        expect(Descriptors.AudioConfigSchema).toBe(AudioConfigSchema);
    expect(Descriptors.ProfileTrackSchema).toBe(ProfileTrackSchema);
        expect(Descriptors.CaptionsTrackSchema).toBe(CaptionsTrackSchema);
        expect(Descriptors.TimeseriesTrackSchema).toBe(TimeseriesTrackSchema);
    });

    test("contains expected descriptor keys", () => {
        const keys = Object.keys(Descriptors);

        expect(keys).toEqual(
            expect.arrayContaining([
                "VideoConfigSchema",
                "AudioConfigSchema",
                "ProfileTrackSchema",
                "CaptionsTrackSchema",
                "TimeseriesTrackSchema",
            ])
        );
    });
});
