import { describe, test, expect, beforeEach, afterEach } from 'vitest';
import { z } from 'zod';
import {
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
    JsonPatchTestOperationSchema,
    JsonPatchError
} from "./json_patch";
import type {
    JsonValue,
    JsonPointer,
    JsonPatchOperationType,
    JsonPatchOperation,
    JsonPatch,
    JsonPatchAddOperation,
    JsonPatchRemoveOperation,
    JsonPatchReplaceOperation,
    JsonPatchMoveOperation,
    JsonPatchCopyOperation,
    JsonPatchTestOperation
} from "./json_patch";

describe("JsonValueSchema", () => {
    describe("valid values", () => {
        test("accepts string values", () => {
            const result = JsonValueSchema.safeParse("hello");
            expect(result.success).toBe(true);
        });

        test("accepts number values", () => {
            const result = JsonValueSchema.safeParse(42);
            expect(result.success).toBe(true);
        });

        test("accepts boolean values", () => {
            expect(JsonValueSchema.safeParse(true).success).toBe(true);
            expect(JsonValueSchema.safeParse(false).success).toBe(true);
        });

        test("accepts null values", () => {
            const result = JsonValueSchema.safeParse(null);
            expect(result.success).toBe(true);
        });

        test("accepts object values", () => {
            const obj = { name: "test", value: 123 };
            const result = JsonValueSchema.safeParse(obj);
            expect(result.success).toBe(true);
        });

        test("accepts array values", () => {
            const arr = [1, "hello", true, null];
            const result = JsonValueSchema.safeParse(arr);
            expect(result.success).toBe(true);
        });

        test("accepts nested objects", () => {
            const nested = {
                user: {
                    name: "John",
                    settings: {
                        theme: "dark",
                        notifications: true
                    }
                }
            };
            const result = JsonValueSchema.safeParse(nested);
            expect(result.success).toBe(true);
        });

        test("accepts nested arrays", () => {
            const nested = [[1, 2], ["a", "b"], [true, false]];
            const result = JsonValueSchema.safeParse(nested);
            expect(result.success).toBe(true);
        });
    });

    describe("invalid values", () => {
        test("rejects undefined", () => {
            const result = JsonValueSchema.safeParse(undefined);
            expect(result.success).toBe(false);
        });

        test("rejects functions", () => {
            const result = JsonValueSchema.safeParse(() => {});
            expect(result.success).toBe(false);
        });

        test("rejects symbols", () => {
            const result = JsonValueSchema.safeParse(Symbol("test"));
            expect(result.success).toBe(false);
        });
    });
});

describe("JsonPointerSchema", () => {
    describe("valid pointers", () => {
        test("accepts empty string", () => {
            const result = JsonPointerSchema.safeParse("");
            expect(result.success).toBe(true);
        });

        test("accepts root pointer", () => {
            const result = JsonPointerSchema.safeParse("/");
            expect(result.success).toBe(true);
        });

        test("accepts simple path", () => {
            const result = JsonPointerSchema.safeParse("/name");
            expect(result.success).toBe(true);
        });

        test("accepts nested path", () => {
            const result = JsonPointerSchema.safeParse("/user/settings/theme");
            expect(result.success).toBe(true);
        });

        test("accepts array index", () => {
            const result = JsonPointerSchema.safeParse("/items/0");
            expect(result.success).toBe(true);
        });

        test("accepts escaped characters", () => {
            const result = JsonPointerSchema.safeParse("/path~1with~0tilde");
            expect(result.success).toBe(true);
        });
    });

    describe("invalid pointers", () => {
        test("rejects non-string values", () => {
            expect(JsonPointerSchema.safeParse(123).success).toBe(false);
            expect(JsonPointerSchema.safeParse(null).success).toBe(false);
            expect(JsonPointerSchema.safeParse(undefined).success).toBe(false);
        });
    });
});

describe("JsonPatchOperationTypeSchema", () => {
    describe("valid operation types", () => {
        test("accepts all valid operations", () => {
            const validOps = ["add", "remove", "replace", "move", "copy", "test"];
            validOps.forEach(op => {
                const result = JsonPatchOperationTypeSchema.safeParse(op);
                expect(result.success).toBe(true);
            });
        });
    });

    describe("invalid operation types", () => {
        test("rejects invalid operations", () => {
            const invalidOps = ["invalid", "delete", "update", "", null, undefined];
            invalidOps.forEach(op => {
                const result = JsonPatchOperationTypeSchema.safeParse(op);
                expect(result.success).toBe(false);
            });
        });
    });
});

describe("JsonPatchAddOperationSchema", () => {
    test("accepts valid add operation", () => {
        const operation: JsonPatchAddOperation = {
            op: "add",
            path: "/newProperty",
            value: "newValue"
        };
        const result = JsonPatchAddOperationSchema.safeParse(operation);
        expect(result.success).toBe(true);
    });

    test("requires value property", () => {
        const operation = {
            op: "add",
            path: "/newProperty"
            // missing value
        };
        const result = JsonPatchAddOperationSchema.safeParse(operation);
        expect(result.success).toBe(false);
    });

    test("rejects wrong operation type", () => {
        const operation = {
            op: "remove",
            path: "/property",
            value: "value"
        };
        const result = JsonPatchAddOperationSchema.safeParse(operation);
        expect(result.success).toBe(false);
    });
});

describe("JsonPatchRemoveOperationSchema", () => {
    test("accepts valid remove operation", () => {
        const operation: JsonPatchRemoveOperation = {
            op: "remove",
            path: "/property"
        };
        const result = JsonPatchRemoveOperationSchema.safeParse(operation);
        expect(result.success).toBe(true);
    });

    test("does not require value property", () => {
        const operation = {
            op: "remove",
            path: "/property"
        };
        const result = JsonPatchRemoveOperationSchema.safeParse(operation);
        expect(result.success).toBe(true);
    });
});

describe("JsonPatchReplaceOperationSchema", () => {
    test("accepts valid replace operation", () => {
        const operation: JsonPatchReplaceOperation = {
            op: "replace",
            path: "/property",
            value: "newValue"
        };
        const result = JsonPatchReplaceOperationSchema.safeParse(operation);
        expect(result.success).toBe(true);
    });

    test("requires value property", () => {
        const operation = {
            op: "replace",
            path: "/property"
            // missing value
        };
        const result = JsonPatchReplaceOperationSchema.safeParse(operation);
        expect(result.success).toBe(false);
    });
});

describe("JsonPatchMoveOperationSchema", () => {
    test("accepts valid move operation", () => {
        const operation: JsonPatchMoveOperation = {
            op: "move",
            path: "/newLocation",
            from: "/oldLocation"
        };
        const result = JsonPatchMoveOperationSchema.safeParse(operation);
        expect(result.success).toBe(true);
    });

    test("requires from property", () => {
        const operation = {
            op: "move",
            path: "/newLocation"
            // missing from
        };
        const result = JsonPatchMoveOperationSchema.safeParse(operation);
        expect(result.success).toBe(false);
    });
});

describe("JsonPatchCopyOperationSchema", () => {
    test("accepts valid copy operation", () => {
        const operation: JsonPatchCopyOperation = {
            op: "copy",
            path: "/newLocation",
            from: "/sourceLocation"
        };
        const result = JsonPatchCopyOperationSchema.safeParse(operation);
        expect(result.success).toBe(true);
    });

    test("requires from property", () => {
        const operation = {
            op: "copy",
            path: "/newLocation"
            // missing from
        };
        const result = JsonPatchCopyOperationSchema.safeParse(operation);
        expect(result.success).toBe(false);
    });
});

describe("JsonPatchTestOperationSchema", () => {
    test("accepts valid test operation", () => {
        const operation: JsonPatchTestOperation = {
            op: "test",
            path: "/property",
            value: "expectedValue"
        };
        const result = JsonPatchTestOperationSchema.safeParse(operation);
        expect(result.success).toBe(true);
    });

    test("requires value property", () => {
        const operation = {
            op: "test",
            path: "/property"
            // missing value
        };
        const result = JsonPatchTestOperationSchema.safeParse(operation);
        expect(result.success).toBe(false);
    });
});

describe("JsonPatchOperationSchema", () => {
    test("accepts all valid operation types", () => {
        const operations: JsonPatchOperation[] = [
            { op: "add", path: "/new", value: "value" },
            { op: "remove", path: "/old" },
            { op: "replace", path: "/existing", value: "newValue" },
            { op: "move", path: "/new", from: "/old" },
            { op: "copy", path: "/copy", from: "/original" },
            { op: "test", path: "/check", value: "expected" }
        ];

        operations.forEach(operation => {
            const result = JsonPatchOperationSchema.safeParse(operation);
            expect(result.success).toBe(true);
        });
    });

    test("rejects invalid operations", () => {
        const invalidOperations = [
            { op: "invalid", path: "/test" },
            { op: "add", path: "/test" }, // missing value
            { op: "move", path: "/test" }, // missing from
            { path: "/test", value: "value" }, // missing op
            { op: "add", value: "value" } // missing path
        ];

        invalidOperations.forEach(operation => {
            const result = JsonPatchOperationSchema.safeParse(operation);
            expect(result.success).toBe(false);
        });
    });
});

describe("JsonPatchSchema", () => {
    test("accepts empty patch array", () => {
        const patch: JsonPatch = [];
        const result = JsonPatchSchema.safeParse(patch);
        expect(result.success).toBe(true);
    });

    test("accepts single operation", () => {
        const patch: JsonPatch = [
            { op: "add", path: "/new", value: "value" }
        ];
        const result = JsonPatchSchema.safeParse(patch);
        expect(result.success).toBe(true);
    });

    test("accepts multiple operations", () => {
        const patch: JsonPatch = [
            { op: "add", path: "/new", value: "value" },
            { op: "remove", path: "/old" },
            { op: "replace", path: "/existing", value: "newValue" }
        ];
        const result = JsonPatchSchema.safeParse(patch);
        expect(result.success).toBe(true);
    });

    test("rejects non-array values", () => {
        const invalidPatches = [
            { op: "add", path: "/test", value: "value" }, // single operation, not array
            "not an array",
            null,
            undefined
        ];

        invalidPatches.forEach(patch => {
            const result = JsonPatchSchema.safeParse(patch);
            expect(result.success).toBe(false);
        });
    });

    test("rejects array with invalid operations", () => {
        const patch = [
            { op: "add", path: "/valid", value: "value" },
            { op: "invalid", path: "/test" } // invalid operation
        ];
        const result = JsonPatchSchema.safeParse(patch);
        expect(result.success).toBe(false);
    });
});

describe("JsonPatchError", () => {
    test("creates error with operation and path", () => {
        const operation: JsonPatchOperation = {
            op: "add",
            path: "/test",
            value: "value"
        };
        const error = new JsonPatchError("Test error", operation, "/test");

        expect(error).toBeInstanceOf(Error);
        expect(error).toBeInstanceOf(JsonPatchError);
        expect(error.message).toBe("Test error");
        expect(error.operation).toBe(operation);
        expect(error.path).toBe("/test");
        expect(error.name).toBe("JsonPatchError");
    });

    test("creates error with different operation types", () => {
        const operations: JsonPatchOperation[] = [
            { op: "remove", path: "/remove" },
            { op: "replace", path: "/replace", value: "new" },
            { op: "move", path: "/to", from: "/from" },
            { op: "copy", path: "/copy", from: "/original" },
            { op: "test", path: "/test", value: "expected" }
        ];

        operations.forEach(operation => {
            const error = new JsonPatchError("Operation failed", operation, operation.path);
            expect(error.operation).toBe(operation);
            expect(error.path).toBe(operation.path);
        });
    });

    test("error is throwable and catchable", () => {
        const operation: JsonPatchOperation = { op: "add", path: "/test", value: "value" };
        
        expect(() => {
            throw new JsonPatchError("Test error", operation, "/test");
        }).toThrow(JsonPatchError);

        try {
            throw new JsonPatchError("Test error", operation, "/test");
        } catch (error) {
            expect(error).toBeInstanceOf(JsonPatchError);
            if (error instanceof JsonPatchError) {
                expect(error.operation).toBe(operation);
                expect(error.path).toBe("/test");
            }
        }
    });
});

describe("Type Definitions", () => {
    test("JsonValue type accepts all JSON-compatible values", () => {
        // These should compile without TypeScript errors
        const stringValue: JsonValue = "hello";
        const numberValue: JsonValue = 42;
        const booleanValue: JsonValue = true;
        const nullValue: JsonValue = null;
        const objectValue: JsonValue = { key: "value" };
        const arrayValue: JsonValue = [1, 2, 3];

        expect(typeof stringValue).toBe("string");
        expect(typeof numberValue).toBe("number");
        expect(typeof booleanValue).toBe("boolean");
        expect(nullValue).toBe(null);
        expect(typeof objectValue).toBe("object");
        expect(Array.isArray(arrayValue)).toBe(true);
    });

    test("JsonPointer type accepts string values", () => {
        const pointer: JsonPointer = "/path/to/property";
        expect(typeof pointer).toBe("string");
    });

    test("JsonPatchOperationType accepts valid operation strings", () => {
        const operations: JsonPatchOperationType[] = ["add", "remove", "replace", "move", "copy", "test"];
        operations.forEach(op => {
            expect(typeof op).toBe("string");
        });
    });
});

describe("Boundary Value Tests", () => {
    test("handles empty strings in pointers", () => {
        const result = JsonPointerSchema.safeParse("");
        expect(result.success).toBe(true);
    });

    test("handles very long paths", () => {
        const longPath = "/" + "a".repeat(1000);
        const result = JsonPointerSchema.safeParse(longPath);
        expect(result.success).toBe(true);
    });

    test("handles deeply nested objects", () => {
        const createNestedObject = (depth: number): any => {
            if (depth === 0) return "value";
            return { nested: createNestedObject(depth - 1) };
        };

        const deepObject = createNestedObject(100);
        const result = JsonValueSchema.safeParse(deepObject);
        expect(result.success).toBe(true);
    });

    test("handles large arrays", () => {
        const largeArray = new Array(1000).fill(0).map((_, i) => i);
        const result = JsonValueSchema.safeParse(largeArray);
        expect(result.success).toBe(true);
    });

    test("handles special characters in values", () => {
        const specialChars = {
            unicode: "🎉🚀⭐",
            newlines: "line1\nline2\r\nline3",
            tabs: "col1\tcol2\tcol3",
            quotes: 'He said "Hello" to me',
            backslashes: "C:\\Users\\test\\file.txt"
        };
        const result = JsonValueSchema.safeParse(specialChars);
        expect(result.success).toBe(true);
    });

    test("handles valid numeric edge cases", () => {
        const numbers = {
            zero: 0,
            negative: -123,
            decimal: 123.456,
            scientific: 1.23e-10,
            maxSafe: Number.MAX_SAFE_INTEGER,
            minSafe: Number.MIN_SAFE_INTEGER
        };
        const result = JsonValueSchema.safeParse(numbers);
        expect(result.success).toBe(true);
    });

    test("rejects invalid numeric values (infinity, NaN)", () => {
        // JSON doesn't support Infinity or NaN
        const invalidNumbers = [
            Number.POSITIVE_INFINITY,
            Number.NEGATIVE_INFINITY,
            NaN
        ];
        
        invalidNumbers.forEach(num => {
            const result = JsonValueSchema.safeParse(num);
            expect(result.success).toBe(false);
        });
    });
});

describe("Schema Performance Tests", () => {
    test("handles large patch arrays efficiently", () => {
        const largePatch: JsonPatch = Array(1000).fill(0).map((_, i) => ({
            op: "add" as const,
            path: `/item${i}`,
            value: `value${i}`
        }));

        const start = Date.now();
        const result = JsonPatchSchema.safeParse(largePatch);
        const end = Date.now();

        expect(result.success).toBe(true);
        expect(end - start).toBeLessThan(1000); // Should complete within 1 second
    });

    test("validates complex nested patches quickly", () => {
        const complexPatch: JsonPatch = [
            {
                op: "add",
                path: "/complex",
                value: {
                    level1: {
                        level2: {
                            level3: {
                                array: [1, 2, { nested: "value" }],
                                boolean: true,
                                null: null
                            }
                        }
                    }
                }
            },
            {
                op: "replace",
                path: "/complex/level1/level2/level3/array/2/nested",
                value: "updated"
            },
            {
                op: "move",
                path: "/moved",
                from: "/complex/level1"
            }
        ];

        const result = JsonPatchSchema.safeParse(complexPatch);
        expect(result.success).toBe(true);
    });
});

describe("Real-world Use Cases", () => {
    test("validates typical object property updates", () => {
        const userUpdatePatch: JsonPatch = [
            { op: "replace", path: "/name", value: "John Doe" },
            { op: "replace", path: "/age", value: 30 },
            { op: "add", path: "/email", value: "john@example.com" },
            { op: "remove", path: "/temporaryField" }
        ];

        const result = JsonPatchSchema.safeParse(userUpdatePatch);
        expect(result.success).toBe(true);
    });

    test("validates array manipulation operations", () => {
        const arrayPatch: JsonPatch = [
            { op: "add", path: "/items/-", value: "new item" },
            { op: "replace", path: "/items/0", value: "updated first item" },
            { op: "remove", path: "/items/1" },
            { op: "move", path: "/items/0", from: "/items/2" }
        ];

        const result = JsonPatchSchema.safeParse(arrayPatch);
        expect(result.success).toBe(true);
    });

    test("validates configuration updates", () => {
        const configPatch: JsonPatch = [
            { op: "replace", path: "/server/port", value: 8080 },
            { op: "add", path: "/server/ssl", value: { enabled: true, cert: "/path/to/cert" } },
            { op: "replace", path: "/database/host", value: "localhost" },
            { op: "test", path: "/database/type", value: "postgresql" }
        ];

        const result = JsonPatchSchema.safeParse(configPatch);
        expect(result.success).toBe(true);
    });

    test("validates JSON API resource updates", () => {
        const jsonApiPatch: JsonPatch = [
            {
                op: "replace",
                path: "/data/attributes/title",
                value: "Updated Title"
            },
            {
                op: "add",
                path: "/data/relationships/author",
                value: {
                    data: { type: "users", id: "123" }
                }
            },
            {
                op: "replace",
                path: "/data/attributes/publishedAt",
                value: "2023-01-01T00:00:00Z"
            }
        ];

        const result = JsonPatchSchema.safeParse(jsonApiPatch);
        expect(result.success).toBe(true);
    });
});

describe("Error Handling Robustness", () => {
    test("provides detailed error information for invalid operations", () => {
        const invalidPatch = [
            { op: "invalid", path: "/test" }
        ];

        const result = JsonPatchSchema.safeParse(invalidPatch);
        expect(result.success).toBe(false);
        if (!result.success) {
            expect(result.error.issues).toHaveLength(1);
            expect(result.error.issues[0].path).toEqual([0, "op"]);
        }
    });

    test("handles mixed valid and invalid operations", () => {
        const mixedPatch = [
            { op: "add", path: "/valid", value: "test" },
            { op: "invalid", path: "/invalid" },
            { op: "replace", path: "/another", value: 123 }
        ];

        const result = JsonPatchSchema.safeParse(mixedPatch);
        expect(result.success).toBe(false);
        if (!result.success) {
            // Should report error on the invalid operation
            const errorPaths = result.error.issues.map(e => e.path);
            expect(errorPaths.some(path => path.includes(1))).toBe(true);
        }
    });

    test("validates operation-specific required fields", () => {
        const incompleteOperations = [
            { op: "add", path: "/test" }, // Missing value
            { op: "replace", path: "/test" }, // Missing value
            { op: "move", path: "/test" }, // Missing from
            { op: "copy", path: "/test" }, // Missing from
            { op: "test", path: "/test" } // Missing value
        ];

        incompleteOperations.forEach(operation => {
            const result = JsonPatchOperationSchema.safeParse(operation);
            expect(result.success).toBe(false);
        });
    });

    test("JsonPatchError serializes correctly", () => {
        const operation: JsonPatchOperation = { op: "add", path: "/test", value: "value" };
        const error = new JsonPatchError("Test error", operation, "/test");

        // Test JSON serialization
        const serialized = JSON.stringify({
            message: error.message,
            name: error.name,
            operation: error.operation,
            path: error.path
        });

        const parsed = JSON.parse(serialized);
        expect(parsed.message).toBe("Test error");
        expect(parsed.name).toBe("JsonPatchError");
        expect(parsed.operation).toEqual(operation);
        expect(parsed.path).toBe("/test");
    });

    test("JsonPatchError maintains stack trace", () => {
        const operation: JsonPatchOperation = { op: "add", path: "/test", value: "value" };
        const error = new JsonPatchError("Test error", operation, "/test");

        expect(error.stack).toBeDefined();
        expect(error.stack).toContain("JsonPatchError");
        expect(error.stack).toContain("Test error");
    });
});

describe("Schema Composition Tests", () => {
    test("JsonPatch can be part of larger schemas", () => {
        const RequestSchema = z.object({
            id: z.string(),
            patch: JsonPatchSchema,
            metadata: z.object({
                timestamp: z.number(),
                user: z.string()
            })
        });

        const validRequest = {
            id: "req-123",
            patch: [
                { op: "add", path: "/name", value: "John" },
                { op: "replace", path: "/age", value: 30 }
            ],
            metadata: {
                timestamp: Date.now(),
                user: "admin"
            }
        };

        const result = RequestSchema.safeParse(validRequest);
        expect(result.success).toBe(true);
    });

    test("JsonPatchOperation works in discriminated unions", () => {
        const ActionSchema = z.discriminatedUnion("type", [
            z.object({
                type: z.literal("patch"),
                operations: JsonPatchSchema
            }),
            z.object({
                type: z.literal("replace"),
                data: JsonValueSchema
            })
        ]);

        const patchAction = {
            type: "patch",
            operations: [
                { op: "add", path: "/test", value: "value" }
            ]
        };

        const result = ActionSchema.safeParse(patchAction);
        expect(result.success).toBe(true);
    });
});
