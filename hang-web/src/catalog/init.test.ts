import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { CatalogInitSchema, CatalogInit, DEFAULT_CATALOG_VERSION } from './init';
import { z } from 'zod';

describe('CatalogInit', () => {
    describe('DEFAULT_CATALOG_VERSION constant', () => {
        it('should be defined with correct format', () => {
            expect(DEFAULT_CATALOG_VERSION).toBe('@gomoqt/v1');
        });

        it('should be a string', () => {
            expect(typeof DEFAULT_CATALOG_VERSION).toBe('string');
        });
    });

    describe('CatalogInitSchema', () => {
        describe('version field', () => {
            it('should parse valid version string', () => {
                const input = { version: 'test-version' };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('test-version');
            });

            it('should parse @gomoqt/v1 version', () => {
                const input = { version: '@gomoqt/v1' };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('@gomoqt/v1');
            });

            it('should parse empty string as version', () => {
                const input = { version: '' };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('');
            });

            it('should throw error when version is missing', () => {
                const input = {};

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });

            it('should throw error when version is not a string', () => {
                const input = { version: 123 };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });

            it('should throw error when version is null', () => {
                const input = { version: null };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });

            it('should throw error when version is undefined', () => {
                const input = { version: undefined };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });
        });

        describe('$schema field', () => {
            it('should parse valid URL in $schema field', () => {
                const input = {
                    version: 'v1',
                    $schema: 'https://example.com/schema.json',
                };
                const result = CatalogInitSchema.parse(input);

                expect(result.$schema).toBe('https://example.com/schema.json');
            });

            it('should parse http URL in $schema field', () => {
                const input = {
                    version: 'v1',
                    $schema: 'http://example.com/schema',
                };
                const result = CatalogInitSchema.parse(input);

                expect(result.$schema).toBe('http://example.com/schema');
            });

            it('should allow $schema to be undefined', () => {
                const input = { version: 'v1' };
                const result = CatalogInitSchema.parse(input);

                expect(result.$schema).toBeUndefined();
            });

            it('should allow $schema to be explicitly omitted', () => {
                const input = { version: 'v1', $schema: undefined };
                const result = CatalogInitSchema.parse(input);

                expect(result.$schema).toBeUndefined();
            });

            it('should throw error when $schema is not a valid URL', () => {
                const input = {
                    version: 'v1',
                    $schema: 'not-a-url',
                };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });

            it('should throw error when $schema is a number', () => {
                const input = {
                    version: 'v1',
                    $schema: 123,
                };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });

            it('should throw error when $schema is an empty string', () => {
                const input = {
                    version: 'v1',
                    $schema: '',
                };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });

            it('should throw error when $schema is null', () => {
                const input = {
                    version: 'v1',
                    $schema: null,
                };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });
        });

        describe('type inference', () => {
            it('should correctly infer CatalogInit type', () => {
                const input = {
                    version: '@gomoqt/v1',
                    $schema: 'https://schema.example.com',
                };
                const result: CatalogInit = CatalogInitSchema.parse(input);

                expect(result).toHaveProperty('version');
                expect(result).toHaveProperty('$schema');
            });

            it('should have optional $schema in inferred type', () => {
                const input = { version: 'v1' };
                const result: CatalogInit = CatalogInitSchema.parse(input);

                expect(result.$schema).toBeUndefined();
            });
        });

        describe('edge cases', () => {
            it('should parse version with special characters', () => {
                const input = { version: '@scope/version-1.0.0-beta' };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('@scope/version-1.0.0-beta');
            });

            it('should parse version with numbers and underscores', () => {
                const input = { version: 'v_1_0_0' };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('v_1_0_0');
            });

            it('should parse URL with query parameters', () => {
                const input = {
                    version: 'v1',
                    $schema: 'https://example.com/schema.json?version=1&format=json',
                };
                const result = CatalogInitSchema.parse(input);

                expect(result.$schema).toBe('https://example.com/schema.json?version=1&format=json');
            });

            it('should parse URL with fragment identifier', () => {
                const input = {
                    version: 'v1',
                    $schema: 'https://example.com/schema.json#definitions',
                };
                const result = CatalogInitSchema.parse(input);

                expect(result.$schema).toBe('https://example.com/schema.json#definitions');
            });

            it('should reject additional properties beyond schema', () => {
                const input = {
                    version: 'v1',
                    $schema: 'https://example.com/schema.json',
                    extraField: 'should be ignored or rejected',
                };

                // Zod by default strips unknown properties unless strict mode is enabled
                const result = CatalogInitSchema.parse(input);
                expect(result).not.toHaveProperty('extraField');
            });

            it('should handle version with long string', () => {
                const longVersion = 'v' + 'x'.repeat(1000);
                const input = { version: longVersion };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe(longVersion);
            });

            it('should handle version with whitespace', () => {
                const input = { version: '  v1  ' };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('  v1  ');
            });

            it('should handle version with unicode characters', () => {
                const input = { version: 'バージョン1.0' };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('バージョン1.0');
            });

            it('should parse URL with unicode in query', () => {
                const input = {
                    version: 'v1',
                    $schema: 'https://example.com/schema?name=テスト',
                };
                const result = CatalogInitSchema.parse(input);

                expect(result.$schema).toBe('https://example.com/schema?name=テスト');
            });
        });

        describe('strict validation', () => {
            it('should parse with both fields present', () => {
                const input = {
                    version: '@gomoqt/v1',
                    $schema: 'https://schema.example.com',
                };
                const result = CatalogInitSchema.parse(input);

                expect(result.version).toBe('@gomoqt/v1');
                expect(result.$schema).toBe('https://schema.example.com');
            });

            it('should be strict about type requirements', () => {
                const input = {
                    version: 'v1',
                    $schema: true,
                };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });

            it('should be strict about required version field', () => {
                const input = { $schema: 'https://example.com' };

                expect(() => CatalogInitSchema.parse(input)).toThrow();
            });
        });

        describe('safe parse (non-throwing validation)', () => {
            it('should safely parse valid input', () => {
                const input = {
                    version: 'v1',
                    $schema: 'https://example.com/schema.json',
                };
                const result = CatalogInitSchema.safeParse(input);

                expect(result.success).toBe(true);
                if (result.success) {
                    expect(result.data.version).toBe('v1');
                    expect(result.data.$schema).toBe('https://example.com/schema.json');
                }
            });

            it('should safely fail on invalid input without throwing', () => {
                const input = { version: 123 };
                const result = CatalogInitSchema.safeParse(input);

                expect(result.success).toBe(false);
                if (!result.success) {
                    expect(result.error).toBeDefined();
                }
            });

            it('should provide error information on invalid $schema', () => {
                const input = {
                    version: 'v1',
                    $schema: 'not-a-url',
                };
                const result = CatalogInitSchema.safeParse(input);

                expect(result.success).toBe(false);
            });
        });
    });
});
