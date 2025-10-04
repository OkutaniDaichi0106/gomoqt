import { z } from "zod"
import { uint62Schema } from "./integers"
import { TrackSchema } from "./track"

export const DEFAULT_CATALOG_VERSION = "1"

export const RootSchema = z.object({
	version: z.string().default(DEFAULT_CATALOG_VERSION),
	description: z.string().max(500).optional(),
	tracks: z.map(z.string(), TrackSchema), // Map of track names included in this catalog
});

export type CatalogRoot = z.infer<typeof RootSchema>;
