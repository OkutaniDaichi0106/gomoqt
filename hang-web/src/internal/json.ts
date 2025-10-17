/**
 * JSON Patch implementation based on RFC 6902
 * Provides types and utilities for JSON patch operations using Zod schemas
 */

import { z } from 'zod';
import type { EncodedChunk } from ".";

/**
 * JSON object schema - a record of string keys to JSON values
 */
const JsonObjectSchema: z.ZodSchema<Record<string, any>> = z.lazy(() => z.record(z.string(), JsonValueSchema));

/**
 * JSON primitive schema - basic JSON values plus extended types for revivers
 */
const JsonPrimitiveSchema = z.union([
    z.string(),
    z.number(),
    z.boolean(),
    z.null(),
    z.bigint(),  // For BigInt reviver
    z.date()     // For Date reviver
]);

/**
 * JSON array schema - array of JSON values
 */
const JsonArraySchema: z.ZodSchema<any[]> = z.lazy(() => z.array(JsonValueSchema));

/**
 * JSON value schema that can be used in patches
 */
const JsonValueSchema: z.ZodSchema<any> = z.lazy(() =>
    z.union([
        JsonPrimitiveSchema,
        JsonObjectSchema,
        JsonArraySchema
    ])
);

// Export schemas
export {
    JsonValueSchema,
    JsonObjectSchema,
    JsonPrimitiveSchema,
    JsonArraySchema,
};

// Export types inferred from schemas
export type JsonValue = JsonPrimitive | JsonObject | JsonArray;
export type JsonObject = z.infer<typeof JsonObjectSchema>;
export type JsonPrimitive = z.infer<typeof JsonPrimitiveSchema>;
export type JsonArray = z.infer<typeof JsonArraySchema>;

export class JsonEncoder {
    #replacer?: (key: string, value: any) => any;
    #space?: string | number;
    #textEncoder: TextEncoder = new TextEncoder();

    constructor() {}

    configure(config: JsonEncoderConfig): void {
        this.#space = config.space ?? this.#space;

        if (config.replacer) {
            const rules = config.replacer;

            // Combine multiple replacers
            this.#replacer = (key: string, value: any) => {
                let result = value;
                for (const ruleName of rules) {
                    const rule = JSON_RULES[ruleName];
                    if (rule) {
                        result = rule.replacer(key, result);
                    }
                }
                return result;
            };
        }
    }

    encode(values: JsonValue[]): EncodedJsonChunk {
        const str = JSON.stringify(values, this.#replacer, this.#space);

        const chunk = new EncodedJsonChunk({
            type: "json",
            data: this.#textEncoder.encode(str),
        });

        return chunk;
    }
}

export interface JsonEncoderConfig {
    space?: string | number;
    replacer?: JsonRuleName[];
}

export class JsonLineEncoder {
    #replacer?: (key: string, value: any) => any;
    #space?: string | number;
    #textEncoder: TextEncoder = new TextEncoder();

    constructor() {}

    configure(config: JsonEncoderConfig): void {
        this.#space = config.space ?? this.#space;

        if (config.replacer) {
            const rules = config.replacer;

            // Combine multiple replacers
            this.#replacer = (key: string, value: any) => {
                let result = value;
                for (const ruleName of rules) {
                    const rule = JSON_RULES[ruleName];
                    if (rule) {
                        result = rule.replacer(key, result);
                    }
                }
                return result;
            };
        }
    }

    encode(values: JsonValue[]): EncodedJsonChunk {
        const lines = values.map(value => JSON.stringify(value, this.#replacer, this.#space));
        const str = lines.join('\n');

        const chunk = new EncodedJsonChunk({
            type: "jsonl",
            data: this.#textEncoder.encode(str),
        });

        return chunk;
    }
}

export class EncodedJsonChunk implements EncodedChunk {
    readonly type: "json" | "jsonl";
    data: Uint8Array;

    constructor(init: EncodedJsonChunkInit) {
        this.type = init.type;
        this.data = init.data;
    }

    get byteLength() {
        return this.data.byteLength;
    }

    copyTo(target: Uint8Array) {
        target.set(this.data);
    }
}

export interface EncodedJsonChunkInit {
    type: "json" | "jsonl";
    data: Uint8Array;
}

export class JsonDecoder {
    #reviver?: (key: string, value: any) => any;
    #textDecoder: TextDecoder = new TextDecoder();

    constructor() {}

    configure(config: JsonDecoderConfig): void {
        if (config.reviverRules) {
            const rules = config.reviverRules;
            // Combine multiple revivers
            this.#reviver = (key: string, value: any) => {
                let result = value;
                for (const ruleName of rules) {
                    const rule = JSON_RULES[ruleName];
                    if (rule) {
                        result = rule.reviver(key, result);
                    }
                }
                return result;
            };
        }
    }

    decode(chunk: EncodedJsonChunk): JsonArray {
        const text = this.#textDecoder.decode(chunk.data);
        const json = JSON.parse(text, this.#reviver);

        const { success, data } = JsonArraySchema.safeParse(json);
        if (success) {
            return data;
        }

        throw new Error("Decoded JSON is not a valid JsonArray");
    }
}

export interface JsonDecoderConfig {
    reviverRules?: JsonRuleName[];
}

export class JsonLineDecoder {
    #reviver?: (key: string, value: any) => any;
    #textDecoder: TextDecoder = new TextDecoder();

    constructor() {}

    configure(config: JsonDecoderConfig): void {
        if (config.reviverRules) {
            const rules = config.reviverRules;
            // Combine multiple revivers
            this.#reviver = (key: string, value: any) => {
                let result = value;
                for (const ruleName of rules) {
                    const rule = JSON_RULES[ruleName];
                    if (rule) {
                        result = rule.reviver(key, result);
                    }
                }
                return result;
            };
        }
    }

    decode(chunk: EncodedJsonChunk): JsonValue[] {
        if (chunk.type !== "jsonl") {
            throw new Error("Invalid chunk type");
        }

        const text = this.#textDecoder.decode(chunk.data);
        const lines = text.split('\n').filter(line => line.trim());
        if (lines.length === 0) throw new Error("No JSON lines found");
        
        const values: JsonValue[] = [];
        for (const line of lines) {
            const json = JSON.parse(line, this.#reviver);
            const { success, data } = JsonValueSchema.safeParse(json);
            if (success) values.push(data);
            else throw new Error("Decoded JSON line is not a valid JsonValue");
        }
        return values;
    }
}

export function replaceBigInt(key: string, value: any): any {
	if (typeof value === "bigint") {
		return value.toString();
	}
	return value;
}

export function reviveBigInt(key: string, value: any): any {
	if (typeof value === "string" && /^\d+$/.test(value)) {
		try {
			return BigInt(value);
		} catch {
			return value;
		}
	}
	return value;
}

export function replaceDate(key: string, value: any): any {
	if (value instanceof Date) {
		return value.toISOString();
	}
	return value;
}

export function reviveDate(key: string, value: any): any {
	if (typeof value === "string" && /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z$/.test(value)) {
		try {
			return new Date(value);
		} catch {
			return value;
		}
	}
	return value;
}

// Rule definitions
export const JSON_RULES = {
    bigint: {
        replacer: replaceBigInt,
        reviver: reviveBigInt,
    },
    date: {
        replacer: replaceDate,
        reviver: reviveDate,
    },
    // Add more rules as needed
} as const;

export type JsonRuleName = keyof typeof JSON_RULES;