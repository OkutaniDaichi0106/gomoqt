import { BroadcastPath } from "./broadcast_path";
import { SubscribeController, SubscribeID } from "./subscribe_stream";
import { TrackReader } from "./track";

export type Subscription = {
    broadcastPath: BroadcastPath;
    trackName: string;
    controller: SubscribeController;
    trackReader: TrackReader;
}