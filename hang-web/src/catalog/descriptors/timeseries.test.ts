import { describe, expect, test } from 'vitest';

import { TimeseriesTrackSchema } from './timeseries';

const createMeasurements = () => new Map([
	['temperature', {
		type: 'temperature',
		unit: 'celsius',
		interval: 1,
		min: 0,
		max: 100,
	}],
]);

const createValidDescriptor = () => ({
	name: 'timeseries-temp',
	priority: 2,
	schema: 'timeseries' as const,
	config: {
		measurements: createMeasurements(),
	},
	dependencies: ['sensor-stream'],
});

describe('TimeseriesTrackSchema', () => {
	test('accepts a valid timeseries descriptor', () => {
		const parsed = TimeseriesTrackSchema.parse(createValidDescriptor());

		expect(parsed.config.measurements.get('temperature')).toMatchObject({
			type: 'temperature',
			unit: 'celsius',
			interval: 1,
			min: 0,
			max: 100,
		});
	});

	test('rejects descriptors with non-map measurements', () => {
		const result = TimeseriesTrackSchema.safeParse({
			...createValidDescriptor(),
			config: {
				measurements: { temperature: createMeasurements().get('temperature') },
			},
		});

		expect(result.success).toBe(false);
	});

	test('rejects measurements with invalid interval', () => {
		const invalidMeasurements = new Map([
			['temperature', {
				type: 'temperature',
				unit: 'celsius',
				interval: 0,
			}],
		]);

		const result = TimeseriesTrackSchema.safeParse({
			...createValidDescriptor(),
			config: {
				measurements: invalidMeasurements,
			},
		});

		expect(result.success).toBe(false);
		if (!result.success) {
			expect(result.error.issues[0].path).toContain('interval');
		}
	});
});
