import { ContainerSchema } from './container';
import { z } from 'zod';
import { describe, test, expect, beforeEach, it } from 'vitest';

describe('Container', () => {
    describe('ContainerSchema', () => {
        test('accepts valid container types', () => {
            const validContainers = ['loc', 'cmaf'];

            validContainers.forEach(container => {
                const result = ContainerSchema.safeParse(container);
                expect(result.success).toBe(true);
                if (result.success) {
                    expect(result.data).toBe(container);
                }
            });
        });

        test('rejects invalid container types', () => {
            const invalidContainers = [
                'mp4',
                'webm',
                'hls',
                'dash',
                'invalid',
                'LOC', // Case sensitive
                'CMAF', // Case sensitive
                'loc ', // Trailing space
                ' loc', // Leading space
                'loc-cmaf',
                'cmaf-loc'
            ];

            invalidContainers.forEach(container => {
                const result = ContainerSchema.safeParse(container);
                expect(result.success).toBe(false);
                if (!result.success) {
                    expect(result.error).toBeInstanceOf(z.ZodError);
                    expect(result.error.issues).toHaveLength(1);
                    expect(result.error.issues[0].code).toBe('invalid_value');
                }
            });
        });

        test('rejects non-string values', () => {
            const nonStringValues = [
                123,
                true,
                false,
                null,
                undefined,
                {},
                [],
                Symbol('loc'),
                new Date(),
                /regex/
            ];

            nonStringValues.forEach(value => {
                const result = ContainerSchema.safeParse(value);
                expect(result.success).toBe(false);
                if (!result.success) {
                    expect(result.error).toBeInstanceOf(z.ZodError);
                    expect(result.error.issues).toHaveLength(1);
                    expect(result.error.issues[0].code).toBe('invalid_value');
                }
            });
        });

        test('provides correct error messages for invalid enum values', () => {
            const result = ContainerSchema.safeParse('invalid');
            expect(result.success).toBe(false);
            
            if (!result.success) {
                const issue = result.error.issues[0];
                expect(issue.code).toBe('invalid_value');
                expect(issue.message).toContain('Invalid option');
                expect(issue.message).toContain('"loc"');
                expect(issue.message).toContain('"cmaf"');
            }
        });

        test('provides correct error messages for wrong types', () => {
            const result = ContainerSchema.safeParse(123);
            expect(result.success).toBe(false);
            
            if (!result.success) {
                const issue = result.error.issues[0];
                expect(issue.code).toBe('invalid_value');
                expect(issue.message).toContain('Invalid option');
                expect(issue.message).toContain('"loc"');
                expect(issue.message).toContain('"cmaf"');
            }
        });

        test('works with parse method for valid values', () => {
            expect(() => ContainerSchema.parse('loc')).not.toThrow();
            expect(() => ContainerSchema.parse('cmaf')).not.toThrow();
            
            expect(ContainerSchema.parse('loc')).toBe('loc');
            expect(ContainerSchema.parse('cmaf')).toBe('cmaf');
        });

        test('throws with parse method for invalid values', () => {
            expect(() => ContainerSchema.parse('invalid')).toThrow(z.ZodError);
            expect(() => ContainerSchema.parse(123)).toThrow(z.ZodError);
            expect(() => ContainerSchema.parse('')).toThrow(z.ZodError);
        });

        test('schema has correct type definition', () => {
            // Type-level test - this should compile without errors
            const validValue: z.infer<typeof ContainerSchema> = 'loc';
            expect(validValue).toBe('loc');

            // Should be assignable to string union type
            const containerType: 'loc' | 'cmaf' = ContainerSchema.parse('cmaf');
            expect(containerType).toBe('cmaf');
        });

        test('schema properties and metadata', () => {
            expect(ContainerSchema).toBeInstanceOf(z.ZodEnum);
            // Test enum options indirectly through parsing
            expect(ContainerSchema.options).toContain('loc');
            expect(ContainerSchema.options).toContain('cmaf');
            expect(ContainerSchema.options).toHaveLength(2);
        });

        test('handles edge cases', () => {
            // Empty string
            const emptyResult = ContainerSchema.safeParse('');
            expect(emptyResult.success).toBe(false);

            // Only whitespace
            const whitespaceResult = ContainerSchema.safeParse('   ');
            expect(whitespaceResult.success).toBe(false);

            // Unicode characters
            const unicodeResult = ContainerSchema.safeParse('löc');
            expect(unicodeResult.success).toBe(false);
        });

        test('integration with complex objects', () => {
            const testObject = {
                container: 'loc' as const,
                otherField: 'value'
            };

            const containerValue = ContainerSchema.safeParse(testObject.container);
            expect(containerValue.success).toBe(true);
            if (containerValue.success) {
                expect(containerValue.data).toBe('loc');
            }
        });

        test('array of containers validation', () => {
            const containers = ['loc', 'cmaf'];
            const arraySchema = z.array(ContainerSchema);
            
            const result = arraySchema.safeParse(containers);
            expect(result.success).toBe(true);
            if (result.success) {
                expect(result.data).toEqual(['loc', 'cmaf']);
            }

            // Invalid array
            const invalidResult = arraySchema.safeParse(['loc', 'invalid']);
            expect(invalidResult.success).toBe(false);
        });

        test('optional container schema', () => {
            const optionalSchema = ContainerSchema.optional();
            
            expect(optionalSchema.safeParse('loc').success).toBe(true);
            expect(optionalSchema.safeParse('cmaf').success).toBe(true);
            expect(optionalSchema.safeParse(undefined).success).toBe(true);
            expect(optionalSchema.safeParse('invalid').success).toBe(false);
        });

        test('nullable container schema', () => {
            const nullableSchema = ContainerSchema.nullable();
            
            expect(nullableSchema.safeParse('loc').success).toBe(true);
            expect(nullableSchema.safeParse('cmaf').success).toBe(true);
            expect(nullableSchema.safeParse(null).success).toBe(true);
            expect(nullableSchema.safeParse('invalid').success).toBe(false);
        });

        test('default value with container schema', () => {
            const schemaWithDefault = ContainerSchema.default('loc');
            
            expect(schemaWithDefault.parse('cmaf')).toBe('cmaf');
            expect(schemaWithDefault.parse(undefined)).toBe('loc');
        });
    });
});
