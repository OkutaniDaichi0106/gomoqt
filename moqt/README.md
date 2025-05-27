# MOQ Lite Implementation

## Overview

This package provides a Go implementation of the MOQ Lite specification for Media over QUIC Transport. MOQ Lite is a simplified version of the Media over QUIC Transport protocol, designed for lower latency and reduced complexity while maintaining the core benefits of QUIC-based media delivery.

## Specification

This implementation follows the [MOQ Lite specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html) (draft-lcurley-moq-transfork). MOQ Lite focuses on essential features for real-time media streaming applications, providing a streamlined approach compared to the full MOQT specification.

### Key Features of MOQ Lite

- **Simplified Track Model**: Streamlined track naming and scoping
- **Efficient Session Management**: Optimized session establishment and maintenance
- **WebTransport & QUIC Support**: Full support for both WebTransport and raw QUIC connections
- **Publisher/Subscriber Pattern**: Clean separation of concerns for media producers and consumers
- **Multiplexed Streaming**: Efficient handling of multiple concurrent media tracks

## Implementation Status

The following table tracks our implementation progress against the MOQ Lite specification sections:

| Section                                      | Implemented        | Tested             | Notes                          |
| -------------------------------------------- | ------------------ | ------------------ | ------------------------------ |
| **2. Data Model**                            |                    |                    |                                |
| 2.1. Frame                                   | :white_check_mark: | :white_check_mark: | Core frame structure complete  |
| 2.3. Group                                   | :white_check_mark: | :white_check_mark: | Group handling implemented     |
| 2.4. Track                                   | :white_check_mark: | :white_check_mark: | Basic track management         |
| 2.4.1. Track Naming and Scopes               | :construction:     | :x:                | Partial implementation         |
| 2.4.2. Scope                                 | :construction:     | :x:                | Under development              |
| 2.4.3. Connection URL                        | :construction:     | :x:                | Planning phase                 |
| **3. Sessions**                              |                    |                    |                                |
| 3.1. Session establishment                   | :white_check_mark: | :white_check_mark: | Full session lifecycle         |
| 3.1.1. WebTransport                          | :white_check_mark: | :white_check_mark: | Browser compatibility          |
| 3.1.2. QUIC                                  | :white_check_mark: | :construction:     | Native QUIC support            |
| 3.2. Version and Extension Negotiation       | :white_check_mark: | :construction:     | Version handling complete      |
| 3.3. Session initialization                  | :white_check_mark: | :white_check_mark: | Setup handshake                |
| 3.4. Stream Cancellation                     | :construction:     | :x:                | Graceful termination           |
| 3.5. Termination                             | :white_check_mark: | :white_check_mark: | Clean shutdown                 |
| 3.6. Migration                               | :construction:     | :x:                | Future enhancement             |
| **4. Data Transmissions**                    |                    |                    |                                |
| 4.1 Track Priority Control                   | :construction:     | :x:                | QoS implementation             |
| 4.2 Group Order Control                      | :construction:     | :x:                | Ordering mechanisms            |
| 4.3 Cache                                    | :construction:     | :x:                | Caching strategies             |
| **5. Relays**                                |                    |                    |                                |
| 5.1. Subscriber Interactions                 | :white_check_mark: | :white_check_mark: | Subscribe/Unsubscribe          |
| 5.1.1. Graceful Publisher Relay Switchover   | :x:                | :x:                | Advanced relay feature         |
| 5.2. Publisher Interactions                  | :white_check_mark: | :white_check_mark: | Announce/Publish               |
| 5.2.1. Graceful Publisher Network Switchover | :x:                | :x:                | Network mobility               |
| 5.2.2. Graceful Publisher Relay Switchover   | :x:                | :x:                | Relay mobility                 |
| 5.3. Relay Object Handling                   | :construction:     | :x:                | Object forwarding              |
| **6. Control Streams**                       |                    |                    |                                |
| 6.1. Session Stream                          | :white_check_mark: | :white_check_mark: | Bidirectional control          |
| 6.2. Announce Stream                         | :white_check_mark: | :white_check_mark: | Track announcements            |
| 6.3. Subscribe Stream                        | :white_check_mark: | :white_check_mark: | Subscription management        |
| **7. Control Messages**                      |                    |                    |                                |
| 7.1. Parameters                              | :construction:     | :white_check_mark: | Parameter negotiation          |
| 7.1.1. Version Specific Parameters           | :white_check_mark: | :white_check_mark: | Version compatibility          |
| 7.2. SESSION_CLIENT                          | :white_check_mark: | :white_check_mark: | Client session setup           |
| 7.3. SESSION_SERVER                          | :white_check_mark: | :white_check_mark: | Server session setup           |
| 7.4. SESSION_UPDATE                          | :white_check_mark: | :white_check_mark: | Session parameter updates      |
| 7.4.1. Versions                              | :white_check_mark: | :white_check_mark: | Version negotiation            |
| 7.4.2. Setup Parameters                      | :white_check_mark: | :white_check_mark: | Configuration exchange         |
| 7.5. ANNOUNCE_PLEASE                         | :white_check_mark: | :white_check_mark: | Announcement requests          |
| 7.6. ANNOUNCE                                | :white_check_mark: | :white_check_mark: | Track announcements            |
| 7.7. SUBSCRIBE                               | :white_check_mark: | :white_check_mark: | Subscription requests          |
| 7.8. SUBSCRIBE_OK                            | :white_check_mark: | :white_check_mark: | Subscription responses         |
| 7.9. SUBSCRIBE_UPDATE                        | :white_check_mark: | :white_check_mark: | Subscription modifications     |
| **8. Data Streams**                          |                    |                    |                                |
| 8.1. Group Stream                            | :white_check_mark: | :white_check_mark: | Media data delivery            |
| **9. Data Messages**                         |                    |                    |                                |
| 9.1. GROUP                                   | :white_check_mark: | :white_check_mark: | Group message format           |
| 9.2. FRAME                                   | :white_check_mark: | :white_check_mark: | Frame message format           |
| **10. Security Considerations**              |                    |                    |                                |
| 10.1. Resource Exhaustion                    | :x:                | :x:                | DoS protection (planned)       |

## Architecture

### Core Components

- **Server**: MOQ Lite server implementation with WebTransport and QUIC support
- **Session**: Session management for client-server communication
- **Publisher**: Media track publishing capabilities
- **Subscriber**: Media track subscription and consumption
- **Track Management**: Efficient track routing and multiplexing
- **Stream Handling**: Bidirectional and unidirectional stream management

### Key Design Principles

1. **Simplicity**: MOQ Lite focuses on essential features for media streaming
2. **Performance**: Optimized for low-latency real-time applications
3. **Scalability**: Efficient handling of multiple concurrent sessions and tracks
4. **Compatibility**: Support for both WebTransport (browsers) and raw QUIC (native apps)

## Interoperability

### Current Status
- **Internal Testing**: Comprehensive unit and integration tests
- **Cross-Implementation**: Interoperability testing with other MOQ Lite implementations is planned
- **Browser Support**: WebTransport implementation tested with modern browsers
- **Native Applications**: Raw QUIC support for server-to-server and native client applications

### Tested Scenarios
- Session establishment and teardown
- Track announcement and subscription
- Group and frame data transmission
- Parameter negotiation and version compatibility

## Development Roadmap

### Immediate Priorities
- [ ] Complete SUBSCRIBE_GAP implementation
- [ ] Enhance stream cancellation mechanisms
- [ ] Improve resource exhaustion protection
- [ ] Comprehensive interoperability testing

### Medium-term Goals
- [ ] Track priority control implementation
- [ ] Group order control mechanisms
- [ ] Advanced caching strategies
- [ ] Relay switchover capabilities

### Future Enhancements
- [ ] Connection migration support
- [ ] Advanced QoS features
- [ ] Performance optimizations
- [ ] Extended security features

## Related Packages

- **[lomc](../lomc/)**: Low Overhead Media Container implementation (planned)
- **[catalog](../catalog/)**: MOQ Catalog for content discovery (planned)
- **[examples](../examples/)**: Sample applications demonstrating usage

## Contributing

When contributing to the MOQ Lite implementation, please ensure:

1. **Specification Compliance**: All changes should align with the MOQ Lite specification
2. **Test Coverage**: Include comprehensive tests for new features
3. **Documentation**: Update this status table when implementing new sections
4. **Performance**: Consider performance implications for real-time streaming

## References

- [MOQ Lite Specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)
- [QUIC Transport Protocol](https://tools.ietf.org/html/rfc9000)
- [WebTransport Protocol](https://tools.ietf.org/html/draft-ietf-webtrans-http3/)
