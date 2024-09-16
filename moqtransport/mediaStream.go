package moqtransport

import (
	"errors"
	"go-moq/moqtransport/moqtmessage"
	"io"
	"sync"
)

type ObjectStream interface {
	io.Reader
	Header() moqtmessage.StreamHeader
	// CancelReader()
	// SetReadDeadline(time.Time) error
}

type trackStream struct {
	mu     sync.RWMutex
	header moqtmessage.StreamHeaderTrack
	chunks []moqtmessage.GroupChunk
	closed bool
}

func (stream *trackStream) Close() {
	stream.closed = true
}

func (stream *trackStream) Read(buf []byte) (int, error) {
	if stream.closed {
		return 0, io.EOF
	}
	stream.mu.RLock()
	defer stream.mu.RUnlock()

	// Check if the next objext exists
	if len(stream.chunks) == 0 {
		return 0, io.EOF
	}

	// Get Data as Group Chunk
	chunk := stream.chunks[0]

	if len(buf) < len(chunk.Payload) {
		return 0, errors.New("too small buffer")
	}

	// Set to the buffer
	n := copy(buf, chunk.Payload)

	//Remove the chunk from the queue
	stream.chunks = stream.chunks[1:]

	// Return the size of the data
	return n, nil
}

func (stream *trackStream) Header() moqtmessage.StreamHeader {
	return &stream.header
}

func (stream *trackStream) write(chunk moqtmessage.GroupChunk) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Queue the chunk
	stream.chunks = append(stream.chunks, chunk)
}

// func (writer *TrackStream) Write(data []byte) error {
// 	writer.queue[id] = data
// 	return nil
// }

type peepStream struct {
	mu     sync.RWMutex
	header moqtmessage.StreamHeaderPeep
	chunks []moqtmessage.ObjectChunk
	closed bool
}

func (stream *peepStream) Close() {
	stream.closed = true
}

func (stream *peepStream) Read(buf []byte) (int, error) {
	if stream.closed {
		return 0, io.EOF
	}
	stream.mu.RLock()
	defer stream.mu.RUnlock()

	// Check if the next objext exists
	if len(stream.chunks) == 0 {
		return 0, io.EOF
	}

	// Get the first chunk
	chunk := stream.chunks[0]

	// Check if the buffer has enough size
	if len(buf) < len(chunk.Payload) {
		return 0, errors.New("too small buffer")
	}

	// Set to the buffer
	n := copy(buf, chunk.Payload)

	// Remove the chunk from the queue
	stream.chunks = stream.chunks[1:]

	// Return the size of the data
	return n, nil
}

func (stream *peepStream) Header() moqtmessage.StreamHeader {
	return &stream.header
}

func (stream *peepStream) write(chunk moqtmessage.ObjectChunk) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Queue the chunk
	stream.chunks = append(stream.chunks, chunk)
}
