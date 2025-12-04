---
title: Errors
weight: 12
---

## General Error Variables

The following error variables are defined with the prefix `Err` and are used for general-purpose error handling:

| Variable Name           | Error Message               | Description (inferred)           |
|-------------------------|-----------------------------|----------------------------------|
| `moqt.ErrInvalidScheme` | "moqt: invalid scheme"      | Invalid scheme error             |
| `moqt.ErrClosedSession` | "moqt: closed session"      | Session has been closed          |
| `moqt.ErrServerClosed`  | "moqt: server closed"       | Server has been closed           |
| `moqt.ErrClientClosed`  | "moqt: client closed"       | Client has been closed           |

## Protocol Error Types

The following error types are defined to represent specific protocol error scenarios. Each type wraps error codes and provides methods for error handling and inspection:

| Error Type           | Description                                                      | Returned By                                      |
|----------------------|------------------------------------------------------------------|--------------------------------------------------|
| `moqt.SessionError`  | Error related to session management and protocol                 | `moqt.Session`, or instance originating from it  |
| `moqt.SubscribeError`| Error during subscribe negotiation or operation                  | `moqt.TrackWriter`, `moqt.TrackReader`           |
| `moqt.AnnounceError` | Error during announcement phase (e.g., broadcast path issues)    | `moqt.AnnouncementsWriter`, `moqt.AnnouncementsReader` |
| `moqt.GroupError`    | Error in group operations (e.g., out of range, expired group)    | `moqt.GroupWriter`, `moqt.GroupReader`           |

### Relationship with QUIC errors

The concrete MOQ error types map directly onto the QUIC error primitives:

- `moqt.SessionError` wraps `*quic.ApplicationError` and represents errors that affect the whole QUIC connection (session-level errors).
- `moqt.AnnounceError`, `moqt.SubscribeError`, and `moqt.GroupError` each wrap `*quic.StreamError` and represent errors that occur on individual QUIC streams (stream-level errors).

This mapping allows protocol-specific error types to be propagated over QUIC using the appropriate QUIC error mechanism.

Each error type implements the `error` interface and works with Go's standard error utilities (`errors.Is`, `errors.As`).


## Built-in Error Codes

Error codes for each custom error type are summarized below. Click each section to toggle visibility.

{{<tabs items="SessionErrorCode, AnnounceErrorCode, SubscribeErrorCode, GroupErrorCode" >}}
{{<tab>}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| `moqt.NoError`                      | 0x0   | Normal termination            |
| `moqt.InternalSessionErrorCode`     | 0x1   | Internal error                |
| `moqt.UnauthorizedSessionErrorCode` | 0x2   | Authentication/authorization  |
| `moqt.ProtocolViolationErrorCode`   | 0x3   | Protocol violation            |
| `moqt.ParameterLengthMismatchErrorCode` | 0x5 | Parameter length mismatch     |
| `moqt.TooManySubscribeErrorCode`    | 0x6   | Too many subscriptions        |
| `moqt.GoAwayTimeoutErrorCode`       | 0x10  | GoAway timeout                |
| `moqt.UnsupportedVersionErrorCode`  | 0x12  | Unsupported version           |
| `moqt.SetupFailedErrorCode`         | 0x13  | Setup failed                  |
{{< /tab >}}

{{< tab >}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| `moqt.InternalAnnounceErrorCode`    | 0x0   | Internal error                |
| `moqt.DuplicatedAnnounceErrorCode`  | 0x1   | Duplicated broadcast path     |
| `moqt.InvalidAnnounceStatusErrorCode` | 0x2 | Invalid announce status       |
| `moqt.UninterestedErrorCode`        | 0x3   | Uninterested                  |
| `moqt.BannedPrefixErrorCode`        | 0x4   | Banned prefix                 |
| `moqt.InvalidPrefixErrorCode`       | 0x5   | Invalid prefix                |
{{< /tab >}}


{{< tab >}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| `moqt.InternalSubscribeErrorCode`   | 0x00  | Internal error                |
| `moqt.InvalidRangeErrorCode`        | 0x01  | Invalid range                 |
| `moqt.DuplicateSubscribeIDErrorCode`| 0x02  | Duplicate subscribe ID        |
| `moqt.TrackNotFoundErrorCode`       | 0x03  | Track not found               |
| `moqt.UnauthorizedSubscribeErrorCode` | 0x04 | Unauthorized                  |
| `moqt.SubscribeTimeoutErrorCode`    | 0x05  | Subscribe timeout             |
{{< /tab >}}


{{< tab >}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| `moqt.InternalGroupErrorCode`       | 0x00  | Internal error                |
| `moqt.OutOfRangeErrorCode`          | 0x02  | Out of range                  |
| `moqt.ExpiredGroupErrorCode`        | 0x03  | Expired group                 |
| `moqt.SubscribeCanceledErrorCode`   | 0x04  | Subscribe canceled            |
| `moqt.PublishAbortedErrorCode`      | 0x05  | Publish aborted               |
| `moqt.ClosedSessionGroupErrorCode`  | 0x06  | Closed session                |
| `moqt.InvalidSubscribeIDErrorCode`  | 0x07  | Invalid subscribe ID          |
{{< /tab >}}
{{< /tabs >}}

## Error Handling

Implementations in `gomoqt/moqt` return specific error types for different error scenarios. You can use type assertions to handle these errors accordingly.

- **Example: When `moqt.TrackWriter` returns an error**

```go
    var subErr *moqt.SubscribeError
    if errors.As(err, &subErr) {
        // Handle SubscribeError
    }
```

> [!NOTE] Note:
> MOQ-related errors are always returned from specific structs. When analyzing errors, make sure to perform error handling and analysis at the correct location in the code, according to the struct that returns the error. This ensures accurate diagnosis and handling of protocol errors.

## Error Propagation

You can get `context.Context` via `Context` method implementated in `gomoqt/moqt` such as `moqt.TrackReader` or `moqt.Session`.
`moqt.Cause` function is provided to access to the root cause of an error propagation and to parse the cause  as a MOQ error if it is a QUIC error.
This is because the `context.Context` holds the original QUIC error and `context.Cause` returns the cause as is.
```go
func Cause(ctx context.Context) error
```

**Example: Get MOQ cause from context.Context**

```go
    var ctx context.Context

    var cause error
    cause = moqt.Cause(ctx)
```

To get the MOQ cause from a context, use `errors.As` function with corresponding error type.

**Example: When `moqt.TrackWriter`'s context is canceled**

```go
    var tw *moqt.TrackWriter
    ctx := tw.Context()

    if err := moqt.Cause(ctx); err != nil {
        var subErr *moqt.SubscribeError
        if errors.As(err, &subErr) {
            // Handle SubscribeError
        }

        var sessErr *moqt.SessionError
        if errors.As(err, &sessErr) {
            // Handle SessionError
        }
    }


```

> [!NOTE] Note: context.Cause
> When using `context.Cause`, the raw QUIC error can be accessed directly from the context.