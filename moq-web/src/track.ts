import { GroupReader, GroupWriter } from "./group_stream";
import { Info } from "./info";
import { Queue } from "./internal";
import { Context } from "./internal/context";
import { ReceiveSubscribeStream, SendSubscribeStream, TrackConfig } from "./subscribe_stream";
import { Writer, Reader } from "./io";
import { UniStreamTypes } from "./stream_type";
import { GroupMessage } from "./message";
import { BroadcastPath } from "./broadcast_path";
import { PublishAbortedErrorCode, SubscribeErrorCode } from "./error";
import { GroupSequence } from ".";

export class TrackWriter {
    broadcastPath: BroadcastPath;
    trackName: string;
    #subscribeStream: ReceiveSubscribeStream;
    #openUniStreamFunc: () => Promise<[Writer?, Error?]>;
    #groups: GroupWriter[] = [];

    constructor(
        broadcastPath: BroadcastPath,
        trackName: string,
        subscribeStream: ReceiveSubscribeStream,
        openUniStreamFunc: () => Promise<[Writer?, Error?]>,
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

    async openGroup(groupSequence: GroupSequence): Promise<[GroupWriter?, Error?]> {
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
        if (!msg) {
            return [undefined, new Error("Failed to encode group message")];
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

    closeWithError(code: SubscribeErrorCode, message: string): void {
        // Cancel all groups with the error first
        for (const group of this.#groups) {
            group.cancel(PublishAbortedErrorCode, message);
        }

        // Then close the subscribe stream with the error
        this.#subscribeStream.closeWithError(code, message);
    }

    close(): void {
        for (const group of this.#groups) {
            group.close();
        }
        this.#subscribeStream.close();
    }
}

export class TrackReader {
    #subscribeStream: SendSubscribeStream;
    #acceptFunc: () => Promise<[Reader, GroupMessage]>;
    #onCloseFunc: () => void;

    constructor(subscribeStream: SendSubscribeStream,
        acceptFunc: () => Promise<[Reader, GroupMessage]>,
        onCloseFunc: () => void,
    ) {
        this.#subscribeStream = subscribeStream;
        this.#acceptFunc = acceptFunc;
        this.#onCloseFunc = onCloseFunc;
    }

    async acceptGroup(ctx?: Promise<void>): Promise<[GroupReader?, Error?]> {
        const promises: Promise<Error | [Reader, GroupMessage]>[] = [
            this.context.done().then(() => new Error(`subscribe stream cancelled: ${this.context.err()}`)),
            this.#acceptFunc(),
        ];
        if (ctx) {
            promises.push(ctx.then((): Error => new Error("Context cancelled")));
        }
        const result = await Promise.race(promises);
        if (result instanceof Error) {
            // Context was cancelled
            return [undefined, result];
        }

        const [reader, msg] = result;

        const group = new GroupReader(this.context, reader, msg);

        return [group, undefined];
    }

    async update(config: TrackConfig): Promise<Error | undefined> {
        return this.#subscribeStream.update(config);
    }

    readInfo(): Info {
        return this.#subscribeStream.info;
    }

    cancel(code: number, message: string): void {
        this.#subscribeStream.cancel(code, message);
        this.#onCloseFunc();
    }

    get trackConfig(): TrackConfig {
        return this.#subscribeStream.config;
    }

    get context(): Context {
        return this.#subscribeStream.context;
    }
}