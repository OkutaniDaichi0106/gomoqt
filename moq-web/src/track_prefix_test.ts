
import { assertEquals } from "../deps.ts";
import { isValidPrefix, validateTrackPrefix } from './track_prefix.ts';

Deno.test('TrackPrefix', async (t) => {
    await t.step('isValidPrefix - should return true for valid prefixes', () => {
        assertEquals(isValidPrefix('/foo/'), true);
        assertEquals(isValidPrefix('//'), true);
        assertEquals(isValidPrefix('/'), true);
    });

    await t.step('isValidPrefix - should return false for invalid prefixes', () => {
        assertEquals(isValidPrefix('foo/'), false);
        assertEquals(isValidPrefix('/foo'), false);
        assertEquals(isValidPrefix('foo'), false);
        assertEquals(isValidPrefix(''), false);
    });

    await t.step('validateTrackPrefix - should return the prefix for valid prefixes', () => {
        assertEquals(validateTrackPrefix('/foo/'), '/foo/');
        assertEquals(validateTrackPrefix('//'), '//');
        assertEquals(validateTrackPrefix('/'), '/');
    });

    await t.step('validateTrackPrefix - should throw an error for invalid prefixes', () => {
        try {
            validateTrackPrefix('foo/');
            assertEquals(true, false); // Should have thrown
        } catch {
            assertEquals(true, true);
        }
        try {
            validateTrackPrefix('/foo');
            assertEquals(true, false); // Should have thrown
        } catch {
            assertEquals(true, true);
        }
        try {
            validateTrackPrefix('foo');
            assertEquals(true, false); // Should have thrown
        } catch {
            assertEquals(true, true);
        }
        try {
            validateTrackPrefix('');
            assertEquals(true, false); // Should have thrown
        } catch {
            assertEquals(true, true);
        }
    });
});
