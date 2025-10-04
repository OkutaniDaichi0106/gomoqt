import { diff } from "./track-patch";
import type { TrackDescriptor} from "./track";
import { describe, test, expect, beforeEach } from 'vitest';

function track(partial: Partial<TrackDescriptor> & { name: string }): TrackDescriptor {
  // Provide sane defaults for required fields
  return {
    name: partial.name,
    description: partial.description,
    priority: partial.priority ?? 0,
    schema: partial.schema ?? "schema://default",
    config: partial.config ?? {},
    dependencies: partial.dependencies,
  };
}

describe("catalog/track-patch diff", () => {
  test("no changes -> no patches", () => {
    const a = new Map<string, TrackDescriptor>([
      ["t1", track({ name: "t1", priority: 1, schema: "s1", config: { a: 1 } })],
      ["t2", track({ name: "t2", priority: 2, schema: "s2", config: { b: 2 }, dependencies: ["t1"] })],
    ]);
    const b = new Map<string, TrackDescriptor>([
      ["t1", track({ name: "t1", priority: 1, schema: "s1", config: { a: 1 } })],
      ["t2", track({ name: "t2", priority: 2, schema: "s2", config: { b: 2 }, dependencies: ["t1"] })],
    ]);

    const patches = diff(a, b);
    expect(patches).toEqual([]);
  });

  test("add track", () => {
    const a = new Map<string, TrackDescriptor>([
      ["t1", track({ name: "t1", priority: 1, schema: "s1" })],
    ]);
    const b = new Map<string, TrackDescriptor>([
      ["t1", track({ name: "t1", priority: 1, schema: "s1" })],
      ["t2", track({ name: "t2", priority: 5, schema: "s2", config: { x: true } })],
    ]);

    const patches = diff(a, b);
    expect(patches).toEqual([
      { op: "add", path: "/tracks/t2", value: b.get("t2")! },
    ]);
  });

  test("remove track", () => {
    const a = new Map<string, TrackDescriptor>([
      ["t1", track({ name: "t1", priority: 1, schema: "s1" })],
      ["t2", track({ name: "t2", priority: 5, schema: "s2", config: { x: true } })],
    ]);
    const b = new Map<string, TrackDescriptor>([
      ["t1", track({ name: "t1", priority: 1, schema: "s1" })],
    ]);

    const patches = diff(a, b);
    expect(patches).toEqual([
      { op: "remove", path: "/tracks/t2" },
    ]);
  });

  test("replace when content differs", () => {
    const a = new Map<string, TrackDescriptor>([[
      "t1", track({ name: "t1", priority: 1, schema: "s1", config: { a: 1 } }),
    ]]);
    const b = new Map<string, TrackDescriptor>([[
      "t1", track({ name: "t1", priority: 2, schema: "s1", config: { a: 1, b: 2 } }),
    ]]);

    const patches = diff(a, b);
    expect(patches).toEqual([
      { op: "replace", path: "/tracks/t1", value: b.get("t1")! },
    ]);
  });

  test("no replace for reference change only", () => {
    const a = new Map<string, TrackDescriptor>([[
      "t1", track({ name: "t1", priority: 1, schema: "s1", config: { a: 1 } }),
    ]]);
    // New object with the same content
    const b = new Map<string, TrackDescriptor>([[
      "t1", track({ name: "t1", priority: 1, schema: "s1", config: { a: 1 } }),
    ]]);

    const patches = diff(a, b);
    expect(patches).toEqual([]);
  });

  test("treat undefined and missing as equal for optional fields", () => {
    const a = new Map<string, TrackDescriptor>([[
      "t1", {
        name: "t1",
        // description intentionally omitted
        priority: 0,
        schema: "s",
        config: {},
        // dependencies intentionally omitted
      },
    ]]);
    const b = new Map<string, TrackDescriptor>([[
      "t1", {
        name: "t1",
        description: undefined,
        priority: 0,
        schema: "s",
        config: {},
        dependencies: undefined,
      },
    ]]);

    const patches = diff(a, b);
    expect(patches).toEqual([]);
  });
});
