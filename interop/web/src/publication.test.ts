import { Publication } from './publication';
import { BroadcastPath } from './broadcast_path';
import { PublishController } from './subscribe_stream';
import { TrackWriter } from './track';

describe('Publication', () => {
    it('should define the correct type structure', () => {
        // This is a type-only test to ensure the interface is correctly defined
        const mockPublication: Publication = {
            broadcastPath: '/test/path' as BroadcastPath,
            trackName: 'test-track',
            controller: {} as PublishController,
            trackWriter: {} as TrackWriter
        };

        expect(mockPublication.broadcastPath).toBe('/test/path');
        expect(mockPublication.trackName).toBe('test-track');
        expect(mockPublication.controller).toBeDefined();
        expect(mockPublication.trackWriter).toBeDefined();
    });

    it('should allow all required properties', () => {
        // Verify that all properties are required by creating a publication
        const publication: Publication = {
            broadcastPath: '/example' as BroadcastPath,
            trackName: 'example-track',
            controller: {} as PublishController,
            trackWriter: {} as TrackWriter
        };

        // Type assertion to ensure all properties exist
        expect(typeof publication.broadcastPath).toBe('string');
        expect(typeof publication.trackName).toBe('string');
        expect(typeof publication.controller).toBe('object');
        expect(typeof publication.trackWriter).toBe('object');
    });
});
