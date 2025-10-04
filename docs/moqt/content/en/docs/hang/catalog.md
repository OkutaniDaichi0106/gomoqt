---
title: Catalog 🚧
weight: 1
---

A Catalog is a manifest-like metadata object a publisher provides as part of the media stream so that a client can discover what media is available before (or without) opening individual subscriptions.
It answers:
What tracks exist?
How are they grouped?
What parameters (priority, grouping cadence, naming) apply?
A Catalog typically lists:

- Track identifiers (names)
- Track attributes (e.g. priority, content type, group period)
- Relationships (e.g. variants, layers, alt audio, caption tracks)
- Optional hints for prefetch, caching, or subscription strategy

Using a Catalog avoids trial‑and‑error subscription attempts.

> [!TL;DR]
> Catalogs are JSON objects published on a `catalog.json` track that describe the available tracks and their metadata. Subscribe to `catalog.json` to discover media. For now, prefer republishing the full catalog when making changes.

> [!NOTE] Experimental
> `gomoqt` currently emits its own (unstable) structure.
> Expect breaking changes until a standard emerges.
> Unknown fields MUST be ignored for forward compatibility.

## Catalog Track

Catalogs are typically published as a special track named `catalog.json` within a broadcast.
To access the catalog, subscribe to this `catalog.json` track and receive the catalog.
This track contains a single JSON object conforming to the schema described below.

## Root Object (catalog.json)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| version | uint62 | No | Format version for optimistic evolution. (Default: 1) |
| description | string (<=500) | No | Human‑readable summary. |
| tracks | map<string, Track> | Yes | Track definitions keyed by name. |

## Track Object

Following fields are defined for each track:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string (non‑empty) | Yes | Track identifier (key MUST match). |
| description | string (<=500) | No | Human description. |
| priority | uint8 (0‑255) | Yes | Relative selection priority (larger often means more important). |
| schema | string | Yes | schema id: `video`, `audio`, ... or URI. |
| config | object | Yes | Schema‑specific config (validated per schema). |
| dependencies | string[] | No | Other track names this track depends on (e.g. captions -> video). |

### Schema‑Specific Config

Below each schema has its own config shape. These are just examples and may evolve.

{{% details title="Video Schema (schema=\"video\")" closed="true" %}}
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| codec | string | Yes | Codec string (e.g. avc1, hvc1, vp09). |
| description | string | No | Extra description. |
| codedWidth | uint53 | No | Encoded frame width. |
| codedHeight | uint53 | No | Encoded frame height. |
| displayAspectWidth | uint53 | No | Display aspect numerator. |
| displayAspectHeight | uint53 | No | Display aspect denominator. |
| framerate | uint53 | No | Approx frames per second. |
| bitrate | uint53 | No | Target / observed bitrate (bps). |
| optimizeForLatency | boolean | No | Hint to favor latency over quality. (Default: true) |
| rotation | number | No | Rotation degrees CW. (Default: 0) |
| flip | boolean | No | Horizontal flip hint. (Default: false) |
| container | `loc` or `cmaf` | Yes | Packaging container. |
{{% /details %}}

{{% details title="Audio Schema (schema=\"audio\")" closed="true" %}}
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| codec | string | Yes | Codec (e.g. opus, mp4a.40.2). |
| description | string | No | Extra description. |
| sampleRate | uint53 | Yes | Sampling rate (Hz). |
| numberOfChannels | uint53 | Yes | Channel count. |
| bitrate | uint53 | No | Target / observed bitrate. |
| container | `loc` or `cmaf` | Yes | Packaging container. |
{{% /details %}}

{{% details title="Captions Schema (schema=\"captions\")" closed="true" %}}
Captions typically require `dependencies` referencing at least one media track (audio or video).
{{% /details %}}

{{% details title="Location Schema (schema=\"location\")" closed="true" %}}
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | uint53 | Yes | Location identifier. |
| name | string | Yes | Display name. |
| avatar | url | Yes | Avatar image URL. |
{{% /details %}}

{{% details title="User Schema (schema=\"user\")" closed="true" %}}
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | uuid | Yes | User id (UUID). |
| name | string | Yes | Display name. |
| avatar | url | Yes | Avatar image URL. |
{{% /details %}}

{{% details title="Timeseries Schema (schema=\"timeseries\")" closed="true" %}}
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| measurements | map<string, Measurement> | Yes | Named measurement descriptors. |

Measurement object:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | string | Yes | Measurement type identifier. |
| unit | string | Yes | Unit string (e.g. ms, dB). |
| interval | uint53 | Yes | Expected sampling interval (ms). |
| min | uint53 | No | Lower bound. |
| max | uint53 | No | Upper bound. |
{{% /details %}}

## `catalog.json` Examples

{{% details title="Web Camera" closed="false" %}}
```json
{
  "version": 1,
  "description": "Example catalog",
  "tracks": {
    "camera-main": {
      "name": "camera-main",
      "priority": 128,
      "schema": "video",
      "config": {
        "codec": "avc1.640028",
        "framerate": 30,
        "container": "cmaf"
      }
    },
    "mic": {
      "name": "mic",
      "priority": 200,
      "schema": "audio",
      "config": {
        "codec": "opus",
        "sampleRate": 48000,
        "numberOfChannels": 2,
        "container": "loc"
      }
    }
  }
}
```
{{% /details %}}

{{% details title="SVC (Scalable Video Coding)" closed="true" %}}
This example models spatial (and implicitly temporal) layers using multiple `video` tracks. Each enhancement layer depends on its lower layer via `dependencies`. The semantics of SVC are application‑level: the MOQ layer only sees independent tracks plus dependencies metadata.

Key points:

- `camera-main-L0` is the base layer (lowest resolution) – highest priority so receivers at constrained bandwidth can still request it first.
- Enhancement layers (`L1`, `L2`) declare `dependencies` on the immediately lower layer (a receiver MAY ignore those it cannot afford).
- Codecs are identical; differing `codedWidth`, `codedHeight`, and `bitrate` hint scalability steps.

```json
{
  "version": 1,
  "description": "SVC catalog example (3 spatial layers)",
  "tracks": {
    "camera-main-L0": {
      "name": "camera-main-L0",
      "priority": 200,
      "schema": "video",
      "config": {
        "codec": "avc1.640028",
        "codedWidth": 640,
        "codedHeight": 360,
        "framerate": 30,
        "bitrate": 400000,
        "container": "cmaf"
      }
    },
    "camera-main-L1": {
      "name": "camera-main-L1",
      "priority": 150,
      "schema": "video",
      "dependencies": ["camera-main-L0"],
      "config": {
        "codec": "avc1.640028",
        "codedWidth": 1280,
        "codedHeight": 720,
        "framerate": 30,
        "bitrate": 1200000,
        "container": "cmaf"
      }
    },
    "camera-main-L2": {
      "name": "camera-main-L2",
      "priority": 120,
      "schema": "video",
      "dependencies": ["camera-main-L1"],
      "config": {
        "codec": "avc1.640028",
        "codedWidth": 1920,
        "codedHeight": 1080,
        "framerate": 30,
        "bitrate": 2500000,
        "container": "cmaf"
      }
    },
    "mic": {
      "name": "mic",
      "priority": 210,
      "schema": "audio",
      "config": {
        "codec": "opus",
        "sampleRate": 48000,
        "numberOfChannels": 2,
        "bitrate": 96000,
        "container": "loc"
      }
    }
  }
}
```

Receiver strategy example:

1. Subscribe base layer `camera-main-L0` and audio `mic` immediately.
2. If bandwidth OK, add `camera-main-L1`; if congestion increases, drop L2 first, then L1.
3. Dependencies allow a relay to pre‑plan cache / forwarding order (base > enhancements).

> [!NOTE] Application Semantics
> The SVC meaning of `dependencies` is a convention; MOQ transport does not enforce decoding order. Implementations SHOULD verify lower layers are present before requesting higher ones.

{{% /details %}}

{{% details title="IoT Drone Telemetry (timeseries)" closed="true" %}}
This example shows an IoT / drone telemetry track using the `timeseries` schema to multiplex several sensor measurements in a single track. A single subscription yields all listed metrics; applications can downsample or filter client‑side.

Design notes:

- One `timeseries` track (`drone-telemetry`) instead of many tiny tracks minimizes subscription signaling.
- Each entry in `measurements` describes expected interval and bounds to help receivers allocate buffers / dashboards.
- Units are illustrative; choose consistent SI units in production.

```json
{
  "version": 1,
  "description": "Drone telemetry catalog example",
  "tracks": {
    "drone-telemetry": {
      "name": "drone-telemetry",
      "priority": 180,
      "schema": "timeseries",
      "config": {
        "measurements": {
          "temperature": { "type": "temperature", "unit": "celsius", "interval": 1000, "min": -40, "max": 85 },
          "humidity": { "type": "humidity", "unit": "percent", "interval": 1500, "min": 0, "max": 100 },
          "altitude": { "type": "altitude", "unit": "meter", "interval": 500, "min": -100, "max": 5000 },
          "latitude": { "type": "latitude", "unit": "degree", "interval": 1000, "min": -90, "max": 90 },
          "longitude": { "type": "longitude", "unit": "degree", "interval": 1000, "min": -180, "max": 180 },
          "battery_voltage": { "type": "battery_voltage", "unit": "volt", "interval": 2000, "min": 0, "max": 30 }
        }
      }
    }
  }
}
```

Subscription strategy:

- Critical dashboards subscribe only `drone-telemetry` early (priority 180).
- If bandwidth constrained, receiver can locally thin high‑frequency fields (e.g. keep every 2nd altitude sample) without renegotiation.
- Additional media (video feed) could be another track; telemetry remains unaffected.

> [!NOTE] Bounds
> `min`/`max` are informative for UI range scaling; producers MAY omit them. Receivers MUST NOT enforce them as hard validation unless policy demands.

{{% /details %}}

## Patch
Catalog changes can be expressed as JSON Patch (RFC 6902) — a compact sequence of operations (add, remove, replace, move, copy, test) that describe a diff against a previous catalog instance.

**Example (JSON Patch array)**:

```json
[
    { "op": "replace", "path": "/tracks/camera-main/priority", "value": 220 },
    { "op": "add", "path": "/tracks/backup-mic", "value": {
        "name": "backup-mic",
        "priority": 150,
        "schema": "audio",
        "config": { "codec": "opus", "sampleRate": 48000, "numberOfChannels": 1, "container": "loc" }
        }
    }
]
```
> [!NOTE] Note: Patch Support
> Patch updates are not yet supported by this implementation; producers SHOULD republish the full `catalog.json` until Patch support is added.


## Validation & Evolution

Subscribers SHOULD:

- Ignore unknown top‑level or per‑schema fields.
- Treat missing optional fields as default values above.
- Not assume numeric ranges beyond those stated.

Producers SHOULD increment `version` only on incompatible format changes.

> [!NOTE] Standardization Status
> A unified Catalog format is not standardized yet[^1] [^2]; disparate implementations (including this one) may diverge[^3]. Alignment work is ongoing.

[^1]: IETF, [Datatracker - Common Catalog Format for moq-transport](https://datatracker.ietf.org/doc/draft-ietf-moq-catalogformat/)
[^2]: kixelated, [Internet Draft - Media over QUIC - Hang#Catalog](https://www.ietf.org/archive/id/draft-lcurley-moq-hang-00.html#name-catalog)
[^3]: kixelated, [GitHub - moq/js/hang/catalog](https://github.com/kixelated/moq/tree/dd3cffc868ca2e0513c785e5adedf5448400f555/js/hang/src/catalog)