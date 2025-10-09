import { VideoContext, VideoObserveNode } from "./internal";

export interface VideoRendererInit {
    width?: number; // Canvas width (default: 320)
    height?: number; // Canvas height (default: 240)
    intersectionThreshold?: number; // Threshold for Intersection Observer (default: 0.01)
    backgroundRendering?: boolean; // Allow rendering when not visible (default: false)
}

export class VideoRenderer {
    #observeNode: VideoObserveNode;
    #context: VideoContext;

    constructor(context: VideoContext, init: VideoRendererInit = {}) {
        this.#context = context;
        // Create observe node for visibility monitoring
        this.#observeNode = new VideoObserveNode(context, {
            threshold: init.intersectionThreshold,
            enableBackground: init.backgroundRendering
        });
        // Connect observe node to destination
        this.#observeNode.connect(context.destination);
    }

    destroy() {
        this.#observeNode.dispose();
    }
}