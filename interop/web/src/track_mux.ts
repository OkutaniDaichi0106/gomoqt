import { Announcement, AnnouncementWriter } from "./announce_stream";
import { BroadcastPath } from "./broadcast_path";
import { Context } from "./internal/context";
import { TrackWriter } from "./track";
import { isValidPrefix, TrackPrefix } from "./track_prefix";

export class TrackMux {
    #handlers: Map<string, TrackHandler> = new Map();
    #announcers: Map<string, Set<AnnouncementWriter>> = new Map();
    #announcements: Map<string, Announcement> = new Map();

    constructor() {}

    announce(announcement: Announcement, handler: TrackHandler): void {
        const path = announcement.broadcastPath;
        this.#announcements.set(path, announcement);     
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
            this.#announcements.delete(path);
            this.#handlers.delete(path);
        })();
    }

    handleTrack(ctx: Context, path: BroadcastPath, handler: TrackHandler) {
        this.announce(new Announcement(path, ctx), handler);
    }

    async serveTrack(trackWriter: TrackWriter): Promise<void> {
        const handler = this.#handlers.get(trackWriter.broadcastPath);
        if (handler) {
            handler.serveTrack(trackWriter);
        } else {
            NotFoundHandler.serveTrack(trackWriter);
        }
    }

    async serveAnnouncement(writer: AnnouncementWriter, prefix: TrackPrefix): Promise<void> {
        if (!isValidPrefix(prefix)) {
            throw new Error(`Invalid track prefix: ${prefix}`);
        }

        console.log(`Serving announcement for prefix: ${prefix}`);

        // Notify all existing announcements that match the prefix
        for (const [path, announcement] of this.#announcements.entries()) {
            console.log(`Checking announcement for path: ${path}`);
            if (path.startsWith(prefix)) {
                console.log(`Sending existing announcement for path: ${path}`);
                // Notify all announcers for this prefix
                writer.send(announcement);
            }
        }

        // Register the announcer for this prefix
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

export const DefaultTrackMux = new TrackMux();

export interface TrackHandler {
    serveTrack(trackWriter: TrackWriter): void;
}

const NotFoundHandler: TrackHandler = {
    serveTrack(trackWriter: TrackWriter): void {
        trackWriter.closeWithError(0x03, "Track not found");
    }
};