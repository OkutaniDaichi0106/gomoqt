import type { BroadcastSubscriber } from "./broadcast";

export interface JoinedLocalMember {
    remote: false;
    name: string;
}

export interface JoinedRemoteMember {
    remote: true;
    name: string;
    broadcast: BroadcastSubscriber;
}

export type JoinedMember = JoinedLocalMember | JoinedRemoteMember;

export interface LeftMember {
    remote: boolean;
    name: string;
}