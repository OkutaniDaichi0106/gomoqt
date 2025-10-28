import { describe, it, assertEquals, assertInstanceOf, assertThrows, assertNotEquals } from "../deps.ts";
import { BytesFrame } from "./frame.ts";

describe("Frame", () => {
  it("reports byteLength correctly", () => {
    const data = new Uint8Array([1, 2, 3]);
    const f = new BytesFrame(data);
    assertEquals(f.byteLength, 3);
  });

  it("copyTo copies into Uint8Array", () => {
    const data = new Uint8Array([10, 20, 30]);
    const f = new BytesFrame(data);
    const dest = new Uint8Array(3);
    f.copyTo(dest);
    assertEquals(dest, data);
  });

  it("copyTo copies into ArrayBuffer", () => {
    const data = new Uint8Array([7, 8, 9]);
    const f = new BytesFrame(data);
    const destBuf = new ArrayBuffer(3);
    f.copyTo(destBuf);
    assertEquals(new Uint8Array(destBuf), data);
  });

  it("copyTo throws on unsupported dest type", () => {
    const data = new Uint8Array([1]);
    const f = new BytesFrame(data);
    // @ts-ignore - intentionally passing unsupported type
    assertThrows(() => f.copyTo(123), Error, "Unsupported destination type");
  });

  it("copyFrom copies from another Source", () => {
    const srcData = new Uint8Array([5, 6, 7]);
    const src = new BytesFrame(srcData);
    const dest = new BytesFrame(new Uint8Array(3));
    dest.copyFrom(src);
    assertEquals(dest.bytes, srcData);
  });

  it("clone creates a new Frame with copied data", () => {
    const data = new Uint8Array([1, 2, 3]);
    const f = new BytesFrame(data);
    const cloned = f.clone();
    assertInstanceOf(cloned, BytesFrame);
    assertEquals(cloned.bytes, data);
    assertNotEquals(cloned.bytes, data); // Ensure it's a copy (different reference)
  });
});
