import type { Reader, Writer } from "./io";
import { withCancelCause } from "golikejs/context";
import type { CancelCauseFunc, Context } from "golikejs/context";
import type { Frame } from "./io";
import { StreamError } from "./io/error";
import type { GroupMessage } from "./message";
import { BytesFrame } from "./frame";
import type { GroupErrorCode } from "./error";
import { PublishAbortedErrorCode,SubscribeCanceledErrorCode } from "./error";

export class GroupWriter {
    #group: GroupMessage;
    #writer: Writer;
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;
    readonly streamId: bigint;

    constructor(trackCtx: Context, writer: Writer, group: GroupMessage) {
        this.#group = group;
        this.#writer = writer;
        this.streamId = writer.streamId ?? 0n;
        [this.#ctx, this.#cancelFunc] = withCancelCause(trackCtx);

        trackCtx.done().then(()=>{
            this.cancel(SubscribeCanceledErrorCode, "track was closed");
        });
    }

    get groupSequence(): bigint {
        return this.#group.sequence;
    }

    async writeFrame(src: Frame): Promise<Error | undefined> {
        this.#writer.copyFrom(src);
        const err = await this.#writer.flush();
        if (err) {
            return err;
        }

        return undefined;
    }

    async close(): Promise<void> {
        if (this.#ctx.err()) {
            return;
        }
        this.#cancelFunc(undefined); // Notify the context about closure
        await this.#writer.close();
    }

    async cancel(code: GroupErrorCode, message: string): Promise<void> {
        if (this.#ctx.err()) {
            // Do nothing if already cancelled
            return;
        }
        const cause = new StreamError(code, message);
        this.#cancelFunc(cause); // Notify the context about cancellation
        await this.#writer.cancel(cause);
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
    #frame?: BytesFrame;
    readonly streamId: bigint;

    constructor(trackCtx: Context, reader: Reader, group: GroupMessage) {
        this.#group = group;
        this.#reader = reader;
        this.streamId = reader.streamId ?? 0n;
        [this.#ctx, this.#cancelFunc] = withCancelCause(trackCtx);

        trackCtx.done().then(()=>{
            this.cancel(PublishAbortedErrorCode, "track was closed");
        });
    }

    get groupSequence(): bigint {
        return this.#group.sequence;
    }

    async readFrame(): Promise<[BytesFrame, undefined] | [undefined, Error]> {
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
            this.#frame = new BytesFrame(new Uint8Array(cap));
        }

        err = await this.#reader.fillN(this.#frame.bytes, len);
        if (err) {
            return [undefined, err];
        }

        return [this.#frame, undefined];
    }

    async cancel(code: GroupErrorCode, message: string): Promise<void> {
        if (this.#ctx.err()) {
            // Do nothing if already cancelled
            return;
        }
        const reason = new StreamError(code, message);
        this.#cancelFunc(reason);
        await this.#reader.cancel(reason);
    }

    get context(): Context {
        return this.#ctx;
    }
}