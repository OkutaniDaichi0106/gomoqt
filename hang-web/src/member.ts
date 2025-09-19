import type { BroadcastViewer } from "./broadcast";

export interface JoinedMember {
    remote: boolean;
    name: string;
    broadcast: BroadcastViewer;
}

export interface LeftMember {
    remote: boolean;
    name: string;
}