---
title: Terminate
weight: 11
---


## Terminate a Session

Use the `Session.Terminate` method to explicitly close a session in response to protocol violations, internal errors, authentication failures, or when immediate resource release is needed. After calling this method, all streams associated with the session are closed. If termination is already in progress, subsequent calls are ignored.

```go
func (s *Session) Terminate(code SessionErrorCode, msg string) error
```
- `code`: Error code indicating the reason for termination (`SessionErrorCode`)
- `msg`: Human-readable message describing the reason

**Termination with No Error**:
```go
    session.Terminate(moq.NoErrorCode, "no error")
```

**Termination with Reserved Reason**:
```go
    session.Terminate(moq.ProtocolViolationErrorCode, "unexpected stream type")
```

**Termination with Custom Error**:
```go
    var code SessionErrorCode = 0x1001
    session.Terminate(code, "custom error message")
```

For most use cases, prefer using a reserved error code for clarity. If you need to indicate a custom reason, you can use your own code.


## Close or Shut Down Client

The client can terminate all its active sessions and shut down.

{{<cards >}}
    {{< card link="../client/#terminate-and-shut-down-session" title="Terminate and Shut Down Session" icon="server" subtitle="How to terminate a session." >}}
{{</ cards >}}


## Closing / Shutting Down Server

The server can also terminate all active sessions and shut down.

{{<cards >}}
    {{< card link="../server/#terminate-and-shut-down-session" title="Terminate and Shut Down Session" icon="server" subtitle="How to terminate a session." >}}
{{</ cards >}}

## Error Codes & Reasons

Some error codes are reserved for specific termination reasons. For clarity and interoperability, use the most appropriate reserved code whenever possible:

- `NoErrorCode`: Normal termination (no error)
- `InternalSessionErrorCode`: Internal error (More specific error code is recommended)
- `UnauthorizedSessionErrorCode`: Authentication/authorization error
- `ProtocolViolationErrorCode`: Protocol violation
- `ParameterLengthMismatchErrorCode`: Parameter length mismatch
- `TooManySubscribeErrorCode`: Too many subscriptions
- `GoAwayTimeoutErrorCode`: GoAway timeout
- `UnsupportedVersionErrorCode`: Unsupported version

For example, to terminate a session due to an internal error:
```go
session.Terminate(moq.InternalSessionErrorCode, "internal server error")
```
