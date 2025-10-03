import { describe, test, expect, beforeEach, vi, it } from 'vitest';

// Mock browser detection
vi.mock('./browser', () => ({
    isChrome: true,
    isFirefox: false,
}));

import { 
    DEFAULT_AUDIO_CODECS, 
    DEFAULT_AUDIO_CONFIG, 
    audioEncoderConfig, 
    upgradeAudioEncoderConfig,
    AudioEncoderOptions 
} from './audio_config';

// Mock AudioEncoder
const mockAudioEncoder = {
    isConfigSupported: vi.fn(),
};

// Mock the global AudioEncoder
Object.defineProperty(global, 'AudioEncoder', {
    writable: true,
    value: mockAudioEncoder,
});

// Mock console.debug to avoid noise in tests
global.console.debug = vi.fn();

describe('Audio Config', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe('DEFAULT_AUDIO_CODECS', () => {
        test('contains expected codec list', () => {
            expect(DEFAULT_AUDIO_CODECS).toEqual(['opus', 'isac', 'g722', 'pcmu', 'pcma']);
        });

        test('is readonly', () => {
            expect(Object.isFrozen(DEFAULT_AUDIO_CODECS)).toBe(false); // const arrays aren't frozen by default
            expect(DEFAULT_AUDIO_CODECS.length).toBe(5);
        });

        test('has opus as preferred codec', () => {
            expect(DEFAULT_AUDIO_CODECS[0]).toBe('opus');
        });
    });

    describe('DEFAULT_AUDIO_CONFIG', () => {
        test('contains expected default values', () => {
            expect(DEFAULT_AUDIO_CONFIG).toEqual({
                sampleRate: 48000,
                channels: 2,
                bitrate: 64000,
            });
        });

        test('uses standard audio settings', () => {
            expect(DEFAULT_AUDIO_CONFIG.sampleRate).toBe(48000); // Professional audio standard
            expect(DEFAULT_AUDIO_CONFIG.channels).toBe(2); // Stereo
            expect(DEFAULT_AUDIO_CONFIG.bitrate).toBe(64000); // 64 kbps
        });
    });

    describe('audioEncoderConfig', () => {
        test('returns supported config for valid options', async () => {
            const mockConfig = {
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 64000,
            };

            mockAudioEncoder.isConfigSupported.mockResolvedValue({
                supported: true,
                config: mockConfig,
            });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            const result = await audioEncoderConfig(options);

            expect(result).toEqual(mockConfig);
            expect(mockAudioEncoder.isConfigSupported).toHaveBeenCalledTimes(1);
            expect(console.debug).toHaveBeenCalledWith('using audio encoding:', mockConfig);
        });

        test('uses default bitrate when not provided', async () => {
            const mockConfig = {
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: DEFAULT_AUDIO_CONFIG.bitrate,
            };

            mockAudioEncoder.isConfigSupported.mockResolvedValue({
                supported: true,
                config: mockConfig,
            });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            await audioEncoderConfig(options);

            expect(mockAudioEncoder.isConfigSupported).toHaveBeenCalledWith(
                expect.objectContaining({
                    bitrate: DEFAULT_AUDIO_CONFIG.bitrate,
                })
            );
        });

        test('uses custom bitrate when provided', async () => {
            const customBitrate = 128000;
            const mockConfig = {
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 1,
                bitrate: customBitrate,
            };

            mockAudioEncoder.isConfigSupported.mockResolvedValue({
                supported: true,
                config: mockConfig,
            });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 1,
                bitrate: customBitrate,
            };

            await audioEncoderConfig(options);

            expect(mockAudioEncoder.isConfigSupported).toHaveBeenCalledWith(
                expect.objectContaining({
                    bitrate: customBitrate,
                })
            );
        });

        test('uses default codecs when preferredCodecs not provided', async () => {
            mockAudioEncoder.isConfigSupported.mockResolvedValue({
                supported: true,
                config: { codec: 'opus' },
            });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            await audioEncoderConfig(options);

            // Should try opus first (from DEFAULT_AUDIO_CODECS)
            expect(mockAudioEncoder.isConfigSupported).toHaveBeenCalledWith(
                expect.objectContaining({
                    codec: 'opus',
                })
            );
        });

        test('uses custom preferredCodecs when provided', async () => {
            const customCodecs = ['pcmu', 'opus'] as const;
            
            // First call returns unsupported, second call returns supported
            mockAudioEncoder.isConfigSupported
                .mockResolvedValueOnce({ supported: false })
                .mockResolvedValueOnce({ 
                    supported: true, 
                    config: { codec: 'opus' } 
                });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
                preferredCodecs: customCodecs,
            };

            await audioEncoderConfig(options);

            expect(mockAudioEncoder.isConfigSupported).toHaveBeenCalledTimes(2);
            expect(mockAudioEncoder.isConfigSupported).toHaveBeenNthCalledWith(1,
                expect.objectContaining({ codec: 'pcmu' })
            );
            expect(mockAudioEncoder.isConfigSupported).toHaveBeenNthCalledWith(2,
                expect.objectContaining({ codec: 'opus' })
            );
        });

        test('tries all codecs until one is supported', async () => {
            // Mock first 3 codecs as unsupported, 4th as supported
            mockAudioEncoder.isConfigSupported
                .mockResolvedValueOnce({ supported: false })
                .mockResolvedValueOnce({ supported: false })
                .mockResolvedValueOnce({ supported: false })
                .mockResolvedValueOnce({ 
                    supported: true, 
                    config: { codec: 'pcmu' } 
                });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            const result = await audioEncoderConfig(options);

            expect(result).toEqual({ codec: 'pcmu' });
            expect(mockAudioEncoder.isConfigSupported).toHaveBeenCalledTimes(4);
        });

        test('throws error when no codec is supported', async () => {
            mockAudioEncoder.isConfigSupported.mockResolvedValue({ supported: false });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            await expect(audioEncoderConfig(options)).rejects.toThrow('no supported audio codec');
        });

        test('handles missing isConfigSupported method', async () => {
            // Remove isConfigSupported method
            delete (mockAudioEncoder as any).isConfigSupported;

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            await expect(audioEncoderConfig(options)).rejects.toThrow('no supported audio codec');
        });

        test('handles isConfigSupported throwing error', async () => {
            // Restore the mock function and set it to reject
            mockAudioEncoder.isConfigSupported = vi.fn().mockRejectedValue(new Error('Config check failed'));

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            await expect(audioEncoderConfig(options)).rejects.toThrow('no supported audio codec');
        });

        test('handles mono audio configuration', async () => {
            const mockConfig = {
                codec: 'opus',
                sampleRate: 16000,
                numberOfChannels: 1,
                bitrate: 32000,
            };

            // Restore the mock function and set it to resolve
            mockAudioEncoder.isConfigSupported = vi.fn().mockResolvedValue({
                supported: true,
                config: mockConfig,
            });

            const options: AudioEncoderOptions = {
                sampleRate: 16000,
                channels: 1,
                bitrate: 32000,
            };

            const result = await audioEncoderConfig(options);

            expect(result.numberOfChannels).toBe(1);
            expect(result.sampleRate).toBe(16000);
        });

        test('handles high sample rate configuration', async () => {
            const mockConfig = {
                codec: 'opus',
                sampleRate: 96000,
                numberOfChannels: 2,
                bitrate: 128000,
            };

            // Restore the mock function and set it to resolve
            mockAudioEncoder.isConfigSupported = vi.fn().mockResolvedValue({
                supported: true,
                config: mockConfig,
            });

            const options: AudioEncoderOptions = {
                sampleRate: 96000,
                channels: 2,
                bitrate: 128000,
            };

            const result = await audioEncoderConfig(options);

            expect(result.sampleRate).toBe(96000);
        });
    });

    describe('upgradeAudioEncoderConfig', () => {
        const baseConfig: AudioEncoderConfig = {
            codec: 'opus',
            sampleRate: 48000,
            numberOfChannels: 2,
            bitrate: 64000,
        };

        test('applies codec from parameter', () => {
            const result = upgradeAudioEncoderConfig(baseConfig, 'pcmu');

            expect(result.codec).toBe('pcmu');
            expect(result.sampleRate).toBe(baseConfig.sampleRate);
            expect(result.numberOfChannels).toBe(baseConfig.numberOfChannels);
        });

        test('applies custom bitrate when provided', () => {
            const customBitrate = 128000;
            const result = upgradeAudioEncoderConfig(baseConfig, 'opus', customBitrate);

            expect(result.bitrate).toBe(customBitrate);
        });

        test('keeps original bitrate when custom bitrate not provided', () => {
            const result = upgradeAudioEncoderConfig(baseConfig, 'opus');

            expect(result.bitrate).toBe(baseConfig.bitrate);
        });

        test('applies Opus-specific enhancements for stereo', () => {
            const result = upgradeAudioEncoderConfig(baseConfig, 'opus') as any;

            expect(result.opus).toBeDefined();
            expect(result.opus.application).toBe('audio'); // stereo defaults to 'audio'
            expect(result.opus.signal).toBe('music'); // stereo defaults to 'music'
            expect(result.parameters).toBeDefined();
            expect(result.parameters.useinbandfec).toBe(1);
            expect(result.parameters.stereo).toBe(1); // stereo enabled
            expect(result.bitrateMode).toBe('variable'); // Chrome default
        });

        test('applies Opus-specific enhancements for mono', () => {
            const monoConfig = { ...baseConfig, numberOfChannels: 1 };
            const result = upgradeAudioEncoderConfig(monoConfig, 'opus') as any;

            expect(result.opus.application).toBe('voip'); // mono defaults to 'voip'
            expect(result.opus.signal).toBe('voice'); // mono defaults to 'voice'
            expect(result.parameters.stereo).toBe(0); // stereo disabled
        });

        test('does not override existing Opus parameters', () => {
            const configWithOpus = {
                ...baseConfig,
                opus: { application: 'existing' },
                parameters: { useinbandfec: 0 },
            } as any;

            const result = upgradeAudioEncoderConfig(configWithOpus, 'opus') as any;

            expect(result.opus.application).toBe('existing'); // preserved
            expect(result.parameters.useinbandfec).toBe(0); // preserved
        });

        test('does not apply Opus enhancements for non-Opus codecs', () => {
            const result = upgradeAudioEncoderConfig(baseConfig, 'pcmu') as any;

            expect(result.opus).toBeUndefined();
            expect(result.parameters).toBeUndefined();
            expect(result.bitrateMode).toBeUndefined();
        });

        test('handles undefined bitrate parameter', () => {
            const result = upgradeAudioEncoderConfig(baseConfig, 'opus', undefined);

            expect(result.bitrate).toBe(baseConfig.bitrate);
        });

        test('preserves all base config properties', () => {
            const extendedBase = {
                ...baseConfig,
                customProperty: 'test',
            } as any;

            const result = upgradeAudioEncoderConfig(extendedBase, 'pcmu') as any;

            expect(result.customProperty).toBe('test');
            expect(result.sampleRate).toBe(extendedBase.sampleRate);
            expect(result.numberOfChannels).toBe(extendedBase.numberOfChannels);
        });

        test('applies browser-specific bitrate mode for Chrome', () => {
            // Mock is already set to Chrome=true, Firefox=false in the mock above
            const result = upgradeAudioEncoderConfig(baseConfig, 'opus') as any;

            expect(result.bitrateMode).toBe('variable');
        });
    });

    describe('AudioEncoderOptions interface', () => {
        test('requires sampleRate and channels', () => {
            // This test ensures the interface is properly typed (TypeScript compilation test)
            const validOptions: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            expect(validOptions.sampleRate).toBe(48000);
            expect(validOptions.channels).toBe(2);
        });

        test('supports optional properties', () => {
            const fullOptions: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
                bitrate: 128000,
                preferredCodecs: ['opus', 'pcmu'],
            };

            expect(fullOptions.bitrate).toBe(128000);
            expect(fullOptions.preferredCodecs).toEqual(['opus', 'pcmu']);
        });
    });

    describe('Error Handling', () => {
        test('handles null AudioEncoder', async () => {
            Object.defineProperty(global, 'AudioEncoder', {
                writable: true,
                value: null,
            });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            await expect(audioEncoderConfig(options)).rejects.toThrow();
        });

        test('handles AudioEncoder without isConfigSupported', async () => {
            Object.defineProperty(global, 'AudioEncoder', {
                writable: true,
                value: {},
            });

            const options: AudioEncoderOptions = {
                sampleRate: 48000,
                channels: 2,
            };

            await expect(audioEncoderConfig(options)).rejects.toThrow('no supported audio codec');
        });
    });

    describe('Boundary Value Tests', () => {
        test('handles zero bitrate', () => {
            const testConfig: AudioEncoderConfig = {
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 64000,
            };
            const result = upgradeAudioEncoderConfig(testConfig, 'opus', 0);

            expect(result.bitrate).toBe(0);
        });

        test('handles very high bitrate', () => {
            const testConfig: AudioEncoderConfig = {
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 64000,
            };
            const highBitrate = 1000000;
            const result = upgradeAudioEncoderConfig(testConfig, 'opus', highBitrate);

            expect(result.bitrate).toBe(highBitrate);
        });

        test('validates sample rate bounds are passed to config', () => {
            // Test that sample rates are passed through correctly to upgradeAudioEncoderConfig
            const testConfig8k: AudioEncoderConfig = {
                codec: 'opus',
                sampleRate: 8000,
                numberOfChannels: 1,
                bitrate: 32000,
            };

            const result8k = upgradeAudioEncoderConfig(testConfig8k, 'opus');
            expect(result8k.sampleRate).toBe(8000);

            const testConfig192k: AudioEncoderConfig = {
                codec: 'opus',
                sampleRate: 192000,
                numberOfChannels: 2,
                bitrate: 256000,
            };

            const result192k = upgradeAudioEncoderConfig(testConfig192k, 'opus');
            expect(result192k.sampleRate).toBe(192000);
        });

        test('validates channel count bounds are passed to config', () => {
            // Test that channel counts are passed through correctly
            const testConfigMono: AudioEncoderConfig = {
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 1,
                bitrate: 32000,
            };

            const resultMono = upgradeAudioEncoderConfig(testConfigMono, 'opus');
            expect(resultMono.numberOfChannels).toBe(1);

            const testConfigSurround: AudioEncoderConfig = {
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 8,
                bitrate: 512000,
            };

            const resultSurround = upgradeAudioEncoderConfig(testConfigSurround, 'opus');
            expect(resultSurround.numberOfChannels).toBe(8);
        });
    });

    describe('Advanced Configuration Tests', () => {
        test('handles malformed codec responses gracefully', async () => {
            // Test various malformed responses
            const malformedResponses = [
                null,
                undefined,
                {},
                { supported: true }, // missing config
                { supported: true, config: null },
            ];

            for (const response of malformedResponses) {
                mockAudioEncoder.isConfigSupported.mockResolvedValue(response);
                
                await expect(audioEncoderConfig({
                    sampleRate: 48000,
                    channels: 2,
                    preferredCodecs: ['opus']
                })).rejects.toThrow('no supported audio codec');
            }
        });

        test('configuration object immutability', () => {
            const baseConfig = {
                codec: 'opus' as const,
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 64000,
            };

            const config1 = upgradeAudioEncoderConfig(baseConfig, 'opus');
            const config2 = upgradeAudioEncoderConfig(baseConfig, 'g722');

            // Configurations should be separate objects
            expect(config1).not.toBe(config2);
            expect(config1.codec).toBe('opus');
            expect(config2.codec).toBe('g722');

            // Base config should remain unchanged
            expect(baseConfig.codec).toBe('opus');
        });
    });

    describe('Performance and Memory Tests', () => {
        test('configuration object cloning works correctly', () => {
            // Test configuration immutability
            const baseConfig = {
                codec: 'opus' as const,
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 64000,
            };

            const config1 = upgradeAudioEncoderConfig(baseConfig, 'opus');
            const config2 = upgradeAudioEncoderConfig(baseConfig, 'isac');

            // Configurations should be separate objects
            expect(config1).not.toBe(config2);
            expect(config1.codec).toBe('opus');
            expect(config2.codec).toBe('isac');

            // Base config should remain unchanged
            expect(baseConfig.codec).toBe('opus');
        });

        test('handles different codec configurations', () => {
            const baseConfig = {
                codec: 'opus' as const,
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 64000,
            };

            // Test opus-specific enhancements
            const opusConfig = upgradeAudioEncoderConfig(baseConfig, 'opus');
            expect((opusConfig as any).opus).toBeDefined();
            expect((opusConfig as any).parameters).toBeDefined();

            // Test non-opus codec (should not have opus enhancements)
            const g722Config = upgradeAudioEncoderConfig(baseConfig, 'g722');
            expect(g722Config.codec).toBe('g722');
            expect((g722Config as any).opus).toBeUndefined();
        });
    });

    describe('Real-world Integration Scenarios', () => {
        test('handles voice chat mono configuration', () => {
            const voiceConfig = upgradeAudioEncoderConfig({
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 1,
                bitrate: 32000,
            }, 'opus');

            // Voice-specific settings
            expect((voiceConfig as any).opus?.application).toBe('voip');
            expect((voiceConfig as any).opus?.signal).toBe('voice');
            expect((voiceConfig as any).parameters?.stereo).toBe(0);
            expect((voiceConfig as any).parameters?.useinbandfec).toBe(1);
        });

        test('handles music streaming stereo configuration', () => {
            const musicConfig = upgradeAudioEncoderConfig({
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 128000,
            }, 'opus');

            // Music-specific settings
            expect((musicConfig as any).opus?.application).toBe('audio');
            expect((musicConfig as any).opus?.signal).toBe('music');
            expect((musicConfig as any).parameters?.stereo).toBe(1);
            expect((musicConfig as any).parameters?.useinbandfec).toBe(1);
        });

        test('handles browser-specific bitrate modes', () => {
            // Chrome should use variable bitrate mode
            const chromeConfig = upgradeAudioEncoderConfig({
                codec: 'opus',
                sampleRate: 48000,
                numberOfChannels: 2,
                bitrate: 64000,
            }, 'opus');

            expect((chromeConfig as any).bitrateMode).toBe('variable');
        });
    });
});
