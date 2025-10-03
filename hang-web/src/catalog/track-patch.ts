import { TrackSchema } from "./track";
import type { TrackDescriptor } from "./track";
import { z } from "zod";

// This is a simplified version of JSON-Patch for Catalog
// Only supports "add", "remove", and "replace" operations on /tracks/{trackName} paths

export const AddTrackPatchSchema = z.object({
    op: z.literal("add"),
    path: z.string().startsWith("/tracks/"),
    value: TrackSchema,
});

export const RemoveTrackPatchSchema = z.object({
    op: z.literal("remove"),
    path: z.string().startsWith("/tracks/"),
});

export const ReplaceTrackPatchSchema = z.object({
    op: z.literal("replace"),
    path: z.string().startsWith("/tracks/"),
    value: TrackSchema,
});

export const TrackPatchSchema = z.union([
    AddTrackPatchSchema,
    RemoveTrackPatchSchema,
    ReplaceTrackPatchSchema,
]);

export type TrackPatch = z.infer<typeof TrackPatchSchema>;

// Compare two Track objects structurally. Treat missing optional fields and
// fields explicitly set to undefined as equivalent. Config is compared deeply.
export function isEqualTrack(a: TrackDescriptor, b: TrackDescriptor): boolean {
    return deepEqual(normalizeUndefined(a), normalizeUndefined(b));
}

// Create a shallow copy of an object with properties whose values are
// strictly undefined removed. This helps treat "missing" and "undefined"
// as equivalent in comparisons.
function normalizeUndefined<T extends object>(obj: T): T {
    // Only normalize plain objects (not arrays)
    if (obj === null || typeof obj !== "object" || Array.isArray(obj)) return obj;
    const out: any = {};
    for (const key of Object.keys(obj as any)) {
        const v = (obj as any)[key];
        if (v !== undefined) out[key] = v;
    }
    return out as T;
}

// A lightweight deep-equal specialized for JSON-like data (primitives, arrays,
// and plain objects). It intentionally does not handle functions, Maps, Sets,
// Dates, etc., because Tracks are JSON-serializable by design.
function deepEqual(a: any, b: any): boolean {
    if (a === b) return true;

    // Handle NaN
    if (typeof a === "number" && typeof b === "number" && Number.isNaN(a) && Number.isNaN(b)) return true;

    // Primitives and null
    if (a === null || b === null || typeof a !== "object" || typeof b !== "object") {
        return false;
    }

    // Arrays
    const aIsArray = Array.isArray(a);
    const bIsArray = Array.isArray(b);
    if (aIsArray || bIsArray) {
        if (!aIsArray || !bIsArray) return false;
        if (a.length !== b.length) return false;
        for (let i = 0; i < a.length; i++) {
            if (!deepEqual(a[i], b[i])) return false;
        }
        return true;
    }

    // Plain objects
    const aKeys = Object.keys(a).filter((k) => a[k] !== undefined).sort();
    const bKeys = Object.keys(b).filter((k) => b[k] !== undefined).sort();
    if (aKeys.length !== bKeys.length) return false;
    for (let i = 0; i < aKeys.length; i++) {
        if (aKeys[i] !== bKeys[i]) return false;
    }
    for (const k of aKeys) {
        if (!deepEqual(a[k], b[k])) return false;
    }
    return true;
}


export function diff(old: Map<string, TrackDescriptor>, curr: Map<string, TrackDescriptor>): TrackPatch[] {
    const patches: TrackPatch[] = [];

    // Find removed tracks
    for (const [name, track] of old) {
        if (!curr.has(name)) {
            patches.push({
                op: "remove",
                path: `/tracks/${name}`
            });
        }
    }

    // Find added or replaced tracks
    for (const [name, track] of curr) {
        if (!old.has(name)) {
            patches.push({
                op: "add",
                path: `/tracks/${name}`,
                value: track
            });
        } else {
            const oldTrack = old.get(name);
            // Replace only when the content is different (not just reference)
            if (oldTrack && !isEqualTrack(oldTrack, track)) {
                patches.push({
                    op: "replace",
                    path: `/tracks/${name}`,
                    value: track
                });
            }
        }
    }

    return patches;
}

export function merge(old: Map<string, TrackDescriptor>, patches: TrackPatch[]): Map<string, TrackDescriptor> {
    const result = new Map(old);

    for (const patch of patches) {
        switch (patch.op) {
            case "add":
                result.set(patch.path.split("/").pop()!, patch.value);
                break;
            case "remove":
                result.delete(patch.path.split("/").pop()!);
                break;
            case "replace":
                result.set(patch.path.split("/").pop()!, patch.value);
                break;
        }
    }

    return result;
}