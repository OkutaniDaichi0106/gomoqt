import { CancelCauseFunc, Context, withCancelCause } from "./internal/context";
import { Reader, Writer } from "./io";
import { StreamError } from "./io/error";
import { GroupMessage } from "./message";

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

    async write(data: Uint8Array): Promise<Error | undefined> {
        this.#writer.writeUint8Array(data);
        const err = await this.#writer.flush();
        return err;
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

    async read(): Promise<[Uint8Array?, Error?]> {
        return this.#reader.readUint8Array();
    }

    cancel(code: number): void {
        this.#reader.cancel(code, "cancelled"); // TODO: Use a more descriptive message
        this.#cancelFunc(new Error("Stream cancelled")); // Notify the context about cancellation
    }

    get context(): Context {
        return this.#ctx;
    }
}