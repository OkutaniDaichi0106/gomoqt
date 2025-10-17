import { z } from "zod"
import { uint8Schema, uint53Schema } from "../integers"
import { ContainerSchema } from "../container"
import { TrackDescriptorSchema } from "../track"

// Helper function to convert hex string to Uint8Array
const hexStringToUint8Array = (hexString: string): Uint8Array => {
	// Remove any whitespace or prefixes like '0x'
	const cleanHex = hexString.replace(/\s+/g, '').replace(/^0x/i, '');
	
	// Validate hex string format
	if (!/^[0-9a-fA-F]*$/.test(cleanHex)) {
		throw new Error(`Invalid hex string: ${hexString}`);
	}
	
	// Convert hex string to Uint8Array
	const bytes = new Uint8Array(cleanHex.length / 2);
	for (let i = 0; i < cleanHex.length; i += 2) {
		bytes[i / 2] = parseInt(cleanHex.substr(i, 2), 16);
	}
	
	return bytes;
};

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
	// Accept both string (hex) and Uint8Array, always output as Uint8Array
	description: z.union([
		z.string().transform(hexStringToUint8Array),
		z.instanceof(Uint8Array)
	]).optional(),

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

export type AudioConfig = z.infer<typeof AudioConfigSchema>;

export const AudioTrackSchema = TrackDescriptorSchema.extend({
	schema: z.literal('audio'),
	config: AudioConfigSchema,
});

export type AudioTrackDescriptor = z.infer<typeof AudioTrackSchema>;
