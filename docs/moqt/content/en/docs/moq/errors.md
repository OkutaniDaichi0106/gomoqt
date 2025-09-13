---
title: Errors
weight: 12
---

## General Error Variables

The following error variables are defined with the prefix `Err` and are used for general-purpose error handling:

| Variable Name      | Error Message                | Description (inferred)           |
|--------------------|-----------------------------|----------------------------------|
| ErrInvalidScheme   | "moqt: invalid scheme"      | Invalid scheme error             |
| ErrInvalidRange    | "moqt: invalid range"       | Invalid range error              |
| ErrClosedSession   | "moqt: closed session"      | Session has been closed          |
| ErrServerClosed    | "moqt: server closed"       | Server has been closed           |
| ErrClientClosed    | "moqt: client closed"       | Client has been closed           |

## Protocol Error Types

The following error types are defined to represent specific protocol error scenarios. Each type wraps error codes and provides methods for error handling and inspection:

| Error Type           | Description                                                      | Returned By                                      |
|----------------------|------------------------------------------------------------------|--------------------------------------------------|
| `moqt.SessionError`  | Error related to session management and protocol                 | `moqt.Session`, or instance originating from it  |
| `moqt.SubscribeError`| Error during subscribe negotiation or operation                  | `moqt.TrackWriter`, `moqt.TrackReader`           |
| `moqt.AnnounceError` | Error during announcement phase (e.g., broadcast path issues)    | `moqt.AnnouncementsWriter`, `moqt.AnnouncementsReader` |
| `moqt.GroupError`    | Error in group operations (e.g., out of range, expired group)    | `moqt.GroupWriter`, `moqt.GroupReader`           |

### Relationship with QUIC Errors

`moqt.SessionError` occurs when a `quic.ApplicationError` occurs, which is transmitted on the QUIC Connection, representing errors that affect the entire session.

`moqt.AnnounceError`, `moqt.SubscribeError`, and `moqt.GroupError` occur when a `quic.StreamError` occurs on individual QUIC Streams, representing errors that occur within those streams.

This design allows protocol-specific error types to be mapped directly to the appropriate QUIC error mechanism, ensuring accurate error propagation and handling at both the connection and stream levels.

Each error type implements the `error` interface. They are compatible with Go's standard error handling (`errors.Is`, `errors.As`).

### Built-in Error Codes

Error codes for each custom error type are summarized below. Click each section to toggle visibility.

{{% details title="SessionErrorCode" closed="true" %}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| NoError                      | 0x0   | Normal termination            |
| InternalSessionErrorCode     | 0x1   | Internal error                |
| UnauthorizedSessionErrorCode | 0x2   | Authentication/authorization  |
| ProtocolViolationErrorCode   | 0x3   | Protocol violation            |
| ParameterLengthMismatchErrorCode | 0x5 | Parameter length mismatch     |
| TooManySubscribeErrorCode    | 0x6   | Too many subscriptions        |
| GoAwayTimeoutErrorCode       | 0x10  | GoAway timeout                |
| UnsupportedVersionErrorCode  | 0x12  | Unsupported version           |
| SetupFailedErrorCode         | 0x13  | Setup failed                  |
{{% /details %}}

{{% details title="AnnounceErrorCode" closed="true" %}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| InternalAnnounceErrorCode    | 0x0   | Internal error                |
| DuplicatedAnnounceErrorCode  | 0x1   | Duplicated broadcast path     |
| InvalidAnnounceStatusErrorCode | 0x2 | Invalid announce status       |
| UninterestedErrorCode        | 0x3   | Uninterested                  |
| BannedPrefixErrorCode        | 0x4   | Banned prefix                 |
| InvalidPrefixErrorCode       | 0x5   | Invalid prefix                |
{{% /details %}}

{{% details title="SubscribeErrorCode" closed="true" %}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| InternalSubscribeErrorCode   | 0x00  | Internal error                |
| InvalidRangeErrorCode        | 0x01  | Invalid range                 |
| DuplicateSubscribeIDErrorCode| 0x02  | Duplicate subscribe ID        |
| TrackNotFoundErrorCode       | 0x03  | Track not found               |
| UnauthorizedSubscribeErrorCode | 0x04 | Unauthorized                  |
| SubscribeTimeoutErrorCode    | 0x05  | Subscribe timeout             |
{{% /details %}}

{{% details title="GroupErrorCode" closed="true" %}}
| Code Name                    | Value | Description                    |
|------------------------------|-------|-------------------------------|
| InternalGroupErrorCode       | 0x00  | Internal error                |
| OutOfRangeErrorCode          | 0x02  | Out of range                  |
| ExpiredGroupErrorCode        | 0x03  | Expired group                 |
| SubscribeCanceledErrorCode   | 0x04  | Subscribe canceled            |
| PublishAbortedErrorCode      | 0x05  | Publish aborted               |
| ClosedSessionGroupErrorCode  | 0x06  | Closed session                |
| InvalidSubscribeIDErrorCode  | 0x07  | Invalid subscribe ID          |
{{% /details %}}

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
    var ctx context.Context

    var cause error
    cause = moqt.Cause(ctx)
```

To get the MOQ cause from a context, use `errors.As` function with corresponding error type.

**Example: When `moqt.TrackWriter`'s context is canceled**

```go
    var tw *moqt.TrackWriter
    ctx := tw.Context()

    var subErr *moqt.SubscribeError
    if errors.As(err, &subErr) {
        // Handle SubscribeError
    }

    var sessErr *moqt.SessionError
    if errors.As(err, &sessErr) {
        // Handle SessionError
    }
```

> [!NOTE] Note: context.Cause
> When using `context.Cause`, the raw QUIC error can be accessed directly from the context.