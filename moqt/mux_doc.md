# TrackMux: Track Routing and Announcement System

## Overview

The `mux.go` file implements a multiplexer system for handling track-based routing and announcements in the moqt package. It provides a path-based routing mechanism similar to HTTP multiplexers but specialized for tracks and their announcement handling.

## Key Components

### TrackMux

The central component that manages routing of track requests to appropriate handlers and handles announcements for new tracks.

### Path Routing

Uses a tree-based routing mechanism that matches track paths to registered handlers, supporting efficient lookup and dispatch.

### Announcement System

Provides a subscription model for track announcements, allowing clients to receive notifications when new tracks matching specific patterns are registered.

---

## 1. File Specifications

### 1.1 Purpose and Scope

This file solves the problem of routing track requests to appropriate handlers based on path patterns, while also providing a system for announcing new tracks to interested subscribers. It centralizes track routing logic and implements an efficient tree-based path matching algorithm.

### 1.2 Input/Output Expectations

- **Input**: Track paths (string-based identifiers), track handlers, and subscription configurations
- **Output**: Proper routing of track requests to handlers, management of track announcements, and retrieval of track information

### 1.3 Technical Requirements

- Thread-safe operations via read-write mutexes
- Efficient path matching using a tree-based routing structure
- Support for wildcards in announcement subscriptions

---

## 2. Architecture Guidelines

### 2.1 Design Patterns

- **Multiplexer Pattern**: Similar to HTTP multiplexers but specialized for tracks
- **Observer Pattern**: For the announcement system where subscribers receive notifications of new tracks
- **Tree-based Routing**: For efficient path matching and dispatching

### 2.2 Code Structure

- **Package**: `moqt`
- **Primary Types**:
  - `TrackMux`: Main multiplexer implementation
  - `routingNode`: Node in the routing tree
  - `announcingNode`: Node in the announcement tree
  - `path` and `pattern`: Path representation and matching

---

## 3. Function Specifications

### Function: Handle

Description: Registers a handler for a specific track path

Signature:
```go
func (mux *TrackMux) Handle(path TrackPath, handler TrackHandler)
```

Parameters:
- path: The track path to register the handler for
- handler: The handler to invoke for the given path

Usage Example:
```go
mux := NewTrackMux()
mux.Handle("/tracks/audio", myAudioHandler)
```

### Function: ServeTrack

Description: Serves a track request by finding and invoking the appropriate handler

Signature:
```go
func (mux *TrackMux) ServeTrack(w TrackWriter, config SubscribeConfig)
```

Parameters:
- w: The track writer to write the track data to
- config: Configuration for the subscription

Usage Example:
```go
mux.ServeTrack(trackWriter, subscribeConfig)
```

### Function: ServeAnnouncements

Description: Registers an announcement writer to receive announcements for new tracks

Signature:
```go
func (mux *TrackMux) ServeAnnouncements(w AnnouncementWriter)
```

Parameters:
- w: The announcement writer to send announcements to

Usage Example:
```go
mux.ServeAnnouncements(announcementWriter)
```

### Function: GetInfo

Description: Retrieves information about a track at the specified path

Signature:
```go
func (mux *TrackMux) GetInfo(path TrackPath) (Info, error)
```

Parameters:
- path: The path of the track to get information for

Return Values:
- Info: Information about the track
- error: Error if the track doesn't exist or another issue occurs

Usage Example:
```go
info, err := mux.GetInfo("/tracks/audio")
if err != nil {
    // Handle error
}
// Use info
```

Edge Cases to Handle:
- Non-existent track paths
- Concurrent access to the mux

---

## 4. Testing Requirements

### 4.1 Test Coverage Goals

- Minimum 80% test coverage
- Focus on path matching logic, especially edge cases
- Verify thread safety with concurrent access tests

### 4.2 Test Cases

- Registration of handlers at different paths
- Routing of requests to the correct handlers
- Handling of non-existent paths
- Announcement subscription and delivery
- Wildcard matching in announcement subscriptions
- Concurrent registration and access

### 4.3 Test Approach

- Unit tests for individual components
- Integration tests for complete request flow
- Benchmark tests for path matching performance
- Mock TrackWriter and AnnouncementWriter for testing

---

## 5. Usage Examples

### Basic Usage

```go
// Create a new mux
mux := moqt.NewTrackMux()

// Register handlers
mux.Handle("/tracks/audio", audioHandler)
mux.Handle("/tracks/video", videoHandler)

// Serve tracks
mux.ServeTrack(trackWriter, subscribeConfig)

// Get track info
info, err := mux.GetInfo("/tracks/audio")
```

### Announcement Subscription

```go
// Subscribe to announcements for all audio tracks
announceWriter := NewMyAnnouncementWriter("/tracks/audio/**")
mux.ServeAnnouncements(announceWriter)

// Later when new audio tracks are registered, announceWriter will be notified
```

### Using Default Mux

```go
// Register handler with the default mux
moqt.Handle("/tracks/audio", audioHandler)

// Serve tracks using the default mux
moqt.ServeTrack(trackWriter, subscribeConfig)
```

---

## 6. Implementation Details

### 6.1 Path Matching

TrackMux uses a tree-based path matching system where each segment of a path is a node in the tree. This allows for efficient routing of requests to the appropriate handler.

### 6.2 Announcement System

The announcement system uses a similar tree-based approach but with support for wildcards:
- `*` matches a single path segment
- `**` matches zero or more path segments

### 6.3 Thread Safety

All operations on TrackMux are thread-safe, protected by a read-write mutex that allows multiple concurrent reads but exclusive writes.

---

## 7. Error Handling

- Returns `ErrTrackDoesNotExist` when a track is not found
- Uses `NotFoundHandler` for handling requests to non-existent paths
- Logs warnings when a handler is overwritten
