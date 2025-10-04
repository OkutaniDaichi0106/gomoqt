import { z } from "zod"

// uint8: 0 to 255
export const uint8Schema = z.number().int().min(0).max(255);

export type uint8 = z.infer<typeof uint8Schema>;

export function uint8(value: number): uint8 {
  return uint8Schema.parse(value);
}

// uint53: Safe integer in JS, 0 to Number.MAX_SAFE_INTEGER
export const uint53Schema = z.number().int().min(0).max(Number.MAX_SAFE_INTEGER);

export type uint53 = z.infer<typeof uint53Schema>;

export function uint53(value: number): uint53 {
  return uint53Schema.parse(value);
}

// uint62: Union of uint53 and bigint up to 62 bits
export const uint62Schema = z.union([
  uint53Schema,
  z.bigint().min(0n).max(2n ** 62n - 1n)
]);

export type uint62 = z.infer<typeof uint62Schema>;

export function uint62(value: number | bigint): uint62 {
  return uint62Schema.parse(value);
}
