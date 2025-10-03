import { describe, expect, test } from 'vitest';

import { ProfileTrackSchema } from './profile';

const createValidDescriptor = () => ({
	name: 'profile-user',
	priority: 3,
	schema: 'profile' as const,
	config: {
		id: 'user-profile',
	},
});

describe('ProfileTrackSchema', () => {
	test('accepts valid profile descriptor', () => {
		const parsed = ProfileTrackSchema.parse(createValidDescriptor());

		expect(parsed.config.id).toBe('user-profile');
	});

	test('requires config.id to be present', () => {
		const result = ProfileTrackSchema.safeParse({
			...createValidDescriptor(),
			config: {},
		});

		expect(result.success).toBe(false);
	});

	test('rejects descriptors with incorrect schema literal', () => {
		const result = ProfileTrackSchema.safeParse({
			...createValidDescriptor(),
			schema: 'profile-custom',
		});

		expect(result.success).toBe(false);
	});
});
