package moqt

import (
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

var DefaultGroupBytes = defaultGroupBytes

const defaultGroupBytes = 1024

var groupBufferPool = sync.Pool{
	New: func() interface{} {
		return &GroupRelayer{
			bytes:       make([]byte, 0, DefaultGroupBytes),
			frameRanges: make([]struct{ start, end int }, 0, 16),
		}
	},
}

var (
	ErrGroupClosed    = errors.New("group is closed")
	ErrSequenceRange  = errors.New("frame sequence is out of range")
	ErrOffsetNotFound = errors.New("frame offset not found")
)

type GroupRelayer struct {
	groupSequence GroupSequence
	bytes         []byte

	frameRanges []struct{ start, end int }

	closed   bool
	cond     *sync.Cond
	closeErr error
}

func NewGroupRelayer(gr GroupReader) *GroupRelayer {
	relayer := groupBufferPool.Get().(*GroupRelayer)

	// Set group sequence
	relayer.groupSequence = gr.GroupSequence()

	// Reset bytes
	relayer.bytes = relayer.bytes[:0]

	// Reset offsets
	relayer.frameRanges = relayer.frameRanges[:0]

	// Reset closed flag
	relayer.closed = false

	relayer.closeErr = nil

	// Reset condition
	relayer.cond = sync.NewCond(&sync.Mutex{})

	// Read frames from group reader
	go func() {
		for {
			frame, err := gr.ReadFrame()
			if err != nil {
				relayer.cond.L.Lock()
				relayer.closeErr = err
				relayer.closed = true
				relayer.cond.L.Unlock()
				relayer.cond.Broadcast()
				return
			}

			relayer.cond.L.Lock()

			// Append frame to bytes
			relayer.bytes = message.AppendBytes(relayer.bytes, frame)

			// Append offset to offsets
			relayer.frameRanges = append(relayer.frameRanges, struct {
				start int
				end   int
			}{
				start: len(relayer.bytes) - len(frame),
				end:   len(relayer.bytes)},
			)

			relayer.cond.L.Unlock()
			relayer.cond.Broadcast()
		}
	}()

	return relayer
}

func (g *GroupRelayer) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g *GroupRelayer) Closed() bool {
	return g.closed
}

func (g *GroupRelayer) Relay(gw GroupWriter) error {
	if gw == nil {
		return errors.New("group writer is nil")
	}

	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	grstr, ok := gw.(*sendGroupStream)
	var (
		readFrames  int
		writeOffset int
	)

	for {
		// Check for new data or closure
		for readFrames >= len(g.frameRanges) {
			if g.closed {
				return g.closeErr // Returns nil if closed normally
			}
			g.cond.Wait()
		}

		// Handle stream writing
		if ok {
			currentSize := len(g.bytes)
			if currentSize > writeOffset {
				n, err := grstr.internalStream.Stream.Write(g.bytes[writeOffset:currentSize])
				if err != nil {
					return err
				}
				writeOffset += n
			}
			readFrames = len(g.frameRanges)
			continue
		}

		// Handle frame-by-frame writing
		frameRange := g.frameRanges[readFrames]
		if err := gw.WriteFrame(g.bytes[frameRange.start:frameRange.end]); err != nil {
			return err
		}
		readFrames++
	}
}

func (g *GroupRelayer) Release() {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	// Close group
	g.closed = true

	// Broadcast to all
	g.cond.Broadcast()

	// Reset for reuse
	g.bytes = g.bytes[:0]

	// Reset for reuse
	g.frameRanges = g.frameRanges[:0]

	// Reset for reuse
	g.closeErr = nil

	groupBufferPool.Put(g)
}
