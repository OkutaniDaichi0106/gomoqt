package moqtransport

import (
	"container/heap"
	"errors"
	"io"
	"sync"
)

type DataStream interface {
	//Close()
	StreamReader
	StreamWriter
	heap.Interface
}

type StreamReader interface {
	io.Reader
}

type StreamWriter interface {
	//Write(objectID, []byte) error
}

func newDataStream(header StreamHeader) DataStream {
	switch h := header.(type) {
	case *StreamHeaderTrack:
		stream := &TrackStream{
			header: *h,
			queue:  make([]GroupChunk, 0, 1<<3),
		}
		return stream
	case *StreamHeaderPeep:
		stream := &PeepStream{
			header: *h,
			queue:  make([]ObjectChunk, 0, 1<<3),
		}
		return stream
	default:
		return nil
	}

}

type TrackStream struct {
	mu     sync.Mutex
	header StreamHeaderTrack
	queue  []GroupChunk
}

// func (TrackStream) Close() {} //TODO

func (stream *TrackStream) Read(buf []byte) (int, error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Check if the next objext exists
	if len(stream.queue) == 0 {
		return 0, io.EOF
	}

	// Get Data as Group Chunk
	chunk := stream.Pop().(GroupChunk)

	if len(buf) < len(chunk.Payload) {
		return 0, errors.New("too small buffer")
	}

	// Set to the buffer
	n := copy(buf, chunk.Payload)

	// Return the size of the data
	return n, nil
}

func (stream *TrackStream) Len() int {
	return len(stream.queue)
}

func (stream *TrackStream) Less(i, j int) bool {
	if stream.queue[i].groupID == stream.queue[j].groupID {
		return stream.queue[i].objectID < stream.queue[j].objectID
	}

	return stream.queue[i].groupID < stream.queue[j].groupID
}

func (stream *TrackStream) Swap(i, j int) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Swap
	stream.queue[i], stream.queue[j] = stream.queue[j], stream.queue[i]
}

func (stream *TrackStream) Push(x any) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	chunk := x.(GroupChunk)
	stream.queue = append(stream.queue, chunk)
}

func (stream *TrackStream) Pop() any {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	old := stream.queue
	n := len(old)
	chunk := stream.queue[n-1]
	stream.queue = old[:n-1]
	return chunk
}

// func (writer *TrackStream) Write(data []byte) error {
// 	writer.queue[id] = data
// 	return nil
// }

type PeepStream struct {
	mu     sync.Mutex
	header StreamHeaderPeep
	queue  []ObjectChunk
}

//func (PeepStream) Close() {} //TODO

func (stream *PeepStream) Read(buf []byte) (int, error) {
	// Check if the next objext exists
	if len(stream.queue) == 0 {
		return 0, errors.New("no data")
	}

	// Get Data as Group Chunk
	chunk, ok := stream.Pop().(ObjectChunk)
	if !ok {
		return 0, errors.New("obtained data is not object chunk")
	}

	// Set to the buffer
	buf = chunk.Payload

	// Clean the read data
	stream.queue = stream.queue[1:]

	// Return the size of the data
	return len(chunk.Payload), nil
}

func (stream *PeepStream) Len() int {
	return len(stream.queue)
}

func (stream *PeepStream) Less(i, j int) bool {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	return stream.queue[i].objectID < stream.queue[j].objectID
}

func (stream *PeepStream) Swap(i, j int) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	stream.queue[i], stream.queue[j] = stream.queue[j], stream.queue[i]
}

func (stream *PeepStream) Push(x any) {
	chunk := x.(ObjectChunk)
	stream.queue = append(stream.queue, chunk)
}

func (stream *PeepStream) Pop() any {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	old := stream.queue
	n := len(old)
	chunk := stream.queue[n-1]
	stream.queue = old[:n-1]
	return chunk
}

// type TrackReader interface {
// 	io.Reader
// }

// type trackReader struct {
// 	subscribeID
// 	index []GroupStream
// }

// func newTrackReader(id subscribeID) TrackReader {
// 	return trackReader{
// 		subscribeID: id,
// 		index:       []GroupStream{newDataStream()},
// 	}
// }

// func (trackReader) Read(buf []byte) (int, error) {

// }
