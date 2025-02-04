package moqt

import (
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

func Relay(gr GroupReader, gw GroupWriter) error {
	if gr == nil || gw == nil {
		return errors.New("group reader or writer is nil")
	}

	if gr.GroupSequence() != gw.GroupSequence() {
		return errors.New("group reader and writer have different group sequences")
	}

	defer gw.Close()

	if r, ok := gr.(*receiveGroupStream); ok {

		// Relay bytes from receive stream
		if w, ok := gw.(*sendGroupStream); ok {
			// Relay bytes to send stream
			_, err := io.Copy(w.internalStream.SendStream, r.internalStream.ReceiveStream)
			if err != nil {
				return err
			}
			return nil
		} else if gb, ok := gw.(*GroupBuffer); ok {
			// Relay bytes to buffer
			buf := make([]byte, 1024)
			for {
				n, err := r.internalStream.ReceiveStream.Read(buf)
				if n > 0 {
					_, err = gb.writeBytes(buf[:n])
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
			return relayFrames(gr, gw)
		}

	} else if gb, ok := gr.(*GroupBuffer); ok {
		// Relay bytes to send stream
		if w, ok := gw.(*sendGroupStream); ok {

			offset := 0
			for offset < len(gb.bytes) {
				bytes, err := gb.readBytesAt(offset)
				if err != nil {
					return err
				}
				_, err = w.internalStream.SendStream.Write(bytes)
				if err != nil {
					return err
				}
				offset += len(bytes)
			}

			return nil
		} else if gb2, ok := gw.(*GroupBuffer); ok {
			// Relay bytes to buffer
			offset := 0
			for offset < len(gb.bytes) {
				bytes, err := gb.readBytesAt(offset)
				if err != nil {
					return err
				}
				_, err = gb2.writeBytes(bytes)
				if err != nil {
					return err
				}
				offset += len(bytes)
			}

			return nil
		} else {
			// Relay frames
			return relayFrames(gr, gw)
		}

	} else {
		// Relay frames
		return relayFrames(gr, gw)
	}
}

func relayFrames(gr GroupReader, gw GroupWriter) error {
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

type GroupBuffer struct {
	groupSequence GroupSequence

	bytes []byte

	frameInfo []struct {
		start, end int
	}

	readFrames int

	closed bool

	cond *sync.Cond
}

// NewGroupBuffer creates a new GroupBuffer with the specified group sequence and initial capacity.
func NewGroupBuffer(seq GroupSequence, size int) *GroupBuffer {
	if size <= 0 {
		size = defaultBufferSize
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

		g.cond.Wait()
	}

	frameInfo := g.frameInfo[g.readFrames]

	frame := g.bytes[frameInfo.start:frameInfo.end]

	g.readFrames++

	return frame, nil
}

// GroupWriter: WriteFrame appends a frame to the internal buffer.
func (g *GroupBuffer) WriteFrame(frame []byte) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		return ErrGroupClosed
	}

	if len(g.bytes)+len(frame) > cap(g.bytes) {
		return ErrBufferFull
	}

	g.bytes = message.AppendBytes(g.bytes, frame)

	g.frameInfo = append(g.frameInfo, struct{ start, end int }{
		start: len(g.bytes) - len(frame),
		end:   len(g.bytes),
	})

	g.cond.Broadcast()
	return nil
}

// io.Writer
func (g *GroupBuffer) writeBytes(p []byte) (int, error) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		return 0, ErrGroupClosed
	}

	g.bytes = append(g.bytes, p...)

	offset := 0
	if len(g.frameInfo) > 0 {
		offset = g.frameInfo[len(g.frameInfo)-1].end
	}

	for offset < len(g.bytes) {
		r := bytes.NewReader(g.bytes[offset:])

		frameSize, err := quicvarint.Read(r)
		if err != nil {
			return 0, err
		}

		g.frameInfo = append(g.frameInfo, struct{ start, end int }{
			start: offset + quicvarint.Len(frameSize),
			end:   offset + quicvarint.Len(frameSize) + int(frameSize),
		})

		offset = g.frameInfo[len(g.frameInfo)-1].end
	}

	g.cond.Broadcast()

	return len(p), nil
}

func (g *GroupBuffer) readBytesAt(offset int) ([]byte, error) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if offset < 0 {
		return nil, errors.New("negative offset")
	}

	// Wait until we have data at the requested offset
	for offset >= len(g.bytes) {
		if g.closed {
			return nil, io.EOF
		}
		g.cond.Wait()
	}

	return g.bytes[offset:], nil
}

// Close marks the buffer as closed and wakes up all waiting goroutines
func (g *GroupBuffer) Close() error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	g.closed = true
	g.cond.Broadcast() // Wake up all waiting goroutines
	return nil
}

// Reset resets the buffer with a new group sequence.
func (g *GroupBuffer) Reset(seq GroupSequence) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	g.groupSequence = seq
	g.bytes = g.bytes[:0]
	g.readFrames = 0
	g.frameInfo = g.frameInfo[:0]
	g.closed = false
}

var (
	ErrGroupClosed   = errors.New("group is closed")
	ErrSequenceRange = errors.New("frame sequence is out of range")
	ErrOutOfCache    = errors.New("group buffer is out of cache")
	ErrBufferFull    = errors.New("buffer is full")
)

const defaultBufferSize = 1024 * 1024 // 1MB
