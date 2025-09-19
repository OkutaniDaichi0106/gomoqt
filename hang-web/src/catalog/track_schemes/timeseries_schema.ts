import { z } from "zod";
import { TrackSchema } from "../track"
import { uint53Schema } from "../integers";

export const MeasurementSchema = z.object({
	// Measurement type
	// e.g. "temperature", "humidity"
	type: z.string().min(1),

	// Measurement unit
	// e.g. "celsius", "latitude"
	unit: z.string().min(1), 

	// Interval in milliseconds to collect data points
	interval: uint53Schema.min(1),

	// Minimum value for the measurement
	// If min is not provided, it will be handled as -Infinity
	min: uint53Schema.optional(),

	// Maximum value for the measurement
	// If max is not provided, it will be handled as Infinity
	max: uint53Schema.optional(),
});

export const TimeseriesTrackSchema = TrackSchema.extend({
	schema: z.literal('timeseries'),
	config: z.object({
		measurements: z.map(z.string(), MeasurementSchema),
	}),
});

export type TimeseriesTrack = z.infer<typeof TimeseriesTrackSchema>;

