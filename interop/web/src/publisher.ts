import { BroadcastPath } from "./broadcast_path";
import { PublishController } from "./subscribe_stream";
import { TrackWriter } from "./track";

export type Publisher = {
    broadcastPath: BroadcastPath;
    trackName: string;
    controller: PublishController;
    trackWriter: TrackWriter;
}