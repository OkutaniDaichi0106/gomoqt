import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import * as index from './index';

describe('Main Index Module', () => {
    it('should be defined', () => {
        expect(index).toBeDefined();
        expect(typeof index).toBe('object');
    });

    it('should export expected public API', () => {
        // Verify that key exports are available
        expect(index).toHaveProperty('Session');
        expect(index).toHaveProperty('Client');
        expect(index).toHaveProperty('isValidBroadcastPath');
        expect(index).toHaveProperty('isValidPrefix');
        expect(index).toHaveProperty('TrackMux');
        expect(index).toHaveProperty('AnnouncementReader');
        expect(index).toHaveProperty('AnnouncementWriter');
    });

    it('should be importable', async () => {
        // This test verifies that the module can be imported without throwing
        await expect(import('./index')).resolves.toBeDefined();
    });

    it('should provide main MoQT functionality', () => {
        // Verify that the main classes and functions are exported
        expect(typeof index.Session).toBe('function'); // Constructor
        expect(typeof index.Client).toBe('function'); // Constructor
        expect(typeof index.isValidBroadcastPath).toBe('function');
        expect(typeof index.validateBroadcastPath).toBe('function');
        expect(typeof index.isValidPrefix).toBe('function');
        expect(typeof index.validateTrackPrefix).toBe('function');
        expect(typeof index.TrackMux).toBe('function'); // Constructor
    });

    it('should export stream types', () => {
        // Verify stream type exports
        expect(index).toHaveProperty('BiStreamTypes');
        expect(index).toHaveProperty('UniStreamTypes');
        expect(typeof index.BiStreamTypes).toBe('object');
        expect(typeof index.UniStreamTypes).toBe('object');
    });

    it('should export track and group stream classes', () => {
        // Verify track and group stream exports
        expect(index).toHaveProperty('TrackReader');
        expect(index).toHaveProperty('TrackWriter');
        expect(index).toHaveProperty('GroupReader');
        expect(index).toHaveProperty('GroupWriter');
        expect(typeof index.TrackReader).toBe('function');
        expect(typeof index.TrackWriter).toBe('function');
        expect(typeof index.GroupReader).toBe('function');
        expect(typeof index.GroupWriter).toBe('function');
    });

    it('should export subscribe stream classes', () => {
        // Verify subscribe stream exports
        expect(index).toHaveProperty('ReceiveSubscribeStream');
        expect(index).toHaveProperty('SendSubscribeStream');
        expect(typeof index.ReceiveSubscribeStream).toBe('function');
        expect(typeof index.SendSubscribeStream).toBe('function');
    });

    it('should export utility functions', () => {
        // Verify utility function exports
        expect(index).toHaveProperty('getExtension');
        expect(index).toHaveProperty('createBroadcastPath');
        expect(typeof index.getExtension).toBe('function');
        expect(typeof index.createBroadcastPath).toBe('function');
    });
});
