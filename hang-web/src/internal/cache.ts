import { 
    type GroupSequence, 
    type Frame,
    type GroupWriter,
    ExpiredGroupErrorCode,
    type GroupErrorCode,
    PublishAbortedErrorCode 
} from "@okutanidaichi/moqt";
import type { Source } from "@okutanidaichi/moqt/io";

export class GroupCache implements TrackCache {
    sequence: GroupSequence = 0n;
    frames: (Frame | Source)[] = [];
    dests: Map<number, GroupWriter[]> = new Map();

    constructor() {}

    append(sequence: GroupSequence, frame: Frame | Source): void {
        if (sequence < this.sequence) {
            return;
        }

        this.frames.push(frame);

        const frameCount = this.frames.length;

        const prevDests = this.dests.get(frameCount - 1) || [];

        for (const gw of prevDests) {
            try {
                gw.writeFrame(frame);
                // Shift the current group writers for the next frame
                if (!this.dests.has(frameCount)) {
                    this.dests.set(frameCount, []);
                }
                this.dests.get(frameCount)!.push(gw);
            } catch (err) {
                // Continue on error
                continue;
            }
        }

        // Clear the previous group writers
        this.dests.set(frameCount - 1, []);
    }

    expire(sequence: GroupSequence): void {
        if (sequence < this.sequence) {
            return;
        }

        this.sequence = sequence;
        const frameCount = this.frames.length;
        for (const [k, groups] of this.dests) {
            const last = k === frameCount;
            for (const gw of groups) {
                if (last) {
                    gw.close();
                } else {
                    gw.cancel(ExpiredGroupErrorCode, "new group was arrived"); // TODO: Use more appropriate error code
                }
		    }
	    }
	    this.frames.length = 0;
    }

    async flush(gw: GroupWriter): Promise<void> {
        if (gw.groupSequence !== this.sequence) {
            return;
        }

        const frameCount = this.frames.length;

        // Add gw to dests
        if (!this.dests.has(frameCount)) {
            this.dests.set(frameCount, []);
        }
        this.dests.get(frameCount)!.push(gw);

        // Create a snapshot of current frames to avoid race conditions
        // This is efficient - just copies the array reference, not the data
        const frames = [...this.frames];

        // Write frames to gw
        for (const frame of frames) {
            if (frame) {
                gw.writeFrame(frame);
            }
        }
    }

    close(): void {
        for (const groups of this.dests.values()) {
            for (const gw of groups) {
                gw.close();
            }
        }

        this.frames.length = 0;

        this.dests.clear();
    }

    closeWithError(reason: string): void {
        for (const groups of this.dests.values()) {
            for (const gw of groups) {
                gw.cancel(PublishAbortedErrorCode, `cache closed: ${reason}`);
            }
        }

        this.frames.length = 0;

        this.dests.clear();
    }
}

export interface TrackCache {
    append(sequence: GroupSequence, frame: Frame | Source): void;
    flush(gw: GroupWriter): Promise<void>;
    expire(sequence: GroupSequence): void;
    close(): void;
    closeWithError(reason: string): void;
}