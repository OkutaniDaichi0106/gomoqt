import { z } from "zod"
import { uint8Schema, uint62Schema } from "../integers"
import { ContainerSchema } from "../container"
import { TrackDescriptorSchema } from "../track"

export const CaptionsTrackSchema = TrackDescriptorSchema.extend({
	schema: z.literal('captions'),
	dependencies: z.array(z.string().min(1)).min(1), // Must depend on a audio or video track
});

export type CaptionsTrackDescriptor = z.infer<typeof CaptionsTrackSchema>;