import type { JsonValue, JsonPatch } from "./json_patch";

export class JsonEncoder {
    #output: (chunk: EncodedJsonChunk, metadata?: EncodedJsonChunkMetadata) => void;
    #error: (error: Error) => void;
    #replacer?: (key: string, value: any) => any;
    #space?: string | number;
    #meta?: EncodedJsonChunkMetadata;
    #textEncoder: TextEncoder = new TextEncoder();
    #buffer: Uint8Array | undefined;

    constructor(init: JsonEncoderInit) {
        this.#output = init.output;
        this.#error = init.error;
    }

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

        this.#meta = {
            decoderConfig: {
                reviverRules: config.replacer,
            }
        }
    }

    encode(value: JsonValue | JsonPatch): void {
        // JSON Patch is always an array, regular JSON can be object, string, number, etc.
        const isJsonPatch = Array.isArray(value);
        const str = JSON.stringify(value, this.#replacer, this.#space);

        // Allocate or resize buffer efficiently
        if (!this.#buffer || this.#buffer.length < str.length) {
            const currentSize = this.#buffer?.length || 0;
            const newSize = Math.max(currentSize * 2, str.length, 1024);
            this.#buffer = new Uint8Array(newSize);
        }

        // Encode into the buffer
        const {written} = this.#textEncoder.encodeInto(str, this.#buffer);
        const chunk = new EncodedJsonChunk({
            type: isJsonPatch ? "delta" : "key",
            data: this.#buffer.subarray(0, written),
            timestamp: Date.now(),
        });
        this.#output(chunk, this.#meta);

        // Reset metadata
        if (this.#meta) {
            this.#meta = undefined;
        }
    }

    close(): void {
        this.#buffer = undefined;
    }
}

export interface JsonEncoderInit {
    output: (chunk: EncodedJsonChunk, metadata?: EncodedJsonChunkMetadata) => void;
    error: (error: Error) => void;
}

export interface JsonEncoderConfig {
    space?: string | number;
    replacer?: JsonRuleName[];
}

export class EncodedJsonChunk {
    type: "key" | "delta";
    data: Uint8Array;
    timestamp: number;

    constructor(init: EncodedJsonChunkInit) {
        this.type = init.type;
        this.data = init.data;
        this.timestamp = init.timestamp;
    }

    get byteLength() {
        return this.data.byteLength;
    }

    copyTo(target: Uint8Array) {
        target.set(this.data);
    }
}

export interface EncodedJsonChunkInit {
    type: "key" | "delta";
    data: Uint8Array;
    timestamp: number;
}

export interface EncodedJsonChunkMetadata {
    // space?: string | number;
    decoderConfig?: JsonDecoderConfig;
}

export class JsonDecoder {
    #output: (chunk: JsonValue | JsonPatch) => void;
    #error: (error: Error) => void;
    #reviver?: (key: string, value: any) => any;
    #textDecoder: TextDecoder = new TextDecoder();

    constructor(init: JsonDecoderInit) {
        this.#output = init.output;
        this.#error = init.error;
    }

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

    decode(chunk: EncodedJsonChunk): void {
        try {
            const jsonString = this.#textDecoder.decode(chunk.data);

            // Simple detection: JSON Patch always starts with '[', regular JSON can start with '{', '"', etc.
            // This is based on RFC 6902 - JSON Patch is always an array
            const parsed = JSON.parse(jsonString, this.#reviver);

            this.#output(parsed);
        } catch (error) {
            if (error instanceof Error) {
                this.#error(error);
            } else {
                this.#error(new Error(String(error)));
            }
        }
    }

    close(): void {
        // No resources to release
    }
}

export interface JsonDecoderInit {
    output: (chunk: JsonValue | JsonPatch) => void;
    error: (error: Error) => void;
    reviver?: (key: string, value: any) => any;
}

export interface JsonDecoderConfig {
    reviverRules?: JsonRuleName[];
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