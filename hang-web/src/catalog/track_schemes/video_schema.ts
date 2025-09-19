import { z } from "zod"
import { uint8Schema, uint53Schema } from "../integers"
import { ContainerSchema } from "../container"
import { TrackSchema } from "../track"

// Based on VideoDecoderConfig
// Copied and modified from https://github.com/kixelated/moq/tree/main/js/hang/src/catalog/video.ts
// Original code is licensed under Apache 2.0 License (https://github.com/kixelated/moq/blob/main/LICENSE)
export const VideoConfigSchema = z.object({
	// See: https://w3c.github.io/webcodecs/codec_registry.html
	codec: z.string(),

	// The description is used for some codecs.
	// If provided, we can initialize the decoder based on the catalog alone.
	// Otherwise, the initialization information is (repeated) before each key-frame.
	description: z.string().optional(), // hex encoded TODO use base64

	// The width and height of the video in pixels
	codedWidth: uint53Schema.optional(),
	codedHeight: uint53Schema.optional(),

	// Ratio of display width/height to coded width/height
	// Allows stretching/squishing individual "pixels" of the video
	// If not provided, the display ratio is 1:1
	displayAspectWidth: uint53Schema.optional(),
	displayAspectHeight: uint53Schema.optional(),

	// The frame rate of the video in frames per second
	framerate: uint53Schema.optional(),

	// The bitrate of the video in bits per second
	// TODO: Support up to Number.MAX_SAFE_INTEGER
	bitrate: uint53Schema.optional(),

	// If true, the decoder will optimize for latency.
	// Default: true
	optimizeForLatency: z.boolean().default(true),

	// The rotation of the video in degrees.
	// Default: 0
	rotation: z.number().default(0),

	// If true, the decoder will flip the video horizontally
	// Default: false
	flip: z.boolean().default(false),

	// The container format of the video
	// e.g. "loc", "cmaf"
	container: ContainerSchema,
});

export type VideoConfig = z.infer<typeof VideoConfigSchema>;

export const VideoTrackSchema = TrackSchema.extend({
	schema: z.literal('video'),
	config: VideoConfigSchema,
});

export type VideoTrack = z.infer<typeof VideoTrackSchema>;