import { VideoTrackDecoder } from "./internal";

export interface VideoRendererInit {
    width?: number; // Canvas width (default: 320)
    height?: number; // Canvas height (default: 240)
    intersectionThreshold?: number; // Threshold for Intersection Observer (default: 0.01)
    backgroundRendering?: boolean; // Allow rendering when not visible (default: false)
}

export class VideoRenderer {
    canvas: HTMLCanvasElement;
    context: CanvasRenderingContext2D;
    #decoder?: VideoTrackDecoder;
    delayFunc?: () => Promise<void>; // Custom interrupt function for delay
    observer?: IntersectionObserver;
    isVisible: boolean = true;
    animateId?: number;

    constructor(init: VideoRendererInit = {}) {
        // Create canvas internally
        this.canvas = document.createElement('canvas');
        this.canvas.width = init.width ?? 320;
        this.canvas.height = init.height ?? 240;

        const context = this.canvas.getContext('2d');
        if (!context) {
            throw new Error('Failed to acquire 2D canvas context');
        }
        this.context = context;

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
            this.observer.observe(this.canvas);
        } else {
            this.isVisible = true; // Always visible if background rendering enabled
        }
    }

    async decoder(): Promise<VideoTrackDecoder> {
        // Clean up existing decoder if it exists
        if (!this.#decoder) {
            // Create new decoder
            this.#decoder = new VideoTrackDecoder({
                destination: new WritableStream<VideoFrame>({
                    write: (frame) => {
                        this.#schedule(frame);
                    }
                })
            });
        }

        return this.#decoder;
    }

    #schedule(frame?: VideoFrame) {
        // Cancel any previously scheduled frame
        if (this.animateId) cancelAnimationFrame(this.animateId);

        if (!frame) {
            this.animateId = undefined;
            return;
        }

        // Schedule the next frame
        this.animateId = requestAnimationFrame(() => this.#renderVideoFrame(frame));
    }

    async #renderVideoFrame(frame?: VideoFrame): Promise<void> {
        if (!frame) {
            return;
        }

        // Skip rendering if canvas is not visible
        if (!this.isVisible) {
            frame.close();
            return;
        }

        // Check if delay function is defined
        if (this.delayFunc) {
            console.log('Rendering delayed');
            try {
                await this.delayFunc();
            } catch (error) {
                console.warn('Error during rendering delay:', error);
            }
        }

        // Clear the canvas
        this.context.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Draw the current frame
        this.context.drawImage(frame, 0, 0, this.canvas.width, this.canvas.height);

        // Close the original frame after processing
        frame.close();
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
        
        // Clean up decoder
        if (this.#decoder) {
            try {
                this.#decoder.close();
            } catch (error) {
                console.warn('Error closing decoder during destroy:', error);
            }
            this.#decoder = undefined;
        }
    }
}