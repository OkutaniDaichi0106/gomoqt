import type { Reader, Source, Writer } from "./webtransport";
import { withCancelCause } from "golikejs/context";
import type { CancelCauseFunc, Context } from "golikejs/context";
import { StreamError } from "./webtransport/error";
import type { GroupMessage } from "./message";
import { BytesFrame } from "./frame";
import type { Frame } from "./frame";
import type { GroupErrorCode } from "./error";
import { PublishAbortedErrorCode,SubscribeCanceledErrorCode } from "./error";

export class GroupWriter {
    readonly sequence: bigint;
    #writer: Writer;
    readonly context: Context;
    #cancelFunc: CancelCauseFunc;
    readonly streamId: bigint;

    constructor(trackCtx: Context, writer: Writer, group: GroupMessage) {
        this.sequence = group.sequence;
        this.#writer = writer;
        this.streamId = writer.streamId ?? 0n;
        [this.context, this.#cancelFunc] = withCancelCause(trackCtx);

        trackCtx.done().then(()=>{
            this.cancel(SubscribeCanceledErrorCode, "track was closed");
        });
    }

    async writeFrame(src: Source): Promise<Error | undefined> {
        this.#writer.copyFrom(src);
        const err = await this.#writer.flush();
        if (err) {
            return err;
        }

        return undefined;
    }

    async close(): Promise<void> {
        if (this.context.err()) {
            return;
        }
        this.#cancelFunc(undefined); // Notify the context about closure
        await this.#writer.close();
    }

    async cancel(code: GroupErrorCode, message: string): Promise<void> {
        if (this.context.err()) {
            // Do nothing if already cancelled
            return;
        }
        const cause = new StreamError(code, message);
        this.#cancelFunc(cause); // Notify the context about cancellation
        await this.#writer.cancel(cause);
    }
}

export class GroupReader {
    readonly sequence: bigint;
    #reader: Reader;
    readonly context: Context;
    #cancelFunc: CancelCauseFunc;
    #frame?: BytesFrame;
    readonly streamId: bigint;

    constructor(trackCtx: Context, reader: Reader, group: GroupMessage) {
        this.sequence = group.sequence;
        this.#reader = reader;
        this.streamId = reader.streamId ?? 0n;
        [this.context, this.#cancelFunc] = withCancelCause(trackCtx);

        trackCtx.done().then(()=>{
            this.cancel(PublishAbortedErrorCode, "track was closed");
        });
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
        if (this.context.err()) {
            // Do nothing if already cancelled
            return;
        }
        const reason = new StreamError(code, message);
        this.#cancelFunc(reason);
        await this.#reader.cancel(reason);
    }
}