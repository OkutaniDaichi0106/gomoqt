import { isChrome, isFirefox } from './browser';

// Exported defaults so callers can reuse instead of recreating per-call.
export const DEFAULT_AUDIO_CODECS = ['opus', 'isac', 'g722', 'pcmu', 'pcma'] as const;

export const DEFAULT_AUDIO_CONFIG = {
    sampleRate: 48000,
    channels: 2,
    bitrate: 64000,
} as const;

// Options for audioEncoderConfig. All fields are optional so callers can pass only what
// they need; defaults are applied inside the function.
export interface AudioEncoderOptions {
    sampleRate: number; // caller must provide sample rate (e.g. 48000)
    channels: number; // caller must provide channel count (e.g. 2)
    bitrate?: number; // default: DEFAULT_AUDIO_CONFIG.bitrate
    preferredCodecs?: readonly string[]; // default: DEFAULT_AUDIO_CODECS
}

// Audio encoder config helper
// Returns a supported AudioEncoderConfig for common codecs (Opus preferred).
// Implements browser-specific tuning informed by MDN's AudioEncoder compatibility notes.
export async function audioEncoderConfig(options: AudioEncoderOptions): Promise<AudioEncoderConfig> {
    const sampleRate = options.sampleRate;
    const channels = options.channels;
    const targetBitrate = options.bitrate ?? DEFAULT_AUDIO_CONFIG.bitrate;
    const preferredCodecs = options.preferredCodecs ?? DEFAULT_AUDIO_CODECS;

    const base: AudioEncoderConfig = {
        codec: 'opus',
        sampleRate,
        numberOfChannels: channels,
        bitrate: targetBitrate,
    };

    // Try preferred codecs in order, applying browser-specific parameter tuning for Opus.
    for (const codec of preferredCodecs) {
        const cfg = upgradeAudioEncoderConfig(base, codec, targetBitrate);

        try {
            const audioEncoderCtor: any = AudioEncoder;
            // Some implementations may not expose isConfigSupported for AudioEncoder; guard with optional chaining.
            const res = await audioEncoderCtor.isConfigSupported?.(cfg);
            if (res && res.supported && res.config) {
                console.debug('using audio encoding:', res.config);
                return res.config;
            }
        } catch (err) {
            // ignore and try next codec
        }
    }

    throw new Error('no supported audio codec');
}

export function upgradeAudioEncoderConfig(base: AudioEncoderConfig, codec: string, bitrate?: number): AudioEncoderConfig {
    const cfg: AudioEncoderConfig = {
        ...base,
        codec,
    };

    if (typeof bitrate === 'number') cfg.bitrate = bitrate;

    // Browser-specific tuning for Opus
    if (codec === 'opus') {
    const anyCfg: any = cfg;

        // Prefer in-band FEC for robustness if not explicitly disabled.
        // Stereo flag: enable when numberOfChannels === 2.
        // application/signal hints: prefer 'voice' for mono/voice, 'music' for stereo/music.
        const isVoice = cfg.numberOfChannels === 1;

        // Default parameters object used by some UAs (Chrome, Edge, etc.). Use safe assignments.
        anyCfg.opus = anyCfg.opus ?? {};
        anyCfg.opus.application = anyCfg.opus.application ?? (isVoice ? 'voip' : 'audio');
        // 'signal' is an optional hint (e.g., 'voice' | 'music'), some browsers support it.
        anyCfg.opus.signal = anyCfg.opus.signal ?? (isVoice ? 'voice' : 'music');
        // Include some robustness and stereo hints where supported.
        anyCfg.parameters = anyCfg.parameters ?? {};
        if (anyCfg.parameters.useinbandfec === undefined) anyCfg.parameters.useinbandfec = 1;
        if (anyCfg.parameters.stereo === undefined) anyCfg.parameters.stereo = cfg.numberOfChannels === 2 ? 1 : 0;

        // bitrateMode is broadly supported; prefer variable for voice/music and constant for high quality streams.
    anyCfg.bitrateMode = (isChrome && !isFirefox) ? 'variable' : (anyCfg.bitrateMode ?? 'variable');
    }

    return cfg;
}
