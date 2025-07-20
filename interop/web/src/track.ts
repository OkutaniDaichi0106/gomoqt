import { GroupReader, GroupWriter } from "./group_stream";
import { Info } from "./info";
import { Context } from "./internal/context";

export class TrackWriter {
    #ctx: Context;
    #openGroupFunc: (trackCtx: Context, groupId: bigint) => Promise<[GroupWriter?, Error?]>;
    #acceptFunc: (info: Info) => Promise<Error | undefined>;

    constructor(trackCtx: Context,
        openGroupFunc: (trackCtx: Context, groupId: bigint) => Promise<[GroupWriter?, Error?]>,
        acceptFunc: (info: Info) => Promise<Error | undefined>
    ) {
        this.#ctx = trackCtx;
        this.#openGroupFunc = openGroupFunc;
        this.#acceptFunc = acceptFunc;
    }

    get context(): Context {
        return this.#ctx;
    }

    async openGroup(groupId: bigint): Promise<[GroupWriter?, Error?]> {
        await this.#acceptFunc({groupOrder: 0, trackPriority: 0});
        return this.#openGroupFunc(this.#ctx, groupId);
    }
}

export class TrackReader {
    #ctx: Context;
    #acceptFunc: () => Promise<[GroupReader?, Error?]>;

    constructor(trackCtx: Context, acceptFunc: () => Promise<[GroupReader?, Error?]>) {
        this.#ctx = trackCtx;
        this.#acceptFunc = acceptFunc;
    }

    async acceptGroup(): Promise<[GroupReader?, Error?]> {
        const ctxErr = this.#ctx.err();
        if (ctxErr != null) {
            return [undefined, ctxErr];
        }

        return await this.#acceptFunc();
    }

    get context(): Context {
        return this.#ctx;
    }
}