import { GroupReader, GroupWriter } from "./group_stream";
import type { Info } from "./info";
import { Queue } from "./internal";
import type { Context} from "./internal/context";
import {ContextCancelledError,withCancel,withPromise } from "./internal/context";
import type { ReceiveSubscribeStream, SendSubscribeStream, TrackConfig } from "./subscribe_stream";
import type { Writer, Reader } from "./io";
import { UniStreamTypes } from "./stream_type";
import { GroupMessage } from "./message";
import type { BroadcastPath } from "./broadcast_path";
import type { SubscribeErrorCode } from "./error";
import { PublishAbortedErrorCode } from "./error";
import type { GroupSequence } from ".";

export class TrackWriter {
    broadcastPath: BroadcastPath;
    trackName: string;
    #subscribeStream: ReceiveSubscribeStream;
    #openUniStreamFunc: () => Promise<[Writer, undefined] | [undefined, Error]>;
    #groups: GroupWriter[] = [];

    constructor(
        broadcastPath: BroadcastPath,
        trackName: string,
        subscribeStream: ReceiveSubscribeStream,
        openUniStreamFunc: () => Promise<[Writer, undefined] | [undefined, Error]>,
    ) {
        this.broadcastPath = broadcastPath;
        this.trackName = trackName;
        this.#subscribeStream = subscribeStream;
        this.#openUniStreamFunc = openUniStreamFunc;
    }

    get context(): Context {
        return this.#subscribeStream.context;
    }

    get subscribeId(): bigint {
        return this.#subscribeStream.subscribeId;
    }

    get config(): TrackConfig {
        return this.#subscribeStream.trackConfig;
    }

    async openGroup(groupSequence: GroupSequence): Promise<[GroupWriter, undefined] | [undefined, Error]> {
        let err: Error | undefined;
        err = await this.#subscribeStream.writeInfo()
        if (err) {
            return [undefined, err];
        }

        let writer: Writer | undefined;
        [writer, err] = await this.#openUniStreamFunc();
        if (err) {
            return [undefined, err];
        }

        writer!.writeUint8(UniStreamTypes.GroupStreamType);
        const msg = new GroupMessage({
            subscribeId: this.subscribeId,
            sequence: groupSequence
        });
        err = await msg.encode(writer!);
        if (err) {
            return [undefined, new Error("Failed to create group message")];
        }

        const group = new GroupWriter(this.context, writer!, msg)

        this.#groups.push(group);

        return [group, undefined];
    }

    async writeInfo(info: Info): Promise<Error | undefined> {
        const err = await this.#subscribeStream.writeInfo(info);
        if (err) {
            return err;
        }
    }

    async closeWithError(code: SubscribeErrorCode, message: string): Promise<void> {
        // Cancel all groups with the error first
        await Promise.allSettled(this.#groups.map(
            (group) => group.cancel(PublishAbortedErrorCode, message)
        ))

        // Then close the subscribe stream with the error
        await this.#subscribeStream.closeWithError(code, message);
    }

    async close(): Promise<void> {
        await Promise.allSettled(this.#groups.map(
            (group) => group.close()
        ))

        await this.#subscribeStream.close();
    }
}

export class TrackReader {
    #subscribeStream: SendSubscribeStream;
    #acceptFunc: (ctx: Promise<void>) => Promise<[Reader, GroupMessage] | undefined>;
    #onCloseFunc: () => void;

    constructor(
        subscribeStream: SendSubscribeStream,
        acceptFunc: (ctx: Promise<void>) => Promise<[Reader, GroupMessage] | undefined>,
        onCloseFunc: () => void,
    ) {
        this.#subscribeStream = subscribeStream;
        this.#acceptFunc = acceptFunc;
        this.#onCloseFunc = onCloseFunc;
    }

    async acceptGroup(signal: Promise<void>): Promise<[GroupReader, undefined]|[undefined, Error]> {
        // Check if context is already cancelled
        const err = this.context.err();
        if (err) {
            return [undefined, err];
        }

        const ctx = withPromise(this.context, signal);

        const dequeued = await this.#acceptFunc(ctx.done());
        if (dequeued === undefined) {
            return [undefined, new Error("[TrackReader] failed to dequeue group message")];
        }

        const [reader, msg] = dequeued;

        const group = new GroupReader(this.context, reader, msg);

        return [group, undefined];
    }

    async update(config: TrackConfig): Promise<Error | undefined> {
        return this.#subscribeStream.update(config);
    }

    readInfo(): Info {
        return this.#subscribeStream.info;
    }

    async closeWithError(code: number, message: string): Promise<void> {
        await this.#subscribeStream.closeWithError(code, message);
        this.#onCloseFunc();
    }

    get trackConfig(): TrackConfig {
        return this.#subscribeStream.config;
    }

    get context(): Context {
        return this.#subscribeStream.context;
    }
}