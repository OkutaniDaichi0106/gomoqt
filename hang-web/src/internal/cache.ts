import {
    ExpiredGroupErrorCode,
    PublishAbortedErrorCode,
    TrackWriter,
InternalGroupErrorCode
} from "@okutanidaichi/moqt";
import {
    Mutex,
    Cond,
} from "golikejs/sync";
import type {
    GroupSequence,
    Frame,
    GroupWriter,
    GroupErrorCode,
} from "@okutanidaichi/moqt";
import type { Source } from "@okutanidaichi/moqt/io";

export class GroupCache {
    readonly sequence: GroupSequence;
    readonly timestamp: number;
    frames: (Frame | Source)[] = [];
    // dests: Set<GroupWriter> = new Set();
    closed: boolean = false;
    expired: boolean = false;
    #mutex = new Mutex()
    #cond: Cond = new Cond(this.#mutex);

    constructor(sequence: GroupSequence, timestamp: number) {
        this.sequence = sequence;
        this.timestamp = timestamp;
    }

    async append(frame: Frame | Source): Promise<void> {
        await this.#mutex.lock();

        if (this.closed) {
            this.#mutex.unlock();
            return;
        }


        this.frames.push(frame);

        // await Promise.allSettled(
        //     Array.from(this.dests, async (group) => {
        //         const err = await group.writeFrame(frame)
        //         if (err) {
        //             this.dests.delete(group);
        //         }
        //     })
        // );

        this.#mutex.unlock();

        this.#cond.broadcast();
    }

    async flush(group: GroupWriter): Promise<void> {
        await this.#mutex.lock();

        // Write current frames to group
        let err: Error | undefined;
        for (const frame of this.frames) {
            if (frame) {
                err = await group.writeFrame(frame);
                if (err) {
                    group.cancel(InternalGroupErrorCode, `failed to write frame: ${err.message}`); // TODO: Is this correct?
                    this.#mutex.unlock();
                    return;
                }
            }
        }

        let framesCount = this.frames.length;

        this.#mutex.unlock();

        while (true) {
            while (framesCount < this.frames.length) {
                const frame = this.frames[framesCount];
                err = await group.writeFrame(frame!);
                if (err) {
                    group.cancel(InternalGroupErrorCode, `failed to write frame: ${err.message}`); // TODO: Is this correct?
                    return;
                }
                framesCount++;
            }
            await this.#cond.wait();

            if (this.closed) {
                group.close();
                return;
            } else if (this.expired) {
                group.cancel(ExpiredGroupErrorCode, `cache expired`);
                return;
            }
        }
    }

    async close(): Promise<void> {
        await this.#mutex.lock();

        if (this.closed) {
            this.#mutex.unlock();
            return;
        }

        this.closed = true;

        this.#mutex.unlock();

        this.#cond.broadcast();
    }

    async expire(): Promise<void> {
        await this.#mutex.lock();

        if (this.expired) {
            this.#mutex.unlock();
            return;
        }

        this.expired = true;

        this.frames.length = 0;

        this.#mutex.unlock();

        this.#cond.broadcast();
    }
}

export interface TrackCache {
    store(group: GroupCache): void;
    close(): Promise<void>;
}