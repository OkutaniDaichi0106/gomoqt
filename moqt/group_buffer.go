package moqt

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

func RelayGroup(gr GroupReader, gw GroupWriter) error {
	if gr == nil || gw == nil {
		return errors.New("group reader or writer is nil")
	}

	if gr.GroupSequence() != gw.GroupSequence() {
		return errors.New("group reader and writer have different group sequences")
	}

	defer gw.Close()

	dr, readable := gr.(directBytesReader)
	dw, writable := gw.(directBytesWriter)

	if readable && writable {
		// Relay bytes
		r := dr.newBytesReader()
		w := dw.newBytesWriter()

		buf := make([]byte, DefaultGroupBufferSize)
		for {
			n, err := r.Read(&buf)
			if n > 0 {
				data := buf[:n]
				_, err = w.Write(&data)
				if err != nil {
					return err
				}
			}
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}

		}
	} else {
		// Relay frames
		for {
			frame, err := gr.ReadFrame()
			if len(frame) > 0 {
				err = gw.WriteFrame(frame)
				if err != nil {
					return err
				}
			}

			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}
}

var _ GroupReader = (*GroupBuffer)(nil)
var _ GroupWriter = (*GroupBuffer)(nil)

var _ directBytesReader = (*GroupBuffer)(nil)
var _ directBytesWriter = (*GroupBuffer)(nil)

type GroupBuffer struct {
	groupSequence GroupSequence

	// bytes
	bytes []byte

	// frameInfo
	frameInfo []struct {
		start, end int
	}

	readFrames int

	closed bool

	closedErr error

	cond *sync.Cond

	cancelErrCode *GroupErrorCode

	// Updated fields for deadline support: separate timers for read and write.
	readDeadline      *time.Time  // read deadline time
	readDeadlineTimer *time.Timer // timer for read deadline

	writeDeadline      *time.Time  // write deadline time
	writeDeadlineTimer *time.Timer // timer for write deadline
}

// NewGroupBuffer creates a new GroupBuffer with the specified group sequence and initial capacity.

func NewGroupBuffer(seq GroupSequence, size int) *GroupBuffer {
	if size <= 0 {
		size = DefaultGroupBufferSize
	}
	return &GroupBuffer{
		groupSequence: seq,
		cond:          sync.NewCond(&sync.Mutex{}),
		bytes:         make([]byte, 0, size),
		frameInfo:     make([]struct{ start, end int }, 0),
	}
}

func (g *GroupBuffer) GroupSequence() GroupSequence {
	return g.groupSequence
}

// GroupReader: ReadFrame returns all available data as one frame.
func (g *GroupBuffer) ReadFrame() ([]byte, error) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	for g.readFrames >= len(g.frameInfo) {
		if g.closed {
			return nil, io.EOF
		}
		// Check if the read deadline has been exceeded.
		if g.readDeadline != nil && time.Now().After(*g.readDeadline) {
			return nil, errors.New("read deadline exceeded")
		}
		// Check if the read has been canceled.
		if g.cancelErrCode != nil {
			return nil, errors.New("read canceled")
		}
		g.cond.Wait()
	}

	frameInfo := g.frameInfo[g.readFrames]

	frame := g.bytes[frameInfo.start:frameInfo.end]
	frameCopy := make([]byte, len(frame))
	copy(frameCopy, frame)

	g.readFrames++

	return frameCopy, nil
}

func (g *GroupBuffer) CancelRead(err GroupErrorCode) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	g.cancelErrCode = &err
	g.cond.Broadcast()
}

// SetReadDeadline sets the deadline for read operations.
// When the deadline is exceeded, waiting read operations will return an error.
func (g *GroupBuffer) SetReadDeadline(t time.Time) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		return g.closedErr
	}

	g.readDeadline = &t

	// Stop any existing read deadline timer
	if g.readDeadlineTimer != nil {
		g.readDeadlineTimer.Stop()
	}

	d := time.Until(t)
	if d <= 0 {
		// Deadline already passed, wake up any waiting goroutines immediately
		g.cond.Broadcast()
		return nil
	}

	// Start a timer that will broadcast on the condition variable when the deadline is reached
	g.readDeadlineTimer = time.AfterFunc(d, func() {
		g.cond.L.Lock()
		defer g.cond.L.Unlock()
		g.cond.Broadcast()
	})

	return nil
}

// GroupWriter: WriteFrame appends a frame to the internal buffer.

func (g *GroupBuffer) WriteFrame(frame []byte) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		return g.closedErr
	}
	// If read was canceled, stop accepting writes.
	if g.cancelErrCode != nil {
		return errors.New("write canceled")
	}

	// Create a copy of the frame before appending
	frameCopy := make([]byte, len(frame))
	copy(frameCopy, frame)

	g.bytes = message.AppendBytes(g.bytes, frameCopy)

	g.frameInfo = append(g.frameInfo, struct{ start, end int }{
		start: len(g.bytes) - len(frame),
		end:   len(g.bytes),
	})

	g.cond.Broadcast()

	return nil
}

func (g *GroupBuffer) CancelWrite(code GroupErrorCode) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	g.cancelErrCode = &code
	g.cond.Broadcast()
}

func (g *GroupBuffer) SetWriteDeadline(t time.Time) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		return g.closedErr
	}

	g.writeDeadline = &t

	// Stop any existing write deadline timer
	if g.writeDeadlineTimer != nil {
		g.writeDeadlineTimer.Stop()
	}

	d := time.Until(t)
	if d <= 0 {
		// Deadline already passed, wake up any waiting goroutines immediately
		g.cond.Broadcast()
		return nil
	}

	// Start a timer that will broadcast on the condition variable when the deadline is reached
	g.writeDeadlineTimer = time.AfterFunc(d, func() {
		g.cond.L.Lock()
		defer g.cond.L.Unlock()
		g.cond.Broadcast()
	})

	return nil
}

// Close marks the buffer as closed and wakes up all waiting goroutines
func (g *GroupBuffer) Close() error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		return g.closedErr
	}

	g.closed = true
	g.closedErr = nil

	// Stop any deadline timers if they exist to avoid unnecessary callbacks.
	if g.readDeadlineTimer != nil {
		g.readDeadlineTimer.Stop()
	}
	if g.writeDeadlineTimer != nil {
		g.writeDeadlineTimer.Stop()
	}

	g.cond.Broadcast() // Wake up all waiting goroutines
	return nil
}

func (g *GroupBuffer) CloseWithError(err error) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if err == nil {
		return g.Close()
	}

	if g.closed {
		return g.closedErr
	}

	g.closedErr = err
	g.closed = true

	// Stop any deadline timers.
	if g.readDeadlineTimer != nil {
		g.readDeadlineTimer.Stop()
	}
	if g.writeDeadlineTimer != nil {
		g.writeDeadlineTimer.Stop()
	}

	g.cond.Broadcast()
	return nil
}

// Reset resets the buffer with a new group sequence.
func (g *GroupBuffer) Reset(seq GroupSequence) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if len(g.bytes) > 0 || len(g.frameInfo) > 0 {
		g.Drop()
	}

	g.groupSequence = seq
	g.readFrames = 0
	g.closed = false
	g.cond = sync.NewCond(&sync.Mutex{})
}

func (g *GroupBuffer) Drop() {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	g.Close()

	g.bytes = g.bytes[:0]
	g.frameInfo = g.frameInfo[:0]
	g.readFrames = 0
}

// internal method
func (g *GroupBuffer) newBytesReader() reader {
	return &groupBufferBytesReader{g, 0}
}

var _ reader = (*groupBufferBytesReader)(nil)

type groupBufferBytesReader struct {
	buffer *GroupBuffer
	offset int
}

func (b *groupBufferBytesReader) Read(p *[]byte) (int, error) {
	b.buffer.cond.L.Lock()
	defer b.buffer.cond.L.Unlock()

	if b.offset < 0 {
		return 0, errors.New("negative offset")
	}

	// Wait until we have data available or the read deadline is exceeded.
	for b.offset >= len(b.buffer.bytes) {
		if b.buffer.closed {
			return 0, io.EOF
		}
		// Check if the read has been canceled.
		if b.buffer.cancelErrCode != nil {
			return 0, errors.New("read canceled")
		}
		if b.buffer.readDeadline != nil && time.Now().After(*b.buffer.readDeadline) {
			return 0, errors.New("read deadline exceeded")
		}

		b.buffer.cond.Wait()
	}

	*p = b.buffer.bytes[b.offset:]
	b.offset += len(*p)

	return len(*p), nil
}

func (g *GroupBuffer) newBytesWriter() writer {
	return &groupBufferBytesWriter{g, 0}
}

var _ writer = (*groupBufferBytesWriter)(nil)

type groupBufferBytesWriter struct {
	buffer *GroupBuffer
	offset int
}

func (b *groupBufferBytesWriter) Write(p *[]byte) (int, error) {
	b.buffer.cond.L.Lock()
	defer b.buffer.cond.L.Unlock()

	if b.buffer.closed {
		return 0, ErrGroupClosed
	}
	// If read was canceled, do not allow further writes.
	if b.buffer.cancelErrCode != nil {
		return 0, errors.New("write canceled")
	}

	b.buffer.bytes = append(b.buffer.bytes, *p...)

	if len(b.buffer.frameInfo) > 0 {
		b.offset = b.buffer.frameInfo[len(b.buffer.frameInfo)-1].end
	}

	for b.offset < len(b.buffer.bytes) {
		r := bytes.NewReader(b.buffer.bytes[b.offset:])

		frameSize, err := quicvarint.Read(r)
		if err != nil {
			return 0, err
		}

		b.buffer.frameInfo = append(b.buffer.frameInfo, struct{ start, end int }{
			start: b.offset + quicvarint.Len(frameSize),
			end:   b.offset + quicvarint.Len(frameSize) + int(frameSize),
		})

		b.offset = b.buffer.frameInfo[len(b.buffer.frameInfo)-1].end
	}

	b.buffer.cond.Broadcast()

	return len(*p), nil
}

var (
	ErrGroupClosed = errors.New("group is closed")

	ErrSequenceRange = errors.New("frame sequence is out of range")
	ErrOutOfCache    = errors.New("group buffer is out of cache")
	ErrBufferFull    = errors.New("buffer is full")
)

var DefaultGroupBufferSize = defaultBufferSize

const defaultBufferSize = 1024 * 1024 // 1MB
