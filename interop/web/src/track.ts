import { GroupReader, GroupWriter } from "./group_stream";
import { Info } from "./info";
import { Queue } from "./internal";
import { Context } from "./internal/context";
import { ReceiveSubscribeStream, SendSubscribeStream, TrackConfig } from "./subscribe_stream";
import { Writer } from "./io";
import { UniStreamTypes } from "./stream_type";
import { GroupMessage } from "./message";
import { BroadcastPath } from "./broadcast_path";

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
            this.#subscribeStream.accept({
                groupOrder: 0,
                trackPriority: 0
            })
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

    closeWithError(code: SubscribeErrorCode, message: string): void {
        for (const group of this.#groups) {
            group.cancel(PublishAbortedErrorCode, message);
        }
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
    #queue: Queue<GroupReader>;
    #onCloseFunc: () => void;

    constructor(subscribeStream: SendSubscribeStream, queue: Queue<GroupReader>,
        onCloseFunc: () => void,
    ) {
        this.#subscribeStream = subscribeStream;
        this.#queue = queue;
        this.#onCloseFunc = onCloseFunc;
    }

    async acceptGroup(): Promise<[GroupReader?, Error?]> {
        const ctxErr = this.context.err();
        if (ctxErr != null) {
            return [undefined, ctxErr];
        }

        const group = await this.#queue.dequeue();
        if (group === undefined) {
            return [undefined, new Error("No group available")];
        }

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