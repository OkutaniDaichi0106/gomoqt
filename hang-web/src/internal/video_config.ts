import { isFirefox } from './browser';

// Export codec lists so callers can reuse and avoid reallocation.
export const VIDEO_HARDWARE_CODECS = [
	'vp09.00.10.08',
	'vp09',
	'avc1.640028',
	'avc1.4D401F',
	'avc1.42E01E',
	'avc1',
	'av01.0.08M.08',
	'av01',
	'hev1.1.6.L93.B0',
	'hev1',
	'vp8',
] as const;

export const VIDEO_SOFTWARE_CODECS = [
	'avc1.640028',
	'avc1.4D401F',
	'avc1.42E01E',
	'avc1',
	'vp8',
	'vp09.00.10.08',
	'vp09',
	'hev1.1.6.L93.B0',
	'hev1',
	'av01.0.08M.08',
	'av01',
] as const;

export interface VideoEncoderOptions {
	width: number; // caller must provide width (px)
	height: number; // caller must provide height (px)
	frameRate: number; // caller must provide frame rate (fps)
	bitrate?: number; // if provided, overrides calculated bitrate
	tryHardware?: boolean; // default true
}

export async function videoEncoderConfig(options: VideoEncoderOptions): Promise<VideoEncoderConfig> {
	const width = options.width;
	const height = options.height;
	const frameRate = options.frameRate;
	const tryHardware = options.tryHardware ?? true;
	const hardwareCodecs = VIDEO_HARDWARE_CODECS;
	const softwareCodecs = VIDEO_SOFTWARE_CODECS;

	// TARGET BITRATE CALCULATION (h264)
	const pixels = width * height;
	const framerateFactor = 30.0 + (frameRate - 30) / 2;
	const calculatedBitrate = Math.round(pixels * 0.07 * framerateFactor);
	const bitrate = options?.bitrate ?? calculatedBitrate;

	const baseConfig: VideoEncoderConfig = {
		codec: 'none',
		width,
		height,
		bitrate,
		latencyMode: 'realtime',
		framerate: frameRate,
	};

	if (tryHardware && !isFirefox) {
		for (const codec of hardwareCodecs) {
			const config = upgradeEncoderConfig(baseConfig, codec, bitrate, true);
			const { supported, config: hardwareConfig } = await VideoEncoder.isConfigSupported(config);
			if (supported && hardwareConfig) {
				console.debug('using hardware encoding: ', hardwareConfig);
				return hardwareConfig;
			}
		}
	} else if (tryHardware && isFirefox) {
		console.warn('Cannot detect hardware encoding on Firefox.');
	}

	for (const codec of softwareCodecs) {
		const config = upgradeEncoderConfig(baseConfig, codec, bitrate, false);
		const { supported, config: softwareConfig } = await VideoEncoder.isConfigSupported(config);
		if (supported && softwareConfig) {
			console.debug('using software encoding: ', softwareConfig);
			return softwareConfig;
		}
	}

	throw new Error('no supported codec');
}

export function upgradeEncoderConfig(base: VideoEncoderConfig, codec: string, bitrate: number, hardware: boolean): VideoEncoderConfig {
	const config: VideoEncoderConfig = {
		...base,
		codec,
		hardwareAcceleration: hardware ? 'prefer-hardware' : undefined,
	};

	if (config.codec.startsWith('avc1')) {
		config.avc = { format: 'annexb' };
	} else if (config.codec.startsWith('hev1')) {
		// @ts-expect-error Typescript needs to be updated.
		config.hevc = { format: 'annexb' };
	} else if (config.codec.startsWith('vp09')) {
		config.bitrate = bitrate * 0.8;
	} else if (config.codec.startsWith('av01')) {
		config.bitrate = bitrate * 0.6;
	} else if (config.codec === 'vp8') {
		config.bitrate = bitrate * 1.1;
	}

	return config;
}
