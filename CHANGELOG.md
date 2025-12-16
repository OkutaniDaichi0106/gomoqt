# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [v0.7.0] - 2025-12-16

### Breaking Changes

- **Message Length Encoding Changed to Varint**: Changed message length encoding from uint16 big-endian to QUIC variable-length integer (varint)
  - Message length is now encoded using standard QUIC varint format (1, 2, 4, or 8 bytes depending on value)
  - This change aligns the implementation with the QUIC specification and improves efficiency for small messages
  - Messages up to 63 bytes now use only 1 byte for length (previously always 2 bytes)
  - Maximum message size increased from 65,535 bytes to 2^62-1 bytes
  - **Impact**: This is a protocol-breaking change. Old clients and servers cannot communicate with new ones
  - **Migration**: All endpoints must be updated simultaneously to maintain compatibility

- **EWMA Bitrate Notification Removed**: Removed experimental EWMA-based bitrate notification feature (v0.6.0)
  - Removed `moqt/bitrate/` package (ewma.go, ewma_test.go, shift_detector.go)
  - Removed `NewShiftDetector` field from `Config`
  - Removed `ConnectionStats()` method from `quic.Connection` interface
  - **Reason**: Feature depended on non-public APIs from forked quic-go, causing instability and preventing library users from using the package due to Go module replace directive limitations
  - **Migration**: This feature has been preserved in the `feature/ewma-bitrate-notification` branch for reference
  - `Session.goAway()` is now a no-op (graceful shutdown is handled by QUIC connection close)

- **Removed Go Module Replace Directives**: Removed replace directives for forked dependencies
  - No longer using `github.com/okdaichi/quic-go` or `github.com/okdaichi/webtransport-go`
  - Now using upstream `github.com/quic-go/quic-go` v0.57.1 and `github.com/quic-go/webtransport-go`
  - **Impact**: Library can now be used as a dependency without type compatibility issues
  - All tests passing with upstream dependencies

### Performance

- **TrackMux Advanced Optimizations**: Further improved performance with lock contention reduction and memory efficiency
  - **Lock Optimization**: Reduced lock hold time in `findTrackHandler` by performing all checks within single RLock
  - **Memory Allocation**: Moved handler struct allocation outside critical section in `registerHandler`
  - **Code Deduplication**: Refactored `serveTrack` to reuse optimized `findTrackHandler`, eliminating duplicate lock acquisition
  - **Read-Write Lock Pattern**: Implemented double-check locking in `getChild` to minimize write lock contention
  - **Worker Pool Enhancement**: Optimized `Announcement.end()` with inline execution for small handler counts and efficient work distribution
  - **Results**: Handler lookup improved to 21-25ns (48-51% from baseline, 12-20% from first optimization)

- **Initial TrackMux Optimizations**: Improved performance of track handler lookups and announcements
  - Reduced lock contention in `findTrackHandler` by simplifying map lookups
  - Pre-allocated maps with initial capacity to reduce allocations during runtime
  - Removed unnecessary defer statements for faster lock/unlock operations
  - Pre-allocated slices in `Announce` function to reduce dynamic allocations
  - **Results**: Handler lookup improved by 42-67% (41ns → 24-31ns), ServeTrack improved by 23% (243ns → 187ns), GC overhead reduced from 55% to 25%

### Fixed

- **Benchmark Test Mocks**: Fixed `BenchmarkTrackMux_ServeAnnouncements` by adding required mock expectations for `Context()` and `Write()` methods

## [v0.6.2] - 2025-12-10

### Changed

- **API Encapsulation**: Changed `sendSubscribeStream.UpdateSubscribe()` from public to private (`updateSubscribe()`) to improve API boundaries
  - `TrackReader.Update()` remains the only public API for updating subscription configurations
  - Prevents unintended direct access to internal implementation methods while maintaining embedding benefits

## [v0.6.1] - 2025-12-09

### Added

- Chinese (Simplified) translation of README (`README.zh-cn.md`)
- Korean translation of README (`README.ko.md`)
- Chinese translation of README (`README.zh.md`)
- Russian translation of README (`README.ru.md`)
- German translation of README (`README.de.md`)
- Japanese translation of README (`README.ja.md`)
- Language selection links in all README files for improved accessibility
- Detailed README files for interop, examples, and moqt package

### Changed

- **Repository ownership**: Changed GitHub username from `OkutaniDaichi0106` to `okdaichi`
- Renamed `Session.SessionUpdated()` to `Session.Updated()`
- Renamed `Session.Terminate()` to `Session.CloseWithError()` for consistency
- Updated all documentation to align with current implementation and reflect correct GitHub username
- Improved README formatting and features section clarity across all languages
- Updated module replace directives to use forked quic-go and webtransport-go commits

## [v0.6.0] - 2025-12-05

### Added

- `bitrate` package: Bitrate monitoring functionality with `ShiftDetector` interface and `EWMAShiftDetector` implementation for detecting bitrate shifts using Exponential Weighted Moving Average

### Fixed

- `AnnouncementWriter`: Avoid deadlock by calling end functions asynchronously

### Changed

- Modernize test code: Replace traditional for loops with range loops

## [v0.5.0] - 2025-11-27

### Changed

- Update Broadcast example: Switch from LiveKit to UDP as media source
- `Mux`: Return `ErrNoSubscribers` on failure to find subscribers instead of GOAWAY

## [v0.4.3] - 2025-11-26

### Changed

- Improve error handling: Distinguish temporary and permanent errors

## [v0.4.2] - 2025-11-25

### Fixed

- Fix duplicate panic in announcement handling

## [v0.4.1] - 2025-11-24

### Fixed

- `TrackWriter.Close()`: Handle stream closure errors
- `GroupWriter`: Add nil check for frame field to prevent panic

## [v0.4.0] - 2025-11-24

### Added

- New track writer implementation (`TrackWriter`, `GroupWriter`, `FrameWriter`)
- Concurrent frame writing support via `TrackWriter.Spawn()`
- `TrackWriter.Write()` method for direct frame writing
- Generic parameter type support for `TrackConfig`

### Changed

- Replace `TrackPublisher` with new `TrackWriter` API
- Simplify parallel group writing with direct track writer operations
- `SendSubscribeStream` now returns `*TrackWriter` instead of `TrackPublisher`

### Removed

- Old `TrackPublisher` API

## [v0.3.0] - 2025-11-21

### Added

- Native QUIC support: Direct QUIC connection examples in `examples/native_quic`
- `quic` package: Wrapper for QUIC functionality used by core library and examples
- Russian translation of README (`README.ru.md`)
- German translation of README (`README.de.md`)
- Japanese translation of README (`README.ja.md`)

### Changed

- Reorganize dependencies: Separate QUIC and WebTransport dependencies for flexible usage
- Update examples to demonstrate both WebTransport and native QUIC usage

## [v0.2.0] - 2025-11-15

### Added

- WebTransport support via `webtransport` package
- Interoperability testing suite in `cmd/interop`
- TypeScript client implementation in `moq-web`

### Changed

- Improve session management and error handling
- Update documentation with WebTransport examples

## [v0.1.0] - 2025-11-01

### Added

- Initial implementation of MOQ Lite protocol
- Core `moqt` package with session, track, group, and frame handling
- Basic examples: broadcast, echo, relay
- Mage build system integration
- Comprehensive test coverage
- MIT License

[Unreleased]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.6.0...HEAD
[v0.6.0]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.4.3...v0.5.0
[v0.4.3]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/OkutaniDaichi0106/gomoqt/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/OkutaniDaichi0106/gomoqt/releases/tag/v0.1.0
