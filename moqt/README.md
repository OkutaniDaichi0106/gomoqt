# MOQ protocol Implementation

## Overview

This package `moqt` provides a Go implementation of the MOQ protocol.

### Specification

This implementation follows the
[MOQ Lite specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html).
MOQ Lite focuses on essential features for real-time media streaming
applications, providing a streamlined approach compared to the full MOQT
specification.

## Features & Implementation Status

The following table lists features and tracks implementation progress against
the MOQ Lite specification sections:

| Section                                    | Implemented        | Tested             |
| ------------------------------------------ | ------------------ | ------------------ |
| **1. Data Model**                          |                    |                    |
| 1.1. Frame                                 | :white_check_mark: | :white_check_mark: |
| 1.2. Group                                 | :white_check_mark: | :white_check_mark: |
| 1.2.1 Group Priority Control               | :construction:     | :x:                |
| 1.3. Track                                 | :white_check_mark: | :white_check_mark: |
| 1.3.1. Track Priority Control              | :construction:     | :x:                |
| 1.4. Session                               | :white_check_mark: | :white_check_mark: |
| **2. Sessions**                            |                    |                    |
| 2.1. Session Establishment                 | :white_check_mark: | :white_check_mark: |
| 2.1.1. WebTransport                        | :white_check_mark: | :white_check_mark: |
| 2.1.2. QUIC                                | :white_check_mark: | :white_check_mark: |
| 2.1.2. Connection URL                      | :white_check_mark: | :white_check_mark: |
| 2.2. Extension Negocation                  | :white_check_mark: | :white_check_mark: |
| 2.2. Bitrate Monitoring                    | :white_check_mark: | :white_check_mark: |
| 2.5. Termination                           | :white_check_mark: | :white_check_mark: |
| 2.6. Migration                             | :construction:     | :x:                |
| **3. Roles**                               |                    |                    |
| 3.1. Subscriber                            | :white_check_mark: | :white_check_mark: |
| 3.1.1. Broadcast Discovery                 | :white_check_mark: | :white_check_mark: |
| 3.1.2. Track Subscription                  | :white_check_mark: | :white_check_mark: |
| 3.1.3. Graceful Publisher Relay Switchover | :x:                | :x:                |
| 3.2. Publisher                             | :white_check_mark: | :white_check_mark: |
| 3.2.1. Publication Announcement            | :white_check_mark: | :white_check_mark: |
| 3.2.2. Track Publishing                    | :white_check_mark: | :white_check_mark: |
| 3.2.3. Gap Notification                    | :x:                | :x:                |
| 3.2.4. Graceful Publisher Relay Switchover | :x:                | :x:                |
| 3.3. Relay                                 | :white_check_mark: | :white_check_mark: |
| 3.3.1. Routing Subscription                | :white_check_mark: | :white_check_mark: |
| **4. Security Considerations**             |                    |                    |
| 4.1. Resource Exhaustion                   | :x:                | :x:                |
| 4.1. Timeouts                              | :x:                | :x:                |

## References

- [MOQ Lite Specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)
- [QUIC Transport Protocol](https://tools.ietf.org/html/rfc9000)
- [WebTransport Protocol](https://tools.ietf.org/html/draft-ietf-webtrans-http3/)
- [CanIUse WebTransport](https://caniuse.com/webtransport)
