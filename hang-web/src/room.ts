import type {
    Session,
    BroadcastPath,
Announcement,
} from "@okutanidaichi/moqt";
import {
    validateBroadcastPath,
InternalAnnounceErrorCode,
} from "@okutanidaichi/moqt";
import {
    BroadcastPublisher,
    BroadcastSubscriber,
} from ".";
import {
    background,
    type Context,
    type CancelCauseFunc,
    withCancelCause
} from "@okutanidaichi/moqt/internal";
import type { JoinedMember,LeftMember } from "./member";

const HANG_EXTENSION = '.hang';

export class Room {
    readonly roomID: string;

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    #local?: BroadcastPublisher;
    #remotes: Map<BroadcastPath, BroadcastSubscriber> = new Map();

    #onmember: MemberHandler;

    constructor(init: RoomInit) {
        this.roomID = init.roomID;
        this.#onmember = init.onmember;
        [this.#ctx, this.#cancelCtx] = withCancelCause(background());
    }

    // get local(): BroadcastPublisher {
    //     return this.#local!;
    // }

    // get remotes(): ReadonlyMap<BroadcastPath, BroadcastSubscriber> {
    //     return this.#remotes;
    // }

    async join(session: Session, localName: string): Promise<BroadcastPublisher> {
        if (this.#local) {
            console.warn("Already joined the room. Leaving and rejoining.");
            this.leave();

            // Reset the context
            [this.#ctx, this.#cancelCtx] = withCancelCause(background());
        }

        const path = validateBroadcastPath(`/${this.roomID}/${localName}${HANG_EXTENSION}`);
        const local = new BroadcastPublisher(path, ""); // TODO: description
        session.mux.publish(this.#ctx.done(), path, local);

        const announced = await session.acceptAnnounce(`/${this.roomID}/`);
        // let err = this.#ctx.err();
        // if (err !== undefined) {
        //     // Context was cancelled
        //     announced.closeWithError(InternalAnnounceErrorCode, err.message);
        //     return err;
        // }

        (async () => {
            // Listen for further announcements until the context is done
            while (true) {
                const [announcement, err] = await announced.receive(this.#ctx.done());

                if (err) {
                    // TODO: Handle specific errors
                    this.#cancelCtx(err);
                    break;
                }

                this.#addRemote(announcement!, session);
            }
        })();

        this.#local = local;

        // Call onJoin for the local member
        this.#onmember.onJoin({
            remote: false,
            name: localName,
            broadcast: local
        });

        return local;
    }

    leave(): void {
        if (!this.#local) return;

        // Cancel the room context
        // This will also stop all broadcasts and remote subscriptions
        this.#cancelCtx(new Error("left the room"));

        this.#local = undefined;
    }

    close(): void {
        this.leave();
        for (const remote of this.#remotes.values()) {
            remote.close();
        }
        this.#remotes.clear();
    }

    #addRemote(announcement: Announcement, session: Session): void {
        const path = announcement.broadcastPath;

        const remote = new BroadcastSubscriber(path, session);

        // If the remote is the same as the existing one, do nothing
        const got = this.#remotes.get(path)
        if (remote === got) {
            return;
        }
        if (got) {
            // Already have this remote.
            // This must not be possible, but just in case, leave the existing one.
            got.close();
        }

        this.#remotes.set(path, remote);
        this.#onmember.onJoin({
            remote: true,
            name: remoteName(this.roomID, path),
            broadcast: remote
        });

        // Clean up the remote when the announcement ends
        announcement!.ended().then(() => {
            // Only remove if the remote matches the existing one
            if (remote === this.#remotes.get(path)) {
                this.#remotes.delete(path);
                this.#onmember.onLeave({
                    remote: true,
                    name: remoteName(this.roomID, path),
                });
            }
        });
    }

    // handleMember(handler?: MemberHandler): void {
    //     this.#memberHandler = handler;
    //     // Call the handler for existing members
    //     if (handler) {
    //         // Call for the local member
    //         handler.onJoin({
    //             remote: false,
    //             name: this.localName,
    //             broadcast: this.#local
    //         });

    //         // Call for existing remote members
    //         for (const [path, remote] of this.#remotes) {
    //             handler.onJoin({
    //                 remote: true,
    //                 name: remoteName(this.roomID, path),
    //                 broadcast: remote
    //             });
    //         }
    //     }
    // }

    get remoteNames(): string[] {
        return Array.from(this.#remotes.keys()).map(path => remoteName(this.roomID, path));
    }
}

export interface RoomInit {
    roomID: string;
    description?: string;

    onmember: MemberHandler;

    // Optional token for authentication
    // token?: string; // TODO: Implement token-based authentication
}

export interface MemberHandler {
    onJoin: (member: JoinedMember) => void;
    onLeave: (member: LeftMember) => void;
}

export function remoteName(roomID: string, broadcastPath: BroadcastPath): string {
    return broadcastPath.substring(roomID.length + 2).replace(HANG_EXTENSION, '');
}