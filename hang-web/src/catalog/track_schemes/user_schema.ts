import { z } from "zod"
import { uint8Schema, uint53Schema } from "../integers"
import { ContainerSchema } from "../container"
import { TrackSchema } from "../track"

export const UserTrackSchema = TrackSchema.extend({
	schema: z.literal('user'),
	config: z.object({
		id: z.uuid(),
		name: z.string().min(1),
		avatar: z.url(),
	}),
});

export type UserTrack = z.infer<typeof UserTrackSchema>;