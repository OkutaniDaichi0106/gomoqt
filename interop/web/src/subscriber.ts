import { BroadcastPath } from "./broadcast_path";
import { SubscribeController, SubscribeID } from "./subscribe_stream";
import { TrackReader } from "./track";

export type Subscriber = {
    broadcastPath: BroadcastPath;
    trackName: string;
    subscribeId: SubscribeID;
    controller: SubscribeController;
    trackReader: TrackReader;
}