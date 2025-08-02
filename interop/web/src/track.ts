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

export class TrackWriter {
    broadcastPath: BroadcastPath;
    trackName: string;
    #subscribeStream: ReceiveSubscribeStream;
    #openUniStreamFunc: () => Promise<[Writer?, Error?]>;
    #accepted: boolean = false;
    #groups: GroupWriter[] = [];

    constructor(broadcastPath: BroadcastPath, trackName: string,
        subscribeStream: ReceiveSubscribeStream,
        openUniStreamFunc: () => Promise<[Writer?, Error?]>
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

    get trackConfig(): TrackConfig {
        return this.#subscribeStream.trackConfig;
    }

    async openGroup(groupId: bigint): Promise<[GroupWriter?, Error?]> {
        if (!this.#accepted) {
            const err = await this.#subscribeStream.accept({
                groupOrder: 0,
                trackPriority: 0
            })
            if (err) {
                return [undefined, err];
            }
            this.#accepted = true;
        }

        const [writer, err] = await this.#openUniStreamFunc();
        if (err) {
            return [undefined, err];
        }

        if (!writer) {
            return [undefined, new Error("Failed to create group writer")];
        }

        writer.writeUint8(UniStreamTypes.GroupStreamType);
        const [msg, err2] = await GroupMessage.encode(writer, this.subscribeId, groupId);
        if (err2) {
            return [undefined, new Error("Failed to create group message")];
        }
        if (!msg) {
            return [undefined, new Error("Failed to encode group message")];
        }

        const group = new GroupWriter(this.context, writer, msg)

        this.#groups.push(group);

        return [group, undefined];
    }

    async writeInfo(info: Info): Promise<Error | undefined> {
        if (this.#accepted) {
            return undefined; // Already accepted, no need to write info again
        }

        const err = await this.#subscribeStream.accept(info);
        if (err) {
            return err;
        }

        this.#accepted = true;
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
    #acceptFunc: () => Promise<[Reader, GroupMessage] | undefined>;
    #onCloseFunc: () => void;

    constructor(subscribeStream: SendSubscribeStream,
        acceptFunc: () => Promise<[Reader, GroupMessage] | undefined>,
        onCloseFunc: () => void,
    ) {
        this.#subscribeStream = subscribeStream;
        this.#acceptFunc = acceptFunc;
        this.#onCloseFunc = onCloseFunc;
    }

    async acceptGroup(): Promise<[GroupReader?, Error?]> {
        const ctxErr = this.context.err();
        if (ctxErr != null) {
            return [undefined, ctxErr];
        }

        const item = await this.#acceptFunc();
        if (item === undefined) {
            return [undefined, new Error("No group available")];
        }

        const [reader, msg] = item;
        const group = new GroupReader(this.context, reader, msg);

        return [group, undefined];
    }

    async update(trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint): Promise<Error | undefined> {
        return this.#subscribeStream.update(trackPriority, minGroupSequence, maxGroupSequence);
    }

    cancel(code: number, message: string): void {
        this.#subscribeStream.cancel(code, message);
        this.#onCloseFunc();
    }

    get trackConfig(): TrackConfig {
        return this.#subscribeStream.trackConfig;
    }

    get context(): Context {
        return this.#subscribeStream.context;
    }
}