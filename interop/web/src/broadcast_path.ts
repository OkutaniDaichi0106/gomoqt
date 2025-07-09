/**
 * Type definition and utilities for broadcast paths.
 * A BroadcastPath is a string that starts with a forward slash '/'.
 */

/**
 * A broadcast path is simply a string with specific format requirements.
 * Use validation functions to ensure correctness at runtime.
 */
export type BroadcastPath = string;

/**
 * Runtime type guard to check if a string is a valid BroadcastPath.
 * @param path - The string to validate
 * @returns true if the path is a valid BroadcastPath
 */
export function isValidBroadcastPath(path: string): boolean {
    return path.startsWith('/') && path.length >= 1;
}

/**
 * Validates and returns a BroadcastPath, throwing an error if invalid.
 * @param path - The string to validate
 * @returns The validated BroadcastPath
 * @throws Error if the path is not a valid BroadcastPath
 */
export function validateBroadcastPath(path: string): BroadcastPath {
    if (!isValidBroadcastPath(path)) {
        throw new Error(`Invalid broadcast path: "${path}". Must start with "/"`);
    }
    return path;
}

/**
 * Extracts the file extension from a BroadcastPath.
 * @param path - The BroadcastPath to extract extension from
 * @returns The file extension including the dot (e.g., ".json", ".txt") or empty string if no extension
 * @example
 * getExtension("/alice.json") // returns ".json"
 * getExtension("/video/stream") // returns ""
 * getExtension("/file.min.js") // returns ".js"
 */
export function getExtension(path: BroadcastPath): string {
    const lastDot = path.lastIndexOf('.');
    const lastSlash = path.lastIndexOf('/');
    
    // If no dot found or dot is before the last slash (part of directory name), no extension
    if (lastDot === -1 || lastDot < lastSlash) {
        return '';
    }
    
    return path.substring(lastDot);
}

/**
 * Creates a BroadcastPath with compile-time and runtime validation.
 * @param path - The string to validate and convert to BroadcastPath
 * @returns A validated BroadcastPath
 * @throws Error if the path is not a valid BroadcastPath
 */
export function createBroadcastPath(path: string): BroadcastPath {
    if (!isValidBroadcastPath(path)) {
        throw new Error(`Invalid broadcast path: "${path}". Must start with "/"`);
    }
    return path as BroadcastPath;
}

/**
 * Template literal type helper for compile-time validation.
 * Use this when you know the path at compile time.
 */
export function broadcastPath<T extends string>(
    path: T extends `/${string}` ? T : T extends "/" ? T : never
): BroadcastPath {
    return path as unknown as BroadcastPath;
}

/**
 * Usage examples:
 * 
 * // Compile-time validation (will show error if invalid format):
 * const validPath = broadcastPath("/alice.json");     // ✅ OK
 * const invalidPath = broadcastPath("alice.json");    // ❌ Compile error
 * 
 * // Runtime validation (for dynamic strings):
 * const dynamicPath = createBroadcastPath(userInput); // ✅ OK with runtime check
 * 
 * // Type guard usage:
 * if (isValidBroadcastPath(someString)) {
 *     // someString is now typed as BroadcastPath
 * }
 */
