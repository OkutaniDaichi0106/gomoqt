export interface VirtualContent {
    backgroundColor?: string; // Default: 'black'
    textColor?: string; // Default: 'white'
    title?: string; // Default: 'No Camera'
    subtitle?: string; // Default: 'Available'
    fontSize?: number; // Default: 24
    customDraw?: (ctx: CanvasRenderingContext2D, width: number, height: number) => void;
}

export interface VideoPreviewerInit {
    source: Promise<MediaStreamTrack>;
    width?: number; // Canvas width (default: 640)
    height?: number; // Canvas height (default: 480)
    intersectionThreshold?: number; // Threshold for Intersection Observer (default: 0.01)
    backgroundRendering?: boolean; // Allow rendering when not visible (default: false)
    virtualContent?: VirtualContent; // Custom virtual content configuration
}

/**
 * VideoPreviewer provides video preview functionality with support for both real and virtual video sources.
 * Similar to VideoRenderer but for preview/encoding scenarios where camera might not be available.
 */
export class VideoPreviewer {
    canvas: HTMLCanvasElement;
    context: CanvasRenderingContext2D;
    #source: Promise<MediaStreamTrack>;
    #isVirtual: boolean = true; // Start as virtual until real source is available
    animationId?: number;
    destroyed: boolean = false;
    delayFunc?: () => Promise<void>; // Custom interrupt function for delay
    observer?: IntersectionObserver;
    isVisible: boolean = true;
    virtualContent: VirtualContent;

    constructor(init: VideoPreviewerInit) {
        this.#source = init.source
        this.canvas = document.createElement('canvas');
        this.canvas.width = init.width ?? 640;
        this.canvas.height = init.height ?? 480;
        const context = this.canvas.getContext('2d');
        if (!context) {
            throw new Error('Failed to acquire 2D canvas context');
        }
        this.context = context;

        // Set up virtual content configuration with defaults
        this.virtualContent = {
            backgroundColor: init.virtualContent?.backgroundColor ?? 'black',
            textColor: init.virtualContent?.textColor ?? 'white',
            title: init.virtualContent?.title ?? 'No Camera',
            subtitle: init.virtualContent?.subtitle ?? 'Available',
            fontSize: init.virtualContent?.fontSize ?? 24,
            customDraw: init.virtualContent?.customDraw
        };

        // Set up canvas dimensions
        this.canvas.width = 640;
        this.canvas.height = 480;

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

        // Start preview loop and handle source resolution
        this.#startPreview();
        this.#source.then(realSource => {
            this.#isVirtual = false;
            console.log('Switched to real video source');
        }).catch((error: unknown) => {
            console.warn('Failed to get real video source, staying virtual:', error);
        });
    }

    #startPreview(): void {
        let video: HTMLVideoElement | undefined;
        
        const animate = async () => {
            if (this.destroyed) return;
            
            if (!this.isVisible) {
                this.animationId = requestAnimationFrame(animate);
                return;
            }

            if (this.delayFunc) {
                try {
                    await this.delayFunc();
                } catch (error) {
                    console.warn('Delay function threw error:', error);
                }
            }

            // Create video element for real source if needed
            if (!this.#isVirtual && !video) {
                try {
                    const track = await this.#source;
                    video = document.createElement('video');
                    video.srcObject = new MediaStream([track]);
                    video.autoplay = true;
                    video.muted = true;
                } catch (error) {
                    console.warn('Failed to create video element:', error);
                    this.#isVirtual = true; // Fallback to virtual
                }
            }
            
            try {
                // Clear canvas
                this.context.clearRect(0, 0, this.canvas.width, this.canvas.height);
                
                if (this.#isVirtual) {
                    // Draw virtual content directly
                    if (this.virtualContent.customDraw) {
                        this.virtualContent.customDraw(this.context, this.canvas.width, this.canvas.height);
                    } else {
                        // Draw background with configurable color
                        this.context.fillStyle = this.virtualContent.backgroundColor!;
                        this.context.fillRect(0, 0, this.canvas.width, this.canvas.height);
                        
                        // Draw static text with configurable properties
                        this.context.fillStyle = this.virtualContent.textColor!;
                        this.context.font = `${this.virtualContent.fontSize}px Arial`;
                        this.context.textAlign = 'center';
                        this.context.fillText(this.virtualContent.title!, this.canvas.width / 2, this.canvas.height / 2 - 10);
                        if (this.virtualContent.subtitle) {
                            this.context.fillText(this.virtualContent.subtitle, this.canvas.width / 2, this.canvas.height / 2 + 20);
                        }
                    }
                } else if (video && video.readyState >= 2) {
                    // Draw real video content
                    this.context.drawImage(video, 0, 0, this.canvas.width, this.canvas.height);
                }
            } catch (error) {
                console.warn('Error rendering preview frame:', error);
            }
            
            this.animationId = requestAnimationFrame(animate);
        };
        
        animate();
    }

    // Method to set custom delay function
    delay(func?: () => Promise<void>) {
        this.delayFunc = func;
    }

    destroy(): void {
        if (this.destroyed) return;
        
        this.destroyed = true;
        
        // Cancel animation frame
        if (this.animationId) {
            cancelAnimationFrame(this.animationId);
            this.animationId = undefined;
        }
        
        // Cleanup observer
        this.observer?.disconnect();
        this.observer = undefined;
        
        // Stop source track (best effort)
        this.#source.then(track => {
            if (track.readyState === 'live') {
                track.stop();
            }
        }).catch(() => {}); // Ignore errors
        

    }
}

// export interface VirtualVideoTrack {
//     track: MediaStreamTrack;
//     stop(): void;
// }

// /**
//  * Create a virtual video track that works well with VideoEncoder
//  * Returns both the track and a stop function to clean up the animation
//  */
// export function createVirtualVideoTrack(virtualContent?: VirtualContent): VirtualVideoTrack {
//     const canvas = document.createElement('canvas');
//     canvas.width = 640;
//     canvas.height = 480;
    
//     const ctx = canvas.getContext('2d');
//     if (!ctx) {
//         throw new Error('Failed to create canvas context');
//     }
    
//     let stopped = false;
    
//     // Default virtual content
//     const content: VirtualContent = {
//         backgroundColor: virtualContent?.backgroundColor ?? 'black',
//         textColor: virtualContent?.textColor ?? 'white',
//         title: virtualContent?.title ?? 'No Camera',
//         subtitle: virtualContent?.subtitle ?? 'Available',
//         fontSize: virtualContent?.fontSize ?? 24,
//         customDraw: virtualContent?.customDraw
//     };
    
//     // Draw static content once (no animation needed)
//     if (content.customDraw) {
//         content.customDraw(ctx, canvas.width, canvas.height);
//     } else {
//         // Draw static content with configurable properties
//         ctx.fillStyle = content.backgroundColor!;
//         ctx.fillRect(0, 0, canvas.width, canvas.height);
        
//         ctx.fillStyle = content.textColor!;
//         ctx.font = `${content.fontSize}px Arial`;
//         ctx.textAlign = 'center';
//         ctx.fillText(content.title!, canvas.width / 2, canvas.height / 2 - 10);
//         if (content.subtitle) {
//             ctx.fillText(content.subtitle, canvas.width / 2, canvas.height / 2 + 20);
//         }
//     }
    
//     const stream = canvas.captureStream(15);
//     const track = stream.getVideoTracks()[0];
//     if (!track) {
//         throw new Error("Failed to create virtual video track");
//     }
    
//     return {
//         track,
//         stop() {
//             if (stopped) return;
//             stopped = true;
            
//             if (track.readyState === 'live') {
//                 track.stop();
//             }
//         }
//     };
// }
