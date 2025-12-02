import type { Reader, Writer } from "@okudai/golikejs/io";
import { readUint16, writeUint16 } from "./message.ts";

// deno-lint-ignore no-empty-interface
export interface SubscribeOkMessageInit {}

export class SubscribeOkMessage {
  constructor(_: SubscribeOkMessageInit = {}) {
  }

  /**
   * Returns the length of the message body (excluding the length prefix).
   */
  get len(): number {
    return 0;
  }

  /**
   * Encodes the message to the writer.
   */
  async encode(w: Writer): Promise<Error | undefined> {
    const [, err] = await writeUint16(w, this.len);
    return err;
  }

  /**
   * Decodes the message from the reader.
   */
  async decode(r: Reader): Promise<Error | undefined> {
    const [msgLen, , err] = await readUint16(r);
    if (err) return err;

    if (msgLen !== this.len) {
      return new Error(
        `message length mismatch: expected ${msgLen}, got ${this.len}`,
      );
    }

    return undefined;
  }
}
