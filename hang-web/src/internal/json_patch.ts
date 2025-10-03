/**
 * JSON Patch implementation based on RFC 6902
 * Provides types and utilities for JSON patch operations using Zod schemas
 */

import { z } from 'zod';

/**
 * JSON value schema that can be used in patches
 */
const JsonValueSchema: z.ZodSchema<any> = z.lazy(() =>
    z.union([
        z.string(),
        z.number(),
        z.boolean(),
        z.null(),
        z.record(z.string(), JsonValueSchema),
        z.array(JsonValueSchema)
    ])
);

/**
 * JSON Pointer schema as defined in RFC 6901
 */
const JsonPointerSchema = z.string();

/**
 * JSON Patch operation type schema
 */
const JsonPatchOperationTypeSchema = z.enum([
    "add",
    "remove", 
    "replace",
    "move",
    "copy",
    "test"
]);

/**
 * Base schema for all JSON Patch operations
 */
const JsonPatchOperationBaseSchema = z.object({
    op: JsonPatchOperationTypeSchema,
    path: JsonPointerSchema
});

/**
 * Add operation schema - adds a value to the target location
 */
const JsonPatchAddOperationSchema = JsonPatchOperationBaseSchema.extend({
    op: z.literal("add"),
    value: JsonValueSchema
});

/**
 * Remove operation schema - removes the value at the target location
 */
const JsonPatchRemoveOperationSchema = JsonPatchOperationBaseSchema.extend({
    op: z.literal("remove")
});

/**
 * Replace operation schema - replaces the value at the target location
 */
const JsonPatchReplaceOperationSchema = JsonPatchOperationBaseSchema.extend({
    op: z.literal("replace"),
    value: JsonValueSchema
});

/**
 * Move operation schema - removes the value at a specified location and adds it to the target location
 */
const JsonPatchMoveOperationSchema = JsonPatchOperationBaseSchema.extend({
    op: z.literal("move"),
    from: JsonPointerSchema
});

/**
 * Copy operation schema - copies the value at a specified location to the target location
 */
const JsonPatchCopyOperationSchema = JsonPatchOperationBaseSchema.extend({
    op: z.literal("copy"),
    from: JsonPointerSchema
});

/**
 * Test operation schema - tests that a value at the target location is equal to a specified value
 */
const JsonPatchTestOperationSchema = JsonPatchOperationBaseSchema.extend({
    op: z.literal("test"),
    value: JsonValueSchema
});

/**
 * Union schema for all JSON Patch operations
 */
const JsonPatchOperationSchema = z.discriminatedUnion("op", [
    JsonPatchAddOperationSchema,
    JsonPatchRemoveOperationSchema,
    JsonPatchReplaceOperationSchema,
    JsonPatchMoveOperationSchema,
    JsonPatchCopyOperationSchema,
    JsonPatchTestOperationSchema
]);

/**
 * JSON Patch schema - array of operations
 */
const JsonPatchSchema = z.array(JsonPatchOperationSchema);

// Export schemas
export {
    JsonValueSchema,
    JsonPointerSchema,
    JsonPatchOperationTypeSchema,
    JsonPatchOperationSchema,
    JsonPatchSchema,
    JsonPatchAddOperationSchema,
    JsonPatchRemoveOperationSchema,
    JsonPatchReplaceOperationSchema,
    JsonPatchMoveOperationSchema,
    JsonPatchCopyOperationSchema,
    JsonPatchTestOperationSchema
};

// Export types inferred from schemas
export type JsonValue = z.infer<typeof JsonValueSchema>;
export type JsonPointer = z.infer<typeof JsonPointerSchema>;
export type JsonPatchOperationType = z.infer<typeof JsonPatchOperationTypeSchema>;
export type JsonPatchOperation = z.infer<typeof JsonPatchOperationSchema>;
export type JsonPatch = z.infer<typeof JsonPatchSchema>;

export type JsonPatchAddOperation = z.infer<typeof JsonPatchAddOperationSchema>;
export type JsonPatchRemoveOperation = z.infer<typeof JsonPatchRemoveOperationSchema>;
export type JsonPatchReplaceOperation = z.infer<typeof JsonPatchReplaceOperationSchema>;
export type JsonPatchMoveOperation = z.infer<typeof JsonPatchMoveOperationSchema>;
export type JsonPatchCopyOperation = z.infer<typeof JsonPatchCopyOperationSchema>;
export type JsonPatchTestOperation = z.infer<typeof JsonPatchTestOperationSchema>;

/**
 * Error thrown when JSON Patch operation fails
 */
export class JsonPatchError extends Error {
    constructor(
        message: string,
        public readonly operation: JsonPatchOperation,
        public readonly path: string
    ) {
        super(message);
        this.name = 'JsonPatchError';
    }
}
