import { z } from "zod"
import { uint8Schema, uint62Schema } from "./integers"
import { ContainerSchema } from "./container"

export const TrackDescriptorSchema = z.object({
	name: z.string().min(1),
	description: z.string().max(500).optional(),
	// priority: uint8Schema,
	schema: z.string().min(1), // name, URL or path to the track schema
	config: z.record(z.string(), z.any()), // Flexible config object as Record<string, any>
	dependencies: z.array(z.string().min(1)).optional(), // List of other track names this track depends on
});

export const TrackDescriptorsSchema = z.array(TrackDescriptorSchema);

export type TrackDescriptor = z.infer<typeof TrackDescriptorSchema>;

export const ActiveTrackSchema = z.object({
    active: z.literal(true),
    track: TrackDescriptorSchema,
})

export type ActiveTrackLine = z.infer<typeof ActiveTrackSchema>;

export const EndedTrackSchema = z.object({
    active: z.literal(false),
    name: z.string(),
})

export type EndedTrackSchema = z.infer<typeof EndedTrackSchema>;

export const CatalogLineSchema = z.union([
    ActiveTrackSchema,
    EndedTrackSchema,
]);

export type CatalogLine = z.infer<typeof CatalogLineSchema>;