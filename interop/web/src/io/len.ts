export const MAX_VARINT1: bigint = (1n << 6n) - 1n; // 63
export const MAX_VARINT2: bigint = (1n << 14n) - 1n; // 16383
export const MAX_VARINT4: bigint = (1n << 30n) - 1n; // 1073741823
export const MAX_VARINT8: bigint = (1n << 62n) - 1n; // 4611686018427387903

export function varintLen(value: number | bigint): number {
    // Handle negative values by converting to unsigned
    if (value < 0) {
        value = BigInt(value) + (1n << 64n);
    } else {
        value = BigInt(value);
    }

    if (value <= MAX_VARINT1) {
        return 1;
    } else if (value <= MAX_VARINT2) {
        return 2;
    } else if (value <= MAX_VARINT4) {
        return 4;
    } else if (value <= MAX_VARINT8) {
        return 8;
    }

    throw new RangeError("Value exceeds maximum varint size");
}

export function stringLen(str: string): number {
    let len = 0;
    len += varintLen(str.length);
    len += str.length;
    return len;
}

export function bytesLen(bytes: Uint8Array): number {
    let len = 0;
    len += varintLen(bytes.length);
    len += bytes.length;
    return len;
}