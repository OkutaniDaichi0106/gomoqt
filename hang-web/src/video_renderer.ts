import type { VideoDecodeStream } from "./internal";

export interface VideoRendererInit {
    source: VideoDecodeStream;
    canvas: HTMLCanvasElement;
    decoderConfig: VideoDecoderConfig;
    intersectionThreshold?: number; // Threshold for Intersection Observer (default: 0.01)
    backgroundRendering?: boolean; // Allow rendering when not visible (default: false)
}

export class VideoRenderer {
    canvas: HTMLCanvasElement;
    context: CanvasRenderingContext2D;
    source: VideoDecodeStream;
    delayFunc?: () => Promise<void>; // Custom interrupt function for delay
    observer?: IntersectionObserver;
    isVisible: boolean = true;
    animateId?: number;

    constructor(init: VideoRendererInit) {
        this.source = init.source;
        this.canvas = init.canvas;
        this.canvas.width = init.decoderConfig.codedWidth || 640; // Use config width if available
        this.canvas.height = init.decoderConfig.codedHeight || 480; // Use config height if available
        this.context = this.canvas.getContext('2d')!;

        // Configure the decoder with the provided config
        this.source.configure(init.decoderConfig);

        // Set up Intersection Observer if background rendering is disabled
        const threshold = init.intersectionThreshold ?? 0.01;
        const enableBackground = init.backgroundRendering ?? false;
        if (!enableBackground) {
            this.observer = new IntersectionObserver(
                (entries) => {
                    const entry = entries[0];
                    if (entry) {
                        this.isVisible = entry.isIntersecting;
                    }
                },
                { threshold }
            );
            this.observer.observe(init.canvas);
        } else {
            this.isVisible = true; // Always visible if background rendering enabled
        }

        // Set up the decode callback to render frames
        this.source.decodeTo({
            output : async ({done, value: frame}) => {
                if (done) {
                    return;
                }

                this.#schedule(frame);
            },
            error: (error: Error) => {
                console.error("VideoDecoder error:", error);
            },
        });

        // Start the rendering loop
        this.#schedule();
    }

    #schedule(frame?: VideoFrame) {
        // Cancel any previously scheduled frame
        if (this.animateId) cancelAnimationFrame(this.animateId);

        // Schedule the next frame
        this.animateId = requestAnimationFrame(() => this.#renderVideoFrame(frame));
    }

    async #renderVideoFrame(frame?: VideoFrame): Promise<void> {
        if (frame) {
            // Skip rendering if canvas is not visible
            if (!this.isVisible) {
                frame.close();
                return;
            }

            // Check if delay function is defined
            if (this.delayFunc) {
                console.log('Rendering delayed');
                await this.delayFunc();
            }

            // Clear the canvas
            this.context.clearRect(0, 0, this.canvas.width, this.canvas.height);

            // Draw the current frame
            this.context.drawImage(frame, 0, 0, this.canvas.width, this.canvas.height);

            // Close the original frame after processing
            frame.close();
        }
    }

    // Method to set custom delay function
    delay(func?: () => Promise<void>) {
        this.delayFunc = func;
    }

    destroy() {
        if (this.animateId) {
            cancelAnimationFrame(this.animateId);
        }
        this.observer?.disconnect();
        // Additional cleanup if needed
    }
}