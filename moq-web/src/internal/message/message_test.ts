import { assertEquals } from "@std/assert";
import { spy } from "@std/testing/mock";
import {
  MAX_BYTES_LENGTH,
  parseBytes,
  parseString,
  parseStringArray,
  parseVarint,
  readBytes,
  readFull,
  readString,
  readStringArray,
  readVarint,
  writeBytes,
  writeString,
  writeStringArray,
  writeVarint,
} from "./message.ts";
import { EOFError } from "@okudai/golikejs/io";

Deno.test("message utilities", async (t) => {
  await t.step("readFull - reads exactly the requested bytes", async () => {
    const data = new Uint8Array([1, 2, 3, 4, 5]);
    let offset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (offset >= data.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, data.length - offset);
        p.set(data.subarray(offset, offset + n));
        offset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const result = new Uint8Array(3);
    const [n, err] = await readFull(reader, result);
    assertEquals(n, 3);
    assertEquals(err, undefined);
    assertEquals(result, new Uint8Array([1, 2, 3]));
  });

  await t.step("readFull - returns EOF when not enough data", async () => {
    const data = new Uint8Array([1, 2]);
    let offset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (offset >= data.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, data.length - offset);
        p.set(data.subarray(offset, offset + n));
        offset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const result = new Uint8Array(5);
    const [n, err] = await readFull(reader, result);
    assertEquals(n, 2);
    assertEquals(err?.name, "EOFError");
  });

  await t.step("writeVarint - writes single byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const [n, err] = await writeVarint(writer, 42);
    assertEquals(n, 1);
    assertEquals(err, undefined);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    assertEquals(allData, new Uint8Array([42]));
  });

  await t.step("writeVarint - writes two byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const [n, err] = await writeVarint(writer, 1000);
    assertEquals(n, 2);
    assertEquals(err, undefined);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    assertEquals(allData, new Uint8Array([0x43, 0xe8]));
  });

  await t.step("writeVarint - writes four byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const [n, err] = await writeVarint(writer, 100000);
    assertEquals(n, 4);
    assertEquals(err, undefined);
  });

  await t.step("writeVarint - writes eight byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const [n, err] = await writeVarint(writer, 4294967296);
    assertEquals(n, 8);
    assertEquals(err, undefined);
  });

  await t.step("writeVarint - rejects negative numbers", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const [n, err] = await writeVarint(writer, -1);
    assertEquals(n, 0);
    assertEquals(err?.message, "Varint cannot be negative");
  });

  await t.step("writeVarint - rejects too large numbers", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const [n, err] = await writeVarint(writer, Infinity);
    assertEquals(n, 0);
    assertEquals(err?.name, "RangeError");
  });

  await t.step("writeBytes - writes bytes with length prefix", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const data = new Uint8Array([1, 2, 3]);
    const [n, err] = await writeBytes(writer, data);
    assertEquals(n, 4);
    assertEquals(err, undefined);
  });

  await t.step("writeString - writes string with length prefix", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const [n, err] = await writeString(writer, "hello");
    assertEquals(n, 6);
    assertEquals(err, undefined);
  });

  await t.step("writeStringArray - writes string array", async () => {
    const writtenData: Uint8Array[] = [];
    const writer = {
      id: 0n,
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const arr = ["hello", "world"];
    const [n, err] = await writeStringArray(writer, arr);
    assertEquals(err, undefined);
    assertEquals(n > 0, true);
  });

  await t.step("readVarint - reads single byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeVarint(writerMock, 42);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [value, n, err] = await readVarint(reader);
    assertEquals(value, 42);
    assertEquals(n, 1);
    assertEquals(err, undefined);
  });

  await t.step("readVarint - reads two byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeVarint(writerMock, 1000);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [value, n, err] = await readVarint(reader);
    assertEquals(value, 1000);
    assertEquals(n, 2);
    assertEquals(err, undefined);
  });

  await t.step("readVarint - reads four byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeVarint(writerMock, 100000);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [value, n, err] = await readVarint(reader);
    assertEquals(value, 100000);
    assertEquals(n, 4);
    assertEquals(err, undefined);
  });

  await t.step("readVarint - reads eight byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeVarint(writerMock, 4294967296);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [value, n, err] = await readVarint(reader);
    assertEquals(value, 4294967296);
    assertEquals(n, 8);
    assertEquals(err, undefined);
  });

  await t.step("readBytes - reads bytes with length prefix", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const data = new Uint8Array([1, 2, 3]);
    await writeBytes(writerMock, data);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [result, n, err] = await readBytes(reader);
    assertEquals(result, data);
    assertEquals(n, 4);
    assertEquals(err, undefined);
  });

  await t.step("readBytes - rejects too large length", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeVarint(writerMock, MAX_BYTES_LENGTH + 1);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [result, _n, err] = await readBytes(reader);
    assertEquals(result.length, 0);
    assertEquals(err?.message, "Bytes length exceeds maximum limit");
  });

  await t.step("readString - reads string with length prefix", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeString(writerMock, "hello");
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [result, n, err] = await readString(reader);
    assertEquals(result, "hello");
    assertEquals(n, 6);
    assertEquals(err, undefined);
  });

  await t.step("readStringArray - reads string array", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    const arr = ["hello", "world"];
    await writeStringArray(writerMock, arr);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [result, _n, err] = await readStringArray(reader);
    assertEquals(result, arr);
    assertEquals(err, undefined);
  });

  await t.step("readStringArray - rejects too large count", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeVarint(writerMock, MAX_BYTES_LENGTH + 1);
    const allData = new Uint8Array(
      writtenData.reduce((a, b) => a + b.length, 0),
    );
    let off = 0;
    for (const d of writtenData) {
      allData.set(d, off);
      off += d.length;
    }
    let readOffset = 0;
    const reader = {
      id: 0n,
      read: spy(async (p: Uint8Array) => {
        if (readOffset >= allData.length) {
          return [0, new EOFError()] as [number, Error | undefined];
        }
        const n = Math.min(p.length, allData.length - readOffset);
        p.set(allData.subarray(readOffset, readOffset + n));
        readOffset += n;
        return [n, undefined] as [number, Error | undefined];
      }),
      cancel: spy(async (_code: number) => {}),
      closed: () => new Promise<void>(() => {}),
    };

    const [result, _n2, err] = await readStringArray(reader);
    assertEquals(result.length, 0);
    assertEquals(err?.message, "String array count exceeds maximum limit");
  });

  await t.step("parseVarint - parses single byte varint", () => {
    const buf = new Uint8Array([42]);
    const [value, n] = parseVarint(buf, 0);
    assertEquals(value, 42);
    assertEquals(n, 1);
  });

  await t.step("parseVarint - parses two byte varint", () => {
    const buf = new Uint8Array([0x43, 0xe8]);
    const [value, n] = parseVarint(buf, 0);
    assertEquals(value, 1000);
    assertEquals(n, 2);
  });

  await t.step("parseVarint - parses four byte varint", () => {
    const buf = new Uint8Array([0x80, 0x01, 0x86, 0xa0]);
    const [value, n] = parseVarint(buf, 0);
    assertEquals(value, 100000);
    assertEquals(n, 4);
  });

  await t.step("parseVarint - parses eight byte varint", async () => {
    const writtenData: Uint8Array[] = [];
    const writerMock = {
      write: spy(async (p: Uint8Array) => {
        writtenData.push(new Uint8Array(p));
        return [p.length, undefined] as [number, Error | undefined];
      }),
    };
    await writeVarint(writerMock, 4294967296);
    const buf = new Uint8Array(writtenData.reduce((a, b) => a + b.length, 0));
    let off = 0;
    for (const d of writtenData) {
      buf.set(d, off);
      off += d.length;
    }
    const [value, n] = parseVarint(buf, 0);
    assertEquals(value, 4294967296);
    assertEquals(n, 8);
  });

  await t.step("parseBytes - parses bytes with length prefix", () => {
    const data = new Uint8Array([3, 1, 2, 3]);
    const [result, n] = parseBytes(data, 0);
    assertEquals(result, new Uint8Array([1, 2, 3]));
    assertEquals(n, 4);
  });

  await t.step("parseString - parses string with length prefix", () => {
    const data = new Uint8Array([5, 104, 101, 108, 108, 111]);
    const [result, n] = parseString(data, 0);
    assertEquals(result, "hello");
    assertEquals(n, 6);
  });

  await t.step("parseStringArray - parses string array", () => {
    const data = new Uint8Array([
      2,
      5,
      104,
      101,
      108,
      108,
      111,
      5,
      119,
      111,
      114,
      108,
      100,
    ]);
    const [result, n] = parseStringArray(data, 0);
    assertEquals(result, ["hello", "world"]);
    assertEquals(n, 13);
  });
});
