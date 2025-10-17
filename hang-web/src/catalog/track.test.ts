import { describe, test, expect } from 'vitest';
import { z } from 'zod';
import { TrackSchema, TrackDescriptor } from './track';

describe('Track', () => {
    describe('TrackSchema', () => {
        test('accepts valid track objects', () => {
            const validTracks = [
                {
                    name: 'video',
                    priority: 0,
                    schema: 'video-schema',
                    config: {}
                },
                {
                    name: 'audio',
                    description: 'Main audio track',
                    priority: 255,
                    schema: 'https://example.com/audio-schema.json',
                    config: { codec: 'aac', bitrate: 128 }
                },
                {
                    name: 'subtitle',
                    description: 'Japanese subtitles',
                    priority: 100,
                    schema: '/path/to/subtitle-schema',
                    config: { language: 'ja', format: 'vtt' },
                    dependencies: ['video']
                },
                {
                    name: 'complex-track',
                    description: 'A track with complex dependencies',
                    priority: 50,
                    schema: 'complex-schema',
                    config: {
                        nested: { deep: { value: 42 } },
                        array: [1, 2, 3],
                        boolean: true,
                        null_value: null
                    },
                    dependencies: ['video', 'audio', 'metadata']
                }
            ];

            validTracks.forEach(track => {
                const result = TrackSchema.safeParse(track);
                expect(result.success).toBe(true);
                if (result.success) {
                    expect(result.data).toEqual(track);
                }
            });
        });

        test('validates required fields', () => {
            const requiredFields = ['name', 'priority', 'schema', 'config'];
            
            requiredFields.forEach(field => {
                const baseTrack = {
                    name: 'test',
                    priority: 0,
                    schema: 'test-schema',
                    config: {}
                };
                
                const invalidTrack = { ...baseTrack };
                delete invalidTrack[field as keyof typeof invalidTrack];
                
                const result = TrackSchema.safeParse(invalidTrack);
                expect(result.success).toBe(false);
                
                if (!result.success) {
                    expect(result.error.issues.some(issue => 
                        issue.path.includes(field)
                    )).toBe(true);
                }
            });
        });

        test('validates name field', () => {
            const baseTrack = {
                name: 'test',
                priority: 0,
                schema: 'test-schema',
                config: {}
            };

            // Empty string should fail
            const emptyNameResult = TrackSchema.safeParse({
                ...baseTrack,
                name: ''
            });
            expect(emptyNameResult.success).toBe(false);

            // Non-string should fail
            const nonStringNameResult = TrackSchema.safeParse({
                ...baseTrack,
                name: 123
            });
            expect(nonStringNameResult.success).toBe(false);

            // Valid strings should pass
            const validNames = ['a', 'track-name', 'track_name', 'Track Name', '日本語'];
            validNames.forEach(name => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    name
                });
                expect(result.success).toBe(true);
            });
        });

        test('validates description field', () => {
            const baseTrack = {
                name: 'test',
                priority: 0,
                schema: 'test-schema',
                config: {}
            };

            // Description is optional
            const noDescriptionResult = TrackSchema.safeParse(baseTrack);
            expect(noDescriptionResult.success).toBe(true);

            // Valid descriptions
            const validDescriptions = [
                '',
                'Short description',
                'A'.repeat(500), // Exactly 500 characters
                'Unicode 描述 🎵'
            ];

            validDescriptions.forEach(description => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    description
                });
                expect(result.success).toBe(true);
            });

            // Too long description should fail
            const tooLongDescriptionResult = TrackSchema.safeParse({
                ...baseTrack,
                description: 'A'.repeat(501) // 501 characters
            });
            expect(tooLongDescriptionResult.success).toBe(false);

            // Non-string description should fail
            const nonStringDescriptionResult = TrackSchema.safeParse({
                ...baseTrack,
                description: 123
            });
            expect(nonStringDescriptionResult.success).toBe(false);
        });

        test('validates priority field using uint8Schema', () => {
            const baseTrack = {
                name: 'test',
                priority: 0,
                schema: 'test-schema',
                config: {}
            };

            // Valid uint8 values (0-255)
            const validPriorities = [0, 1, 127, 128, 255];
            validPriorities.forEach(priority => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    priority
                });
                expect(result.success).toBe(true);
            });

            // Invalid values
            const invalidPriorities = [-1, 256, 1000, 3.14, '100', null, undefined];
            invalidPriorities.forEach(priority => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    priority
                });
                expect(result.success).toBe(false);
            });
        });

        test('validates schema field', () => {
            const baseTrack = {
                name: 'test',
                priority: 0,
                schema: 'test-schema',
                config: {}
            };

            // Empty string should fail
            const emptySchemaResult = TrackSchema.safeParse({
                ...baseTrack,
                schema: ''
            });
            expect(emptySchemaResult.success).toBe(false);

            // Non-string should fail
            const nonStringSchemaResult = TrackSchema.safeParse({
                ...baseTrack,
                schema: 123
            });
            expect(nonStringSchemaResult.success).toBe(false);

            // Valid schemas
            const validSchemas = [
                'schema-name',
                'https://example.com/schema.json',
                '/path/to/schema',
                'urn:schema:example',
                'schema_with_underscores',
                '日本語スキーマ'
            ];

            validSchemas.forEach(schema => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    schema
                });
                expect(result.success).toBe(true);
            });
        });

        test('validates config field', () => {
            const baseTrack = {
                name: 'test',
                priority: 0,
                schema: 'test-schema',
                config: {}
            };

            // Config is required
            const noConfigResult = TrackSchema.safeParse({
                name: 'test',
                priority: 0,
                schema: 'test-schema'
            });
            expect(noConfigResult.success).toBe(false);

            // Non-object config should fail
            const nonObjectConfigs = [null, 'string', 123, [], true];
            nonObjectConfigs.forEach(config => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    config
                });
                expect(result.success).toBe(false);
            });

            // Various valid configs (catchall allows any properties)
            const validConfigs = [
                {},
                { codec: 'h264' },
                { a: 1, b: 'string', c: true, d: null },
                { nested: { object: { deep: 'value' } } },
                { array: [1, 2, 3], mixed: { types: true } },
                { unicode: '日本語', emoji: '🎵' }
            ];

            validConfigs.forEach(config => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    config
                });
                expect(result.success).toBe(true);
            });
        });

        test('validates dependencies field', () => {
            const baseTrack = {
                name: 'test',
                priority: 0,
                schema: 'test-schema',
                config: {}
            };

            // Dependencies is optional
            const noDependenciesResult = TrackSchema.safeParse(baseTrack);
            expect(noDependenciesResult.success).toBe(true);

            // Valid dependencies arrays
            const validDependencies = [
                [],
                ['video'],
                ['video', 'audio'],
                ['track1', 'track2', 'track3'],
                ['track_with_underscores', 'track-with-dashes'],
                ['日本語トラック']
            ];

            validDependencies.forEach(dependencies => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    dependencies
                });
                expect(result.success).toBe(true);
            });

            // Invalid dependencies
            const invalidDependencies = [
                'string', // Should be array
                123, // Should be array
                [''], // Empty strings not allowed
                [123], // Numbers not allowed
                [null], // Null not allowed
                ['valid', ''], // Mixed valid/invalid
                ['valid', 123] // Mixed valid/invalid
            ];

            invalidDependencies.forEach(dependencies => {
                const result = TrackSchema.safeParse({
                    ...baseTrack,
                    dependencies
                });
                expect(result.success).toBe(false);
            });
        });

        test('works with parse method for valid values', () => {
            const validTrack = {
                name: 'test-track',
                description: 'Test track description',
                priority: 100,
                schema: 'test-schema',
                config: { key: 'value' },
                dependencies: ['parent-track']
            };

            expect(() => TrackSchema.parse(validTrack)).not.toThrow();
            const parsed = TrackSchema.parse(validTrack);
            expect(parsed).toEqual(validTrack);
        });

        test('throws with parse method for invalid values', () => {
            const invalidTracks = [
                {}, // Missing required fields
                { name: '', priority: 0, schema: 'test', config: {} }, // Empty name
                { name: 'test', priority: -1, schema: 'test', config: {} }, // Invalid priority
                { name: 'test', priority: 0, schema: '', config: {} }, // Empty schema
                { name: 'test', priority: 0, schema: 'test' } // Missing config
            ];

            invalidTracks.forEach(track => {
                expect(() => TrackSchema.parse(track)).toThrow(z.ZodError);
            });
        });

        test('schema has correct type definition', () => {
            // Type checking - this should compile without errors
            const track: TrackDescriptor = {
                name: 'test',
                // priority: 0,
                schema: 'test-schema',
                config: {}
            };

            expect(typeof track.name).toBe('string');
            // expect(typeof track.priority).toBe('number');
            expect(typeof track.schema).toBe('string');
            expect(typeof track.config).toBe('object');
        });

        test('integration with complex nested structures', () => {
            const complexTrack = {
                name: 'multimedia-track',
                description: 'Complex multimedia track with all features',
                priority: 200,
                schema: 'https://schemas.example.com/multimedia/v2.json',
                config: {
                    video: {
                        codec: 'h264',
                        bitrate: 5000,
                        resolution: { width: 1920, height: 1080 },
                        framerate: 30
                    },
                    audio: {
                        codec: 'aac',
                        bitrate: 320,
                        channels: 2,
                        sampleRate: 48000
                    },
                    metadata: {
                        title: 'Sample Video',
                        description: 'A sample video for testing',
                        tags: ['test', 'sample', 'video'],
                        timestamps: [0, 30, 60, 90]
                    }
                },
                dependencies: ['base-video', 'base-audio', 'subtitle-track']
            };

            const result = TrackSchema.safeParse(complexTrack);
            expect(result.success).toBe(true);
            
            if (result.success) {
                expect(result.data).toEqual(complexTrack);
                expect(result.data.config.video.codec).toBe('h264');
                expect(result.data.dependencies).toHaveLength(3);
            }
        });

        test('handles edge cases and boundary values', () => {
            // Minimum valid track
            const minTrack = {
                name: 'a', // Minimum length 1
                priority: 0, // Minimum uint8
                schema: 'x', // Minimum length 1
                config: {}
            };

            const minResult = TrackSchema.safeParse(minTrack);
            expect(minResult.success).toBe(true);

            // Maximum valid track
            const maxTrack = {
                name: 'track-with-maximum-length-name-that-is-still-valid',
                description: 'A'.repeat(500), // Maximum length 500
                priority: 255, // Maximum uint8
                schema: 'schema-with-very-long-name-that-should-still-be-valid',
                config: {
                    // Large config object
                    ...Array.from({ length: 100 }, (_, i) => ({ [`key${i}`]: `value${i}` }))
                        .reduce((acc, obj) => ({ ...acc, ...obj }), {})
                },
                dependencies: Array.from({ length: 50 }, (_, i) => `dependency-${i}`)
            };

            const maxResult = TrackSchema.safeParse(maxTrack);
            expect(maxResult.success).toBe(true);
        });

        test('handles additional unknown properties', () => {
            const trackWithExtra = {
                name: 'test',
                priority: 0,
                schema: 'test-schema',
                config: {},
                unknownField: 'should be stripped by schema'
            };

            // Zod object schema strips unknown properties by default
            const result = TrackSchema.safeParse(trackWithExtra);
            expect(result.success).toBe(true);
            
            if (result.success) {
                // Unknown field should be stripped
                expect(result.data).not.toHaveProperty('unknownField');
                expect(result.data).toEqual({
                    name: 'test',
                    priority: 0,
                    schema: 'test-schema',
                    config: {}
                });
            }
        });
    });
});
