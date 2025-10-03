import { z } from "zod"
import { uint8Schema, uint62Schema } from "./integers"
import { ContainerSchema } from "./container"

export const TrackSchema = z.object({
	name: z.string().min(1),
	description: z.string().max(500).optional(),
	priority: uint8Schema,
	schema: z.string().min(1), // name, URL or path to the track schema
	config: z.object({}).catchall(z.any()), // Flexible config object
	dependencies: z.array(z.string().min(1)).optional(), // List of other track names this track depends on
});

export type TrackDescriptor = z.infer<typeof TrackSchema>;