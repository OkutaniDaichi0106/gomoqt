/**
 * Message encoding/decoding utilities.
 * Go-like design with io.Reader/io.Writer interfaces.
 */

import type { Reader, Writer } from "@okudai/golikejs/io";
import { EOFError } from "@okudai/golikejs/io";
import {
  bytesLen,
  MAX_VARINT1,
  MAX_VARINT2,
  MAX_VARINT4,
  MAX_VARINT8,
  stringLen,
  varintLen,
} from "../webtransport/len.ts";

export { bytesLen, stringLen, varintLen };

// Maximum bytes length (1 GiB)
export const MAX_BYTES_LENGTH = 1 << 30;

/**
 * Reads exactly len(p) bytes from r into p.
 * Like Go's io.ReadFull.
 */
export async function readFull(
  r: Reader,
  p: Uint8Array,
): Promise<[number, Error | undefined]> {
  let totalRead = 0;
  while (totalRead < p.length) {
    const [n, err] = await r.read(p.subarray(totalRead));
    totalRead += n;
    if (err) {
      return [totalRead, err];
    }
    if (n === 0) {
      return [totalRead, new EOFError()];
    }
  }
  return [totalRead, undefined];
}

/**
 * Writes a varint to the writer.
 * Returns number of bytes written and any error.
 */
export async function writeVarint(
  w: Writer,
  num: number,
): Promise<[number, Error | undefined]> {
  if (num < 0) {
    return [0, new Error("Varint cannot be negative")];
  }

  let buf: Uint8Array;
  if (num <= MAX_VARINT1) {
    buf = new Uint8Array([num]);
  } else if (num <= MAX_VARINT2) {
    buf = new Uint8Array(2);
    buf[0] = (num >> 8) | 0x40;
    buf[1] = num & 0xff;
  } else if (num <= MAX_VARINT4) {
    buf = new Uint8Array(4);
    buf[0] = (num >> 24) | 0x80;
    buf[1] = (num >> 16) & 0xff;
    buf[2] = (num >> 8) & 0xff;
    buf[3] = num & 0xff;
  } else if (num <= Number(MAX_VARINT8)) {
    buf = new Uint8Array(8);
    const bn = BigInt(num);
    buf[0] = Number((bn >> 56n) | 0xc0n);
    buf[1] = Number((bn >> 48n) & 0xffn);
    buf[2] = Number((bn >> 40n) & 0xffn);
    buf[3] = Number((bn >> 32n) & 0xffn);
    buf[4] = Number((bn >> 24n) & 0xffn);
    buf[5] = Number((bn >> 16n) & 0xffn);
    buf[6] = Number((bn >> 8n) & 0xffn);
    buf[7] = Number(bn & 0xffn);
  } else {
    return [0, new RangeError("Value exceeds maximum varint size")];
  }

  return await w.write(buf);
}

/**
 * Writes bytes with a varint length prefix to the writer.
 * Returns number of bytes written and any error.
 */
export async function writeBytes(
  w: Writer,
  data: Uint8Array,
): Promise<[number, Error | undefined]> {
  const [n1, err1] = await writeVarint(w, data.length);
  if (err1) {
    return [n1, err1];
  }
  const [n2, err2] = await w.write(data);
  return [n1 + n2, err2];
}

/**
 * Writes a string with a varint length prefix to the writer.
 * Returns number of bytes written and any error.
 */
export async function writeString(
  w: Writer,
  str: string,
): Promise<[number, Error | undefined]> {
  const encoder = new TextEncoder();
  const data = encoder.encode(str);
  return await writeBytes(w, data);
}

/**
 * Writes a string array to the writer.
 * Returns number of bytes written and any error.
 */
export async function writeStringArray(
  w: Writer,
  arr: string[],
): Promise<[number, Error | undefined]> {
  let total = 0;
  const [n1, err1] = await writeVarint(w, arr.length);
  if (err1) {
    return [n1, err1];
  }
  total += n1;

  for (const str of arr) {
    const [n, err] = await writeString(w, str);
    if (err) {
      return [total + n, err];
    }
    total += n;
  }

  return [total, undefined];
}

/**
 * Reads a varint from the reader.
 * Returns the value, number of bytes read, and any error.
 */
export async function readVarint(
  r: Reader,
): Promise<[number, number, Error | undefined]> {
  const firstByte = new Uint8Array(1);
  const [n, err] = await readFull(r, firstByte);
  if (err) {
    return [0, n, err];
  }

  const len = 1 << (firstByte[0]! >> 6);
  let value = firstByte[0]! & 0x3f;

  if (len === 1) {
    return [value, 1, undefined];
  }

  const remaining = new Uint8Array(len - 1);
  const [n2, err2] = await readFull(r, remaining);
  if (err2) {
    return [0, 1 + n2, err2];
  }

  for (let i = 0; i < len - 1; i++) {
    value = value * 256 + remaining[i]!;
  }

  return [value, len, undefined];
}

/**
 * Reads bytes with a varint length prefix from the reader.
 * Returns the bytes, number of bytes read, and any error.
 */
export async function readBytes(
  r: Reader,
): Promise<[Uint8Array, number, Error | undefined]> {
  const [len, n1, err1] = await readVarint(r);
  if (err1) {
    return [new Uint8Array(0), n1, err1];
  }

  if (len > MAX_BYTES_LENGTH) {
    return [
      new Uint8Array(0),
      n1,
      new Error("Bytes length exceeds maximum limit"),
    ];
  }

  const data = new Uint8Array(len);
  const [n2, err2] = await readFull(r, data);
  if (err2) {
    return [new Uint8Array(0), n1 + n2, err2];
  }

  return [data, n1 + n2, undefined];
}

/**
 * Reads a string with a varint length prefix from the reader.
 * Returns the string, number of bytes read, and any error.
 */
export async function readString(
  r: Reader,
): Promise<[string, number, Error | undefined]> {
  const [bytes, n, err] = await readBytes(r);
  if (err) {
    return ["", n, err];
  }
  const str = new TextDecoder().decode(bytes);
  return [str, n, undefined];
}

/**
 * Reads a string array from the reader.
 * Returns the array, number of bytes read, and any error.
 */
export async function readStringArray(
  r: Reader,
): Promise<[string[], number, Error | undefined]> {
  const [count, n1, err1] = await readVarint(r);
  if (err1) {
    return [[], n1, err1];
  }

  if (count > MAX_BYTES_LENGTH) {
    return [[], n1, new Error("String array count exceeds maximum limit")];
  }

  let total = n1;
  const arr: string[] = [];

  for (let i = 0; i < count; i++) {
    const [str, n, err] = await readString(r);
    if (err) {
      return [[], total + n, err];
    }
    arr.push(str);
    total += n;
  }

  return [arr, total, undefined];
}

// ============================================================
// Byte array parsing utilities (for parsing message body from buffer)
// ============================================================

/**
 * Parses a varint from a byte array at the given offset.
 * Returns [value, bytesRead].
 */
export function parseVarint(buf: Uint8Array, offset: number): [number, number] {
  const firstByte = buf[offset]!;
  const len = 1 << (firstByte >> 6);
  let value = firstByte & 0x3f;
  for (let i = 1; i < len; i++) {
    value = value * 256 + buf[offset + i]!;
  }
  return [value, len];
}

/**
 * Parses bytes with a varint length prefix from a byte array.
 * Returns [bytes, bytesRead].
 */
export function parseBytes(
  buf: Uint8Array,
  offset: number,
): [Uint8Array, number] {
  const [len, n] = parseVarint(buf, offset);
  const bytes = buf.subarray(offset + n, offset + n + len);
  return [bytes, n + len];
}

/**
 * Parses a string with a varint length prefix from a byte array.
 * Returns [string, bytesRead].
 */
export function parseString(buf: Uint8Array, offset: number): [string, number] {
  const [bytes, n] = parseBytes(buf, offset);
  const str = new TextDecoder().decode(bytes);
  return [str, n];
}

/**
 * Parses a string array from a byte array.
 * Returns [array, bytesRead].
 */
export function parseStringArray(
  buf: Uint8Array,
  offset: number,
): [string[], number] {
  const [count, n1] = parseVarint(buf, offset);
  let total = n1;
  const arr: string[] = [];
  for (let i = 0; i < count; i++) {
    const [str, n] = parseString(buf, offset + total);
    arr.push(str);
    total += n;
  }
  return [arr, total];
}
