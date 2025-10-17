import { z } from "zod"
import { uint8Schema, uint53Schema } from "../integers"
import { ContainerSchema } from "../container"
import { TrackDescriptorSchema } from "../track"

export const ProfileTrackSchema = TrackDescriptorSchema.extend({
	schema: z.literal('profile'),
	config: z.object({
		id: z.string(),
	}),
});

export type ProfileTrackDescriptor = z.infer<typeof ProfileTrackSchema>;