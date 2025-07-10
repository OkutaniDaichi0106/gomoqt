import { Subscription } from './subscription';
import { BroadcastPath } from './broadcast_path';
import { SubscribeController } from './subscribe_stream';
import { TrackReader } from './track';

describe('Subscription', () => {
    it('should define the correct interface structure', () => {
        // This is a type-only test to ensure the interface is correctly defined
        const mockSubscription: Subscription = {
            broadcastPath: '/test/path' as BroadcastPath,
            trackName: 'test-track',
            controller: {} as SubscribeController,
            trackReader: {} as TrackReader
        };

        expect(mockSubscription.broadcastPath).toBe('/test/path');
        expect(mockSubscription.trackName).toBe('test-track');
        expect(mockSubscription.controller).toBeDefined();
        expect(mockSubscription.trackReader).toBeDefined();
    });

    it('should allow all required properties', () => {
        // Verify that all properties are required by creating a subscription
        const subscription: Subscription = {
            broadcastPath: '/example' as BroadcastPath,
            trackName: 'example-track',
            controller: {} as SubscribeController,
            trackReader: {} as TrackReader
        };

        // Type assertion to ensure all properties exist
        expect(typeof subscription.broadcastPath).toBe('string');
        expect(typeof subscription.trackName).toBe('string');
        expect(typeof subscription.controller).toBe('object');
        expect(typeof subscription.trackReader).toBe('object');
    });
});
