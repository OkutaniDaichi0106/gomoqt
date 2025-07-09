import { GroupReader, GroupWriter } from "./group_stream";
import { Context } from "./internal/context";

export class TrackWriter {
    #ctx: Context;
    #openGroupFunc: (ctx: Context, groupId: bigint) => Promise<GroupWriter>;

    constructor(trackCtx: Context, openGroupFunc: (ctx: Context, groupId: bigint) => Promise<GroupWriter>) {
        this.#ctx = trackCtx;
        this.#openGroupFunc = openGroupFunc;
    }

    get context(): Context {
        return this.#ctx;
    }

    async openGroup(groupId: bigint): Promise<GroupWriter> {
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
        const error = this.#ctx.err();
        if (error != null) {
            return [undefined, error];
        }

        return await this.#acceptFunc();
    }

    get context(): Context {
        return this.#ctx;
    }
}