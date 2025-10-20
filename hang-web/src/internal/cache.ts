import {
    ExpiredGroupErrorCode,
    PublishAbortedErrorCode,
    TrackWriter,
    InternalGroupErrorCode
} from "@okutanidaichi/moqt";
import {
    Cond,
    Mutex,
} from "golikejs/sync";
import type {
    GroupSequence,
    GroupWriter,
    GroupErrorCode,
} from "@okutanidaichi/moqt";
import type { Frame } from "@okutanidaichi/moqt/io";

export class GroupCache {
    readonly sequence: GroupSequence;
    readonly timestamp: number;
    frames: Frame[] = [];
    // dests: Set<GroupWriter> = new Set();
    closed: boolean = false;
    expired: boolean = false;
    #mutex = new Mutex();
    #cond = new Cond(this.#mutex);

    constructor(sequence: GroupSequence, timestamp: number) {
        this.sequence = sequence;
        this.timestamp = timestamp;
    }

    async append(frame: Frame): Promise<void> {
        await this.#mutex.lock();

        if (this.closed) {
            this.#mutex.unlock();
            return;
        }

        this.frames.push(frame);

        // Signal one waiting flush that a new frame is available
        this.#cond.signal();

        this.#mutex.unlock();
    }

    async connect(group: GroupWriter): Promise<void> {
        await this.#mutex.lock();

        // Write current frames to group
        let err: Error | undefined;
        let framesCount = 0;
        for (const frame of this.frames) {
            if (frame) {
                err = await group.writeFrame(frame);
                if (err) {
                    group.cancel(InternalGroupErrorCode, `failed to write frame: ${err.message}`);
                    this.#mutex.unlock();
                    return;
                }
                framesCount++;
            }
        }

        // Release lock before entering the wait loop
        this.#mutex.unlock();

        // Wait for new frames or close/expire signals
        while (true) {
            // Write any new frames that arrived
            while (framesCount < this.frames.length) {
                const frame = this.frames[framesCount];
                err = await group.writeFrame(frame!);
                if (err) {
                    group.cancel(InternalGroupErrorCode, `failed to write frame: ${err.message}`);
                    return;
                }
                framesCount++;
            }

            // Lock and wait for signal
            this.#mutex.lock();
            await this.#cond.wait();
            this.#mutex.unlock();

            // Check termination conditions
            if (this.closed) {
                group.close();
                return;
            }
            if (this.expired) {
                group.cancel(ExpiredGroupErrorCode, 'cache expired');
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

        // Broadcast to all waiting flush operations
        this.#cond.broadcast();

        this.#mutex.unlock();
    }

    async expire(): Promise<void> {
        await this.#mutex.lock();

        if (this.expired) {
            this.#mutex.unlock();
            return;
        }

        this.expired = true;
        this.frames.length = 0;

        // Broadcast to all waiting flush operations
        this.#cond.broadcast();

        this.#mutex.unlock();
    }
}

export interface TrackCache {
    store(group: GroupCache): void;
    close(): Promise<void>;
}