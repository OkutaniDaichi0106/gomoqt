import type {
    Session,
    BroadcastPath,
    Announcement,
    AnnouncementReader,
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
    withCancelCause,
} from "golikejs/context";
import type {
    JoinedMember,
    LeftMember
} from "./member";

const HANG_EXTENSION = '.hang';

export class Room {
    readonly roomID: string;

    #remotes: Map<string, BroadcastSubscriber> = new Map();
    #cancel?: CancelCauseFunc;

    #onmember: MemberHandler;

    #wg: Promise<void>[] = [];

    constructor(init: RoomInit) {
        this.roomID = init.roomID;
        this.#onmember = init.onmember;
    }

    async join(session: Session, local: BroadcastPublisher): Promise<void> { // TODO: use session interface from moqt when available
        if (this.#cancel) {
            // If already joined, leave first
            await this.leave();
        }

        let ctx: Context
        [ctx, this.#cancel] = withCancelCause(background());

        const path = broadcastPath(this.roomID, local.name);

        // Publish the local broadcast to the track mux and make it available to others
        // This broadcast will end when the local broadcast is closed
        session.mux.publish(ctx.done(), path, local);

        const [announcements, err] = await session.acceptAnnounce(`/${this.roomID}/`);
        if (err) {
            console.warn(`[Room] failed to accept announcements for room: ${this.roomID}: ${err}`);
            throw err;
        }

        let resolveAck: (() => void);
        const ack = new Promise<void>((resolve) => {
            resolveAck = resolve;
        });

        this.#wg.push(
            this.#handleAnnouncements(ctx.done(), announcements!, session, local, resolveAck!)
        );

        await ack;

        return;
    }

    async #handleAnnouncements(
        signal: Promise<void>,
        announcements: AnnouncementReader,
        session: Session,
        local: BroadcastPublisher,
        resolveAck: () => void
    ): Promise<void> {
        const localPath = broadcastPath(this.roomID, local.name);
        // Listen for further announcements until the context is done
        while (true) {
            const [announcement, err] = await announcements.receive(signal);
            if (err) {
                // TODO: Handle errors
                break;
            }

            // Handle announcement for ourselves (e.g. re-announcement) as ACK
            if (announcement!.broadcastPath === localPath) {
                resolveAck();

                this.#addLocal(local);

                this.#wg.push(
                    announcement!.ended().then(() => {
                        this.#removeLocal(local);
                    })
                );

                return;
            }

            // Try to subscribe to the announced broadcast
            try {
                const broadcast = new BroadcastSubscriber(announcement!.broadcastPath, this.roomID, session);
                this.#addRemote(broadcast);
                // Clean up the remote when the announcement ends
                announcement!.ended().then(() => {
                    this.#removeRemote(broadcast);
                });
            } catch (e) {
                console.warn(`[Room] failed to subscribe to ${announcement}: ${e}`);
            }
        }

        // Ensure announcements reader is closed
        await announcements?.close();
    }

    async leave(): Promise<void> {
        if (this.#cancel) {
            this.#cancel(new Error("hang: room left"));
        }

        for (const [path, remote] of this.#remotes) {
            try {
                this.#removeRemote(remote);
            } catch (e) {
                console.warn(`hang: Error removing remote broadcast for path ${path}: ${e}`);
            }
        }
        this.#remotes.clear();

        await Promise.all(this.#wg);
        this.#wg = [];
    }

    #addLocal(local: BroadcastPublisher): void {
        this.#onmember.onJoin({
            remote: false,
            name: local.name,
            // broadcast: local
        });
    }

    #removeLocal(local: BroadcastPublisher): void {
        this.#onmember.onLeave({
            remote: false,
            name: local.name,
        });
    }

    #removeRemote(remote: BroadcastSubscriber): void {
        const got = this.#remotes.get(remote.name);

        if (!got) {
            return;
        }

        // Close the broadcast to clean up resources
        remote.close();

        if (got !== remote) {
            return;
        }

        // Remove from map first to prevent re-entrancy issues
        this.#remotes.delete(remote.name);

        // Notify about remote member leaving
        this.#onmember.onLeave({
            remote: true,
            name: remote.name,
        });
    }

    #addRemote(remote: BroadcastSubscriber): void {
        // If the remote is the same as the existing one, do nothing
        const got = this.#remotes.get(remote.name);

        // Ignore if already have this exact remote
        if (remote === got) {
            return;
        }

        // If there is an existing remote with the same path, properly remove it first
        if (got) {
            // Properly remove the existing remote using #removeRemote
            // This ensures onLeave notification is sent and cleanup is done correctly
            this.#removeRemote(got);
        }

        this.#remotes.set(remote.name, remote);

        // Notify about new remote member joining
        this.#onmember.onJoin({
            remote: true,
            name: remote.name,
            broadcast: remote
        });
    }

    // get isJoined(): boolean {
    //     return this.#local !== undefined;
    // }
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

export function participantName(roomID: string, broadcastPath: BroadcastPath): string {
    // Extract the participant name from the broadcast path
    // Assumes the path format is "/<roomID>/<name>.hang"
    const name = broadcastPath.substring(roomID.length + 2).replace(HANG_EXTENSION, '');
    return name;
}

export function broadcastPath(roomID: string, name: string): BroadcastPath {
    return validateBroadcastPath(`/${roomID}/${name}${HANG_EXTENSION}`);
}