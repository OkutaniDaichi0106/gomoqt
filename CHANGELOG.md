# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Frame `Write()` method implementing `io.Writer` interface
- Frame `Body()` method for direct payload access
- Frame `Clone()` method for deep copying
- Comprehensive Frame tests (20+ test cases)
- GroupReader `Frames()` iterator with optional buffer parameter

### Changed

- Replaced Frame `Append()` with private `append()` method
- Frame `Bytes()` method renamed to `Body()`
- Updated all Frame test methods to use `Write()` instead of `Append()`
- Improved Frame encapsulation and API design
- Enhanced memory efficiency through optimized buffer reuse
- Expanded test coverage for GroupReader iterator pattern

### Fixed

- Fixed a bug where `Frame.encode` could write extra zero bytes beyond the actual payload, causing clients to receive a large number of empty frames (`frame_length=0`). Now only the header and payload are written, ensuring protocol correctness and efficiency.

