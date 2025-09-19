import type { AnnouncementWriter } from "./announce_stream";
import { Announcement } from "./announce_stream";
import type { BroadcastPath } from "./broadcast_path";

import type { TrackWriter } from "./track";
import type { TrackPrefix } from "./track_prefix";
import { isValidPrefix } from "./track_prefix";
import { TrackNotFoundErrorCode } from ".";

type AnnouncedTrackHandler = {
    announcement: Announcement;
    handler: TrackHandler;
}

export class TrackMux {
    #handlers: Map<BroadcastPath, AnnouncedTrackHandler> = new Map();
    #announcers: Map<TrackPrefix, Set<AnnouncementWriter>> = new Map();
    // #announcements: Map<string, Announcement> = new Map();

    constructor() {}

    announce(announcement: Announcement, handler: TrackHandler): void {
        const path = announcement.broadcastPath;
        this.#handlers.set(path, { announcement, handler });

        for (const [prefix, announcers] of this.#announcers.entries()) {
            if (path.startsWith(prefix)) {
                // Notify all announcers for this prefix
                for (const announcer of announcers) {
                    announcer.send(announcement);
                }
            }
        }

        // Wait for the announcement to end
        announcement.ended().then(() => {
            // Remove the handler when the context is done
            this.#handlers.delete(path);
        });
    }

    publish(ctx: Promise<void>, path: BroadcastPath, handler: TrackHandler) {
        this.announce(new Announcement(path, ctx), handler);
    }

    publishFunc(ctx: Promise<void>, path: BroadcastPath, handler: (ctx: Promise<void>, trackWriter: TrackWriter) => Promise<void>) {
        this.publish(ctx, path, { serveTrack: handler });
    }

    async serveTrack(trackWriter: TrackWriter): Promise<void> {
        const announced = this.#handlers.get(trackWriter.broadcastPath);
        if (announced) {
            await announced.handler.serveTrack(announced.announcement.ended(), trackWriter);
        } else {
            await NotFoundHandler.serveTrack(Promise.resolve(), trackWriter);
        }
    }

    async serveAnnouncement(writer: AnnouncementWriter, prefix: TrackPrefix): Promise<void> {
        if (!isValidPrefix(prefix)) {
            throw new Error(`Invalid track prefix: ${prefix}`);
        }

        console.log(`Serving announcement for prefix: ${prefix}`);

        let announced: AnnouncedTrackHandler;
        const init: Announcement[] = [];
        for (announced of this.#handlers.values()) {
            if (announced.announcement.broadcastPath.startsWith(prefix)) {
                init.push(announced.announcement);
            }
        }

        // Initialize the announcers map for this prefix if it doesn't exist
        if (!this.#announcers.has(prefix)) {
            this.#announcers.set(prefix, new Set());
        }

        // Register the writer as an announcer for this prefix
        const announcers = this.#announcers.get(prefix)!;
        announcers.add(writer);

        // Send initial announcements
        await writer.init(init);

        // Wait for the context to be done
        await writer.context.done();

        // Remove the announcer when the context is done
        announcers.delete(writer);
        if (announcers.size === 0) {
            this.#announcers.delete(prefix);
        }
    }
}

export const DefaultTrackMux = new TrackMux();

export interface TrackHandler {
    serveTrack(ctx: Promise<void>, trackWriter: TrackWriter): Promise<void>;
}

const NotFoundHandler: TrackHandler = {
    async serveTrack(ctx: Promise<void>, trackWriter: TrackWriter): Promise<void> {
        trackWriter.closeWithError(TrackNotFoundErrorCode, "Track not found");
    }
};