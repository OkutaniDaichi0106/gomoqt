import { describe, expect, test } from 'vitest';
import { CaptionsTrackSchema } from './captions';

const validDescriptor = {
	name: 'captions-en',
	description: 'English closed captions',
	priority: 1,
	schema: 'captions' as const,
	config: {
		language: 'en-US',
	},
	dependencies: ['video-main'],
};

describe('CaptionsTrackSchema', () => {
	test('accepts a valid captions descriptor', () => {
		const parsed = CaptionsTrackSchema.parse(validDescriptor);

		expect(parsed).toMatchObject(validDescriptor);
	});

	test('rejects descriptors without dependencies', () => {
		const result = CaptionsTrackSchema.safeParse({
			...validDescriptor,
			dependencies: [],
		});

		expect(result.success).toBe(false);
		if (!result.success) {
			expect(result.error.issues[0].path).toContain('dependencies');
		}
	});

	test('rejects descriptors with wrong schema literal', () => {
		const result = CaptionsTrackSchema.safeParse({
			...validDescriptor,
			schema: 'text',
		});

		expect(result.success).toBe(false);
	});
});
