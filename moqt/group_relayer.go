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

	closed bool
	cond   *sync.Cond
	err    error
}

func NewGroupBuffer(gr GroupReader) *GroupRelayer {
	gb := groupBufferPool.Get().(*GroupRelayer)

	// Set group sequence
	gb.groupSequence = gr.GroupSequence()

	// Reset bytes
	gb.bytes = gb.bytes[:0]

	// Reset offsets
	gb.frameRanges = gb.frameRanges[:0]

	// Reset closed flag
	gb.closed = false

	gb.err = nil

	// Reset condition
	gb.cond = sync.NewCond(&sync.Mutex{})

	go func() {
		for {
			frame, err := gr.ReadFrame()
			if err != nil {
				gb.cond.L.Lock()
				gb.err = err
				gb.closed = true
				gb.cond.L.Unlock()
				gb.cond.Broadcast()
				return
			}

			gb.cond.L.Lock()

			// Append frame to bytes
			gb.bytes = message.AppendBytes(gb.bytes, frame)

			// Append offset to offsets
			gb.frameRanges = append(gb.frameRanges, struct {
				start int
				end   int
			}{
				start: len(gb.bytes) - len(frame),
				end:   len(gb.bytes)},
			)

			gb.cond.L.Unlock()
			gb.cond.Broadcast()
		}
	}()

	return gb
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
				return g.err // Returns nil if closed normally
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
	g.err = nil

	groupBufferPool.Put(g)
}
