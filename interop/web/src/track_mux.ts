import { Announcement, AnnouncementWriter } from "./announce_stream";
import { BroadcastPath } from "./broadcast_path";
import { Context } from "./internal/context";
import { Publisher } from "./publisher";
import { PublishController } from "./subscribe_stream";
import { TrackWriter } from "./track";
import { isValidPrefix, TrackPrefix } from "./track_prefix";

export class TrackMux {
    #handlers: Map<string, TrackHandler> = new Map();
    #announcers: Map<string, Set<AnnouncementWriter>> = new Map();

    constructor() {}

    announce(announcement: Announcement, handler: TrackHandler): void {
        const path = announcement.broadcastPath;
        this.#handlers.set(path, handler);

        for (const [prefix, announcers] of this.#announcers.entries()) {
            if (path.startsWith(prefix)) {
                // Notify all announcers for this prefix
                for (const announcer of announcers) {
                    announcer.send(announcement);
                }
            }
        }

        (async () => {
            // Wait for the context to be done
            await announcement.ended();
            // Remove the handler when the context is done
            this.#handlers.delete(path);
        })();
    }

    handlerTrack(ctx: Context, path: BroadcastPath, handler: TrackHandler) {
        this.announce(new Announcement(path, ctx), handler);
    }

    async serveTrack(publisher: Publisher): Promise<void> {
        const path = publisher.broadcastPath;
        const handler = this.#handlers.get(path);
        if (handler) {
            handler.serveTrack(publisher);
        } else {
            console.warn(`No handler found for track at path: ${path}`);
        }
    }

    async serveAnnouncement(writer: AnnouncementWriter, prefix: TrackPrefix): Promise<void> {
        if (!isValidPrefix(prefix)) {
            throw new Error(`Invalid track prefix: ${prefix}`);
        }

        if (!this.#announcers.has(prefix)) {
            this.#announcers.set(prefix, new Set());
        }

        const announcers = this.#announcers.get(prefix)!;
        announcers.add(writer);

        (async () => {
            // Wait for the context to be done
            await writer.context.done();
            // Remove the announcer when the context is done
            announcers.delete(writer);
            if (announcers.size === 0) {
                this.#announcers.delete(prefix);
            }
        })();
    }
}



export interface TrackHandler {
    serveTrack(publisher: Publisher): void;
}