# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Korean translation of README (`README.ko.md`)

### Changed

- Renamed `Session.SessionUpdated()` to `Session.Updated()`
- docs: Updated all documentation to align with current implementation

## [v0.6.0] - 2025-12-05

### Added

- `bitrate` package: Bitrate monitoring functionality with `ShiftDetector` interface and `EWMAShiftDetector` implementation for detecting bitrate shifts using Exponential Weighted Moving Average

### Fixed

- `AnnouncementWriter`: Avoid deadlock by calling end functions asynchronously

### Changed

- Modernize test code: Replace traditional for loops with range loops
- Stabilize flaky `TestAnnouncement_AfterFunc_ConcurrentRegistrationAndEnd` test
- Replace `WritableStream` and `ReceiveStream` with `Buffer` in message tests

## [v0.5.1] - 2025-12-04

### Changed

- docs: Update README files to add table of contents and improve clarity
- docs: Update README for correct package references

## [v0.5.0] - 2025-12-04

### Changed

- SUBSCRIBE message: removed `Min Group Sequence` and `Max Group Sequence` fields for simplified group sequence handling
- TrackWriter `OpenGroup()`: now internally manages group sequence increments without requiring it as an argument

## [v0.4.0] - 2025-12-04

### Added

- Frame `Write()` method implementing `io.Writer` interface
- Frame `Body()` method for direct payload access
- Frame `Clone()` method for deep copying
- Comprehensive Frame tests (20+ test cases)
- GroupReader `Frames()` iterator with optional buffer parameter
- Interop Mage targets: `mage interop:server` and `mage interop:client` with `-go`/`-ts` flags for running interop tests and demos
- `TrackReader` and `TrackWriter` classes with associated unit tests
- Additional unit tests for `SessionStream` and Stream types covering new scenarios

### Changed

- Changed default protocol version from Develop (0xffffff00) to LiteDraft01 (0xff0dad01)
- Changed message length encoding from QUIC variable-length integer to big-endian u16
- Replaced Frame `Append()` with private `append()` method
- Frame `Bytes()` method renamed to `Body()`
- Updated all Frame test methods to use `Write()` instead of `Append()`
- Improved Frame encapsulation and API design
- Enhanced memory efficiency through optimized buffer reuse
- Expanded test coverage for GroupReader iterator pattern
- Migrated moq-web from Node.js to Deno runtime
- Moved hang-web directory to moqrtc-js repository
- Refactor WebTransport stream handling: introduced `StreamID` type and `WebTransportSession`; improved error handling and logging
- Refactor interop server and client: improved address/config handling, context management, and added secure `mkcert` wrapper
- Refactor subscription stream and track handling: use `SubscribeErrorCode`, numeric group sequence types, graceful closure, and enhanced logging
- Replace `session.Terminate` with `session.CloseWithError` for consistent session closure behavior
- Refactor announcement handling in `TrackMux` and related components
- Update dependencies and improve type-safety in translate and interop client scripts

### Fixed

- Fixed a bug where `Frame.encode` could write extra zero bytes beyond the actual payload, causing clients to receive a large number of empty frames (`frame_length=0`). Now only the header and payload are written, ensuring protocol correctness and efficiency.
