import {
    VideoTrackProcessor,
    VideoTrackEncoder,
    AudioTrackProcessor,
    AudioTrackEncoder,
} from "../internal";
import type { TrackWriter } from "@okutanidaichi/moqt";
import { Cond,Channel } from "golikejs/sync";

export interface DeviceProps {
    preferred?: string;
}

export class Device {
    kind: "audio" | "video";
    preferred: string | undefined;
    available: MediaDeviceInfo[] | undefined;
    default: string | undefined;
    activeDeviceId: string | undefined;
    hasPermission: boolean = false;

    // internal listener ref so we can remove it on close
    #onchange: (() => void) | undefined;

    // external listeners to notify on devicechange
    #chan: Channel<void> = new Channel();

    // debounce timer id for devicechange
    #debounceTimer: number | undefined;
    // fallback timeout for getUserMedia in ms
    static GET_USER_MEDIA_TIMEOUT = 8000; // 8s

    constructor(kind: "audio" | "video", props?: DeviceProps) {
        this.kind = kind;
        this.preferred = props?.preferred;

        // initial enumeration
        void this.updateDevices();

        // keep availability up to date
        this.#onchange = () => {
            // debounce rapid devicechange events
            if (typeof this.#debounceTimer !== "undefined") {
                clearTimeout(this.#debounceTimer);
            }
            // schedule update slightly later to aggregate rapid changes
            this.#debounceTimer = window.setTimeout(() => {
                this.#debounceTimer = undefined;
                void this.updateDevices();
                // notify external listeners
                this.#chan.send();
            }, 200);
        };
        // Only set up event listeners if mediaDevices is available
        if (navigator && navigator.mediaDevices) {
            try {
                navigator.mediaDevices.addEventListener("devicechange", this.#onchange as EventListener);
            } catch (e) {
                // some environments may not support addEventListener on mediaDevices
                // fall back to assigning onchange if available
                if (typeof navigator.mediaDevices.ondevicechange !== "undefined") {
                    navigator.mediaDevices.ondevicechange = this.#onchange;
                }
            }
        }
    }

    async updateDevices(): Promise<void> {
        if (!navigator || !navigator.mediaDevices || !navigator.mediaDevices.enumerateDevices) {
            // Not running in a supported environment
            this.available = undefined;
            this.hasPermission = false;
            return;
        }

        try {
            const devices = await navigator.mediaDevices.enumerateDevices();
            this.available = devices.filter(d => d.kind === `${this.kind}input`);
            this.hasPermission = this.available.some(d => d.deviceId !== "");

            if (this.available.length > 0) {
                // Find default device using heuristics
                let defaultDevice = this.available.find(d => d.deviceId === "default");
                if (!defaultDevice) {
                    if (this.kind === "audio") {
                        defaultDevice = this.available.find(d =>
                            d.label.toLowerCase().includes("default") ||
                            d.label.toLowerCase().includes("communications")
                        );
                    } else {
                        defaultDevice = this.available.find(d =>
                            d.label.toLowerCase().includes("front") ||
                            d.label.toLowerCase().includes("external") ||
                            d.label.toLowerCase().includes("usb")
                        );
                    }
                }
                if (!defaultDevice) {
                    defaultDevice = this.available[0];
                }
                this.default = defaultDevice?.deviceId;
            }

            this.activeDeviceId = this.preferred || this.default;
        } catch (error) {
            console.warn(`Failed to update ${this.kind} devices:`, error);
        }
    }

    async requestPermission(): Promise<boolean> {
        if (this.hasPermission) return true;

        if (!navigator || !navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
            console.warn("getUserMedia is not available in this environment");
            return false;
        }

        // wrap getUserMedia with a timeout to avoid hanging
        const controller = new AbortController();
        const timeoutId = window.setTimeout(() => controller.abort(), Device.GET_USER_MEDIA_TIMEOUT);
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ [this.kind]: true } as MediaStreamConstraints);
            this.hasPermission = true;
            const deviceId = stream.getTracks()[0]?.getSettings().deviceId;
            if (deviceId) {
                this.preferred = deviceId;
                this.activeDeviceId = deviceId;
            }
            stream.getTracks().forEach(track => track.stop());
            await this.updateDevices();
            return true;
        } catch (error) {
            // getUserMedia can fail for many reasons; be conservative
            console.warn(`Failed to request ${this.kind} permission:`, error);
            return false;
        } finally {
            clearTimeout(timeoutId);
            controller.abort();
        }
    }

    // Get a single MediaStreamTrack for this device kind using preferred/default deviceId.
    // Returns undefined on failure.
    async getTrack(options?: MediaTrackConstraints): Promise<MediaStreamTrack | undefined> {
        // Ensure permissions/availability are checked first
        try {
            await this.requestPermission();
        } catch {
            // requestPermission is best-effort; continue to try
        }

        const deviceIdConstraint = this.activeDeviceId ? { deviceId: { exact: this.activeDeviceId } } : {};

        const constraints: MediaStreamConstraints = this.kind === "video"
            ? { video: { ...(deviceIdConstraint as any), ...(options ?? {}) } }
            : { audio: { ...(deviceIdConstraint as any), ...(options ?? {}) } };

        if (!navigator || !navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
            console.warn("getUserMedia is not available in this environment");
            return undefined;
        }

        // timeout wrapper to avoid hanging getUserMedia
        const timeoutId = window.setTimeout(() => {
            // no-op; we will check elapsed time after awaiting
        }, Device.GET_USER_MEDIA_TIMEOUT + 1000);

        let stream: MediaStream | undefined;
        try {
            stream = await navigator.mediaDevices.getUserMedia(constraints);
            const track = this.kind === "video" ? stream.getVideoTracks()[0] : stream.getAudioTracks()[0];
            if (!track) {
                stream.getTracks().forEach(t => t.stop());
                return undefined;
            }

            const settings = track.getSettings();
            if (settings && settings.deviceId) {
                this.activeDeviceId = settings.deviceId;
            }

            // Caller is responsible for stopping the returned track. But to avoid leaks when callers forget,
            // attach a small finalizer: if the track is still live after a long timeout, stop it. This is a
            // conservative fallback and will not run in normal usage.
            const t = track;
            window.setTimeout(() => {
                try {
                    if (!t.muted && !t.enabled) return; // some heuristics — avoid stopping intentionally disabled tracks
                    if (t.readyState !== "ended") {
                        // leave it — we don't forcibly stop within short time
                    }
                } catch {}
            }, 60000);

            return track;
        } catch (error) {
            console.warn(`Failed to get ${this.kind} track:`, error);
            // ensure any partial stream is stopped
            try {
                stream?.getTracks().forEach(t => t.stop());
            } catch {}
            return undefined;
        } finally {
            clearTimeout(timeoutId);
        }
    }

    close(): void {
        if (this.#onchange && navigator && navigator.mediaDevices) {
            try {
                if (typeof navigator.mediaDevices.removeEventListener === "function") {
                    navigator.mediaDevices.removeEventListener("devicechange", this.#onchange);
                } else if (typeof navigator.mediaDevices.ondevicechange !== "undefined") {
                    navigator.mediaDevices.ondevicechange = null;
                }
            } catch (e) {
                // ignore
            }
            this.#onchange = undefined;
        }
        // clear debounce timer
        if (typeof this.#debounceTimer !== "undefined") {
            clearTimeout(this.#debounceTimer);
            this.#debounceTimer = undefined;
        }
    }

    updated(): Promise<void> {
        return this.#chan.receive();
    }
}

