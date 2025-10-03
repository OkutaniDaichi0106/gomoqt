import type { TrackDescriptor } from "../catalog";
import type { TrackEncoder } from ".";

export interface EncoderStreamInit {
    output: (descriptor: TrackDescriptor, encoder: TrackEncoder) => void;
    error: (error: Error) => void;
}

export class TrackEncodeStream {

    constructor(init: EncoderStreamInit) {
    }
}