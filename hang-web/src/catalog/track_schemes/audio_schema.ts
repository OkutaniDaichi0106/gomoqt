import { z } from "zod"
import { uint8Schema, uint53Schema } from "../integers"
import { ContainerSchema } from "../container"
import { TrackSchema } from "../track"


// Mirrors AudioDecoderConfig
// https://w3c.github.io/webcodecs/#audio-decoder-config
// Copied and modified from https://github.com/kixelated/moq/tree/main/js/hang/src/catalog/audio.ts
// Original code is licensed under Apache 2.0 License (https://github.com/kixelated/moq/blob/main/LICENSE)
export const AudioConfigSchema = z.object({
	// See: https://w3c.github.io/webcodecs/codec_registry.html
	codec: z.string(),

	// The description is used for some codecs.
	// If provided, we can initialize the decoder based on the catalog alone.
	// Otherwise, the initialization information is in-band.
	description: z.string().optional(), // hex encoded TODO use base64

	// The sample rate of the audio in Hz
	sampleRate: uint53Schema,

	// The number of channels in the audio
	numberOfChannels: uint53Schema,

	// The bitrate of the audio in bits per second
	// TODO: Support up to Number.MAX_SAFE_INTEGER
	bitrate: uint53Schema.optional(),

	// The container format of the audio
	// e.g. "loc", "cmaf"
	container: ContainerSchema,
});

export const AudioTrackSchema = TrackSchema.extend({
	schema: z.literal('audio'),
	config: AudioConfigSchema,
});

export type AudioTrack = z.infer<typeof AudioTrackSchema>;
