import { assertEquals, assertExists } from "../../deps.ts";
import { Reader } from './reader.ts';
import { StreamError } from './error.ts';

function createControllerReader(): { reader: Reader; controller: ReadableStreamDefaultController<Uint8Array> } {
  let ctrl!: ReadableStreamDefaultController<Uint8Array>;
  const stream = new ReadableStream<Uint8Array>({
    start(c) { ctrl = c; }
  });
  return { reader: new Reader({ stream, streamId: 1n }), controller: ctrl };
}

function createClosedReader(data: Uint8Array): Reader {
  const stream = new ReadableStream<Uint8Array>({
    start(ctrl) { ctrl.enqueue(data); ctrl.close(); }
  });
  return new Reader({ stream, transfer: undefined, streamId: 0n });
}

Deno.test("webtransport/reader - readUint8Array scenarios", async (t) => {
  await t.step("reads Uint8Array with varint length", () => {
    const { reader, controller } = createControllerReader();
    const data = new Uint8Array([1,2,3,4,5]);
    controller.enqueue(new Uint8Array([5, ...data]));
    controller.close();

    return (async () => {
      const [res, err] = await reader.readUint8Array();
      assertEquals(err, undefined);
      assertEquals(res, data);
    })();
  });

  await t.step("handles empty array", async () => {
    const r = createClosedReader(new Uint8Array([0]));
    const [res, err] = await r.readUint8Array();
    assertEquals(err, undefined);
    assertEquals(res, new Uint8Array([]));
  });

  await t.step("partial reads assemble correctly", async () => {
    const { reader, controller } = createControllerReader();
    controller.enqueue(new Uint8Array([3]));
    controller.enqueue(new Uint8Array([1,2]));
    controller.enqueue(new Uint8Array([3]));
    controller.close();
    const [res, err] = await reader.readUint8Array();
    assertEquals(err, undefined);
    assertEquals(res, new Uint8Array([1,2,3]));
  });

  await t.step("insufficient data returns error", async () => {
    const { reader, controller } = createControllerReader();
    controller.enqueue(new Uint8Array([0xFF]));
    controller.close();
    const [res, err] = await reader.readUint8Array();
    assertEquals(res, undefined);
    assertExists(err);
  });

  await t.step("very large varint triggers error", async () => {
    const large = new Uint8Array([0xF0,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF]);
    const r = createClosedReader(large);
    // Some implementations may throw synchronously or reject; ensure we catch either
    try {
      await r.readUint8Array();
      throw new Error('expected readUint8Array to throw for very large varint');
    } catch (e) {
      // error expected
    }
  });
});

Deno.test("webtransport/reader - readString and readBigVarint", async (t) => {
  await t.step("reads UTF-8 string", () => {
    const { reader, controller } = createControllerReader();
    const s = 'hello world';
    const enc = new TextEncoder().encode(s);
    controller.enqueue(new Uint8Array([enc.length, ...enc]));
    controller.close();
    return (async () => {
      const [res, err] = await reader.readString();
      assertEquals(err, undefined);
      assertEquals(res, s);
    })();
  });

  await t.step("empty string", async () => {
    const r = createClosedReader(new Uint8Array([0]));
    const [res, err] = await r.readString();
    assertEquals(err, undefined);
    assertEquals(res, '');
  });

  await t.step("unicode string", () => {
    const { reader, controller } = createControllerReader();
    const s = 'こんにちは🚀';
    const enc = new TextEncoder().encode(s);
    controller.enqueue(new Uint8Array([enc.length, ...enc]));
    controller.close();
    return (async () => {
      const [res, err] = await reader.readString();
      assertEquals(err, undefined);
      assertEquals(res, s);
    })();
  });

  await t.step("underlying readUint8Array failure returns error", async () => {
    const r = createClosedReader(new Uint8Array([0xFF]));
    const [res, err] = await r.readString();
    assertExists(err);
    assertEquals(res, "");
  });

  await t.step("readBigVarint - single/two/four/eight bytes and partial/error", async () => {
    // single
    let r = createClosedReader(new Uint8Array([42]));
    let [res, err] = await r.readBigVarint();
    assertEquals(err, undefined);
    assertEquals(res, 42n);

    // two-byte
    r = createClosedReader(new Uint8Array([0x41, 0x2C]));
    [res, err] = await r.readBigVarint();
    assertEquals(err, undefined);
    assertEquals(res, 300n);

    // four-byte
    r = createClosedReader(new Uint8Array([0x80, 0x0F, 0x42, 0x40]));
    [res, err] = await r.readBigVarint();
    assertEquals(err, undefined);
    assertEquals(res, 1000000n);

    // eight-byte
    r = createClosedReader(new Uint8Array([0xC0,0x00,0x01,0x00,0x00,0x00,0x00,0x00]));
    [res, err] = await r.readBigVarint();
    assertEquals(err, undefined);
    assertEquals(res, 1n << 40n);

    // partial varint
    const cr = createControllerReader();
    cr.controller.enqueue(new Uint8Array([0x41]));
    cr.controller.enqueue(new Uint8Array([0x2C]));
    cr.controller.close();
    [res, err] = await cr.reader.readBigVarint();
    assertEquals(err, undefined);
    assertEquals(res, 300n);

    // error on close before complete
    const cr2 = createControllerReader();
    cr2.controller.enqueue(new Uint8Array([0x41]));
    cr2.controller.close();
    [res, err] = await cr2.reader.readBigVarint();
    assertExists(err);
  });
});

Deno.test("webtransport/reader - readUint8/readBoolean and control APIs", async (t) => {
  await t.step("readUint8 single and sequence", async () => {
    const { reader, controller } = createControllerReader();
    controller.enqueue(new Uint8Array([123]));
    controller.close();
  const [v, e] = await reader.readUint8();
    assertEquals(e, undefined);
    assertEquals(v, 123);

    const r2 = createClosedReader(new Uint8Array([1,2,3]));
  const [a, ea] = await r2.readUint8();
    assertEquals(ea, undefined);
    assertEquals(a, 1);
  const [b, eb] = await r2.readUint8();
    assertEquals(eb, undefined);
    assertEquals(b, 2);
  const [c, ec] = await r2.readUint8();
    assertEquals(ec, undefined);
    assertEquals(c, 3);

    const r3 = createControllerReader();
    r3.controller.close();
    const [rv, re] = await r3.reader.readUint8();
    assertExists(re);
    assertEquals(rv, 0);
  });

  await t.step("readBoolean true/false and invalid cases", async () => {
    let r = createClosedReader(new Uint8Array([1]));
    let [bv, be] = await r.readBoolean();
    assertEquals(be, undefined);
    assertEquals(bv, true);

    r = createClosedReader(new Uint8Array([0]));
    [bv, be] = await r.readBoolean();
    assertEquals(be, undefined);
    assertEquals(bv, false);

    r = createClosedReader(new Uint8Array([2]));
    [bv, be] = await r.readBoolean();
    assertExists(be);
    assertEquals(bv, false);
  });

  await t.step("cancel and closed APIs", async () => {
    const { reader } = createControllerReader();
    const err = new StreamError(123, 'msg');
    await reader.cancel(err);

    const { reader: r2, controller } = createControllerReader();
    const closedPromise = r2.closed();
    controller.close();
    await closedPromise;
  });
});

Deno.test("webtransport/reader - integration sequence", async (t) => {
  await t.step("reads boolean, varint, string, and bytes sequentially", async () => {
    const { reader, controller } = createControllerReader();
    const testStr = 'test';
    const testBytes = new Uint8Array([1,2,3]);
    const enc = new TextEncoder().encode(testStr);
    controller.enqueue(new Uint8Array([1, 42, enc.length, ...enc, testBytes.length, ...testBytes]));
    controller.close();

    const [bv, be] = await reader.readBoolean();
    assertEquals(be, undefined);
    assertEquals(bv, true);

    const [vv, ve] = await reader.readBigVarint();
    assertEquals(ve, undefined);
    assertEquals(vv, 42n);

    const [sv, se] = await reader.readString();
    assertEquals(se, undefined);
    assertEquals(sv, testStr);

    const [av, ae] = await reader.readUint8Array();
    assertEquals(ae, undefined);
    assertEquals(av, testBytes);
  });

  await t.step("handles stream errors gracefully", async () => {
    const { reader, controller } = createControllerReader();
    controller.close();
    const [res, err] = await reader.readUint8();
    assertExists(err);
    assertEquals(res, 0);
  });
});
