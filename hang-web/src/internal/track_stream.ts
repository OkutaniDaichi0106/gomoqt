import { ExpiredGroupErrorCode, TrackReader, GroupReader,type Frame,type GroupSequence } from "@okutanidaichi/moqt";
import type { Context } from "golikejs/context";

export class TrackStream {
    #track: TrackReader;
    #latency: number;
    #playhead: GroupSequence = 0n;
    #groups: GroupReader[] = [];

    readonly reader: ReadableStream<Frame>;

    constructor(track: TrackReader, latency: number) {
        this.#track = track;
        this.#latency = latency;

        this.reader = new ReadableStream({

        });
    }

    async #handle(ctx: Context): Promise<void> {
        while (true) {
            const [group, err] = await this.#track.acceptGroup(ctx.done());
            if (err) {
                // TODO: handle error
                return;
            }

            if (this.#playhead > group.sequence) {
                // Received group is already played or outdated
                // Close and skip it
                group.cancel(ExpiredGroupErrorCode, "no longer needed group");
                continue;
            }
        }
    }
}