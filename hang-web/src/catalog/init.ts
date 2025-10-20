import { z } from "zod"
import { uint62Schema } from "./integers"
import { TrackDescriptorSchema } from "./track"

export const DEFAULT_CATALOG_VERSION = "@gomoqt/v1"

export const CatalogInitSchema = z.object({
	version: z.string(),
	$schema: z.url().optional(),
});

export type CatalogInit = z.infer<typeof CatalogInitSchema>;
