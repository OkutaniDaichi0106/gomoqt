/**
 * Type definition and utilities for track prefixes.
 * A TrackPrefix is a string that starts and ends with a forward slash '/'.
 */

/**
 * A track prefix is simply a string with specific format requirements.
 * Use validation functions to ensure correctness at runtime.
 */
export type TrackPrefix = string & { readonly brand: unique symbol };

/**
 * Runtime type guard to check if a string is a valid TrackPrefix.
 * @param path - The string to validate
 * @returns true if the path is a valid TrackPrefix
 */
export function isValidPrefix(path: string): boolean {
    return path.startsWith('/') && path.endsWith('/');
}

/**
 * Validates and returns a TrackPrefix, throwing an error if invalid.
 * @param path - The string to validate
 * @returns The validated TrackPrefix
 * @throws Error if the path is not a valid TrackPrefix
 */
export function validateTrackPrefix(path: string): TrackPrefix {
    if (!isValidPrefix(path)) {
        throw new Error(`Invalid track prefix: "${path}". Must start and end with "/"`);
    }
    return path as TrackPrefix;
}
