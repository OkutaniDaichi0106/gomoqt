import { CancelCauseFunc, Context, withCancelCause } from "./internal/context";
import { Reader, Writer, Source } from "./io";
import { StreamError } from "./io/error";
import { GroupMessage } from "./message";
import { Frame } from "./frame";

export class GroupWriter {
    #group: GroupMessage;
    #writer: Writer;
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;

    constructor(trackCtx: Context, writer: Writer, group: GroupMessage) {
        this.#group = group;
        this.#writer = writer;
        [this.#ctx, this.#cancelFunc] = withCancelCause(trackCtx);

        (async () => {
            // Wait for the writer to close
            await trackCtx.done();
            const err = trackCtx.err();
            if (err) {
                // If the context is cancelled, cancel the stream with the error
                this.cancel(0, err.message); // TODO: Use a more descriptive message
                return;
            } else {
                // If the context is not cancelled, cancel the stream with a code of 0 (indicating normal closure)
                this.cancel(0, "normal closure");
            }
        })();
    }

    get groupSequence(): bigint {
        return this.#group.sequence;
    }

    async writeFrame(src: Frame | Source): Promise<Error | undefined> {
        if (src instanceof Frame) {
            this.#writer.writeUint8Array(src.bytes);
        } else {
            this.#writer.copyFrom(src);
        }
        const err = await this.#writer.flush();
        if (err) {
            console.error("Error writing frame:", err);
            return err;
        }

        return undefined;
    }

    close(): void {
        this.#writer.close();
        this.#cancelFunc(new Error("Stream closed")); // Notify the context about closure
    }

    cancel(code: number, message: string): void {
        const err = new StreamError(code, message);
        this.#writer.cancel(err);
        this.#cancelFunc(err); // Notify the context about cancellation
    }

    get context(): Context {
        return this.#ctx;
    }
}

export class GroupReader {
    #group: GroupMessage;
    #reader: Reader;
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;
    #frame?: Frame;

    constructor(trackCtx: Context, reader: Reader, group: GroupMessage) {
        this.#group = group;
        this.#reader = reader;
        [this.#ctx, this.#cancelFunc] = withCancelCause(trackCtx);

        (async () => {
            await trackCtx.done();
            this.cancel(0);
        })();
    }

    get groupSequence(): bigint {
        return this.#group.sequence;
    }

    async readFrame(): Promise<[Frame?, Error?]> {
        let err: Error | undefined;
        let len: number;
        [len, err] = await this.#reader.readVarint();
        if (err) {
            return [undefined, err];
        }

        if (len > Number.MAX_SAFE_INTEGER) {
            return [undefined, new Error("Varint too large")];
        }

        if (!this.#frame || this.#frame.bytes.byteLength < len) {
            const currentSize = this.#frame?.bytes.byteLength || 0;
            const cap = Math.max(currentSize * 2, len);
            // Swap buffers
            this.#frame = new Frame(new Uint8Array(cap));
        }

        err = await this.#reader.fillN(this.#frame.bytes, len);
        if (err) {
            return [undefined, err];
        }

        return [this.#frame, undefined];
    }

    cancel(code: number): void {
        const reason = new StreamError(code, "Stream cancelled");
        this.#reader.cancel(reason);
        this.#cancelFunc(reason);
    }

    get context(): Context {
        return this.#ctx;
    }
}