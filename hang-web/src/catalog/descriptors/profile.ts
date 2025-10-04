import { z } from "zod"
import { uint8Schema, uint53Schema } from "../integers"
import { ContainerSchema } from "../container"
import { TrackSchema } from "../track"

export const ProfileTrackSchema = TrackSchema.extend({
	schema: z.literal('profile'),
	config: z.object({
		id: z.string(),
	}),
});

export type ProfileTrackDescriptor = z.infer<typeof ProfileTrackSchema>;