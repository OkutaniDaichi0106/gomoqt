import * as index from './index';

describe('Main Index Module', () => {
    it('should be defined', () => {
        expect(index).toBeDefined();
        expect(typeof index).toBe('object');
    });

    it('should export expected public API', () => {
        // Verify that key exports are available
        expect(index).toHaveProperty('Session');
        expect(index).toHaveProperty('dial');
        expect(index).toHaveProperty('isValidBroadcastPath');
        expect(index).toHaveProperty('isValidPrefix');
        expect(index).toHaveProperty('MOQOptions');
    });

    it('should be importable', () => {
        // This test verifies that the module can be imported without throwing
        expect(() => require('./index')).not.toThrow();
    });

    it('should provide main MoQT functionality', () => {
        // Verify that the main classes and functions are exported
        expect(typeof index.Session).toBe('function'); // Constructor
        expect(typeof index.dial).toBe('function');
        expect(typeof index.isValidBroadcastPath).toBe('function');
        expect(typeof index.validateBroadcastPath).toBe('function');
        expect(typeof index.isValidPrefix).toBe('function');
        expect(typeof index.validateTrackPrefix).toBe('function');
    });
});
