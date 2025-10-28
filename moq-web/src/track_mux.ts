import type { AnnouncementWriter } from "./announce_stream.ts";
import { Announcement } from "./announce_stream.ts";
import type { BroadcastPath } from "./broadcast_path.ts";

import type { TrackWriter } from "./track.ts";
import type { TrackPrefix } from "./track_prefix.ts";
import { isValidPrefix } from "./track_prefix.ts";
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

    async announce(announcement: Announcement, handler: TrackHandler): Promise<void> {
        const path = announcement.broadcastPath;

        console.debug(`[TrackMux] announcing track for path: ${path}`);

        if (!announcement.isActive()) {
            console.warn(`[TrackMux] announcement is not active for path: ${path}`);
            return;
        }

        const existing = this.#handlers.get(path);
        if (existing !== undefined && existing.announcement !== announcement) {
            console.warn(`[TrackMux] track already announced for path: ${path}`);

            // End the old announcement
            existing.announcement.end();
        }

        // Register the new announcement and handler
        this.#handlers.set(path, { announcement, handler });

        const completePromises: Promise<void>[] = [];

        for (const [prefix, announcers] of this.#announcers.entries()) {
            if (path.startsWith(prefix)) {
                console.debug(`[TrackMux] announcing track for path: ${path} to prefix: ${prefix}`);

                const sendPromises: Promise<void>[] = [];
                const failed: AnnouncementWriter[] = [];

                // Notify all announcers for this prefix
                for (const announcer of announcers) {
                    const sendPromise = announcer.send(announcement).then((err) => {
                        if (err) {
                            console.warn(`[TrackMux] failed to announce track for path: ${path} to prefix: ${prefix}: ${err}`);
                            failed.push(announcer);
                        }
                    });
                    sendPromises.push(sendPromise);
                }

                // Wait for all sends to complete and then clean up failed announcers
                const completed = Promise.all(sendPromises).then(() => {
                    for (const failedAnnouncer of failed) {
                        announcers.delete(failedAnnouncer);
                    }
                    if (announcers.size === 0) {
                        this.#announcers.delete(prefix);
                    }
                });

                completePromises.push(completed);
            }
        }

        // Wait for all notifications to complete
        await Promise.all(completePromises);

        // Wait for the announcement to end
        announcement.ended().then(() => {
            console.debug(`[TrackMux] track ended for path: ${path}`);
            // Remove the handler when the context is done
            this.#handlers.delete(path);
        });
    }

    async publish(ctx: Promise<void>, path: BroadcastPath, handler: TrackHandler) {
        await this.announce(new Announcement(path, ctx), handler);
    }

    async publishFunc(ctx: Promise<void>, path: BroadcastPath, handler: (ctx: Promise<void>, trackWriter: TrackWriter) => Promise<void>) {
        await this.publish(ctx, path, { serveTrack: handler });
    }

    async serveTrack(trackWriter: TrackWriter): Promise<void> {
        const path = trackWriter.broadcastPath;
        const announced = this.#handlers.get(path);
        if (announced) {
            console.debug(`[TrackMux] serving track for path: ${path}`);
            await announced.handler.serveTrack(announced.announcement.ended(), trackWriter);
        } else {
            console.warn(`[TrackMux] no handler for track for path: ${path}`);
            await NotFoundHandler.serveTrack(Promise.resolve(), trackWriter);
        }
    }

    async serveAnnouncement(writer: AnnouncementWriter, prefix: TrackPrefix): Promise<void> {
        // if (!isValidPrefix(prefix)) {
        //     throw new Error(`Invalid track prefix: ${prefix}`);
        // }

        console.debug(`[TrackMux] serving announcements for prefix: ${prefix}`);

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

    async close(): Promise<void> {
        const closePromises: Promise<void>[] = [];
        for (const announcers of this.#announcers.values()) {
            for (const announcer of announcers) {
                closePromises.push(announcer.close());
            }
        }
        await Promise.allSettled(closePromises);
        this.#announcers.clear();
        this.#handlers.clear();
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