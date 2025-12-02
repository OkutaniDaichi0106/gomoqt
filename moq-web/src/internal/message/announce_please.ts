import type { Reader, Writer } from "@okudai/golikejs/io";
import {
  parseString,
  readFull,
  readUint16,
  stringLen,
  writeString,
  writeUint16,
} from "./message.ts";

export interface AnnouncePleaseMessageInit {
  prefix?: string;
}

export class AnnouncePleaseMessage {
  prefix: string;

  constructor(init: AnnouncePleaseMessageInit = {}) {
    this.prefix = init.prefix ?? "";
  }

  /**
   * Returns the length of the message body (excluding the length prefix).
   */
  get len(): number {
    return stringLen(this.prefix);
  }

  /**
   * Encodes the message to the writer.
   */
  async encode(w: Writer): Promise<Error | undefined> {
    const msgLen = this.len;
    let err: Error | undefined;

    [, err] = await writeUint16(w, msgLen);
    if (err) return err;

    [, err] = await writeString(w, this.prefix);
    if (err) return err;

    return undefined;
  }

  /**
   * Decodes the message from the reader.
   */
  async decode(r: Reader): Promise<Error | undefined> {
    const [msgLen, , err1] = await readUint16(r);
    if (err1) return err1;

    const buf = new Uint8Array(msgLen);
    const [, err2] = await readFull(r, buf);
    if (err2) return err2;

    [this.prefix] = parseString(buf, 0);

    return undefined;
  }
}
