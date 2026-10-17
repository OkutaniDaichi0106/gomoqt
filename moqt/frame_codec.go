package moqt

import "io"

// FrameWriter writes frames to an io.Writer in the form they take on a group
// stream: the FRAME messages that follow the GROUP header, each timestamp a
// zigzag delta from the previous frame's, the first from 0. Those bytes are a
// group's frames as the wire carries them, so a store that keeps them replays
// the group unchanged.
type FrameWriter struct {
	w io.Writer
	// prevTimestamp is the timestamp the next frame's delta is taken from.
	prevTimestamp uint64
}

// NewFrameWriter returns a FrameWriter writing a group's frames to w.
func NewFrameWriter(w io.Writer) *FrameWriter {
	return &FrameWriter{w: w}
}

// WriteFrame writes frame, made by [NewFrame], after the frames written
// before it. A nil frame writes nothing, as [GroupWriter.WriteFrame].
func (fw *FrameWriter) WriteFrame(frame *Frame) error {
	if frame == nil {
		return nil
	}
	if err := frame.encode(fw.w, fw.prevTimestamp); err != nil {
		return err
	}
	fw.prevTimestamp = frame.Timestamp
	return nil
}

// FrameReader reads the frames a [FrameWriter] wrote, or a group stream's
// frames after its GROUP header, from an io.Reader.
type FrameReader struct {
	r io.Reader
	// prevTimestamp is the timestamp the next frame's delta applies to.
	prevTimestamp uint64
}

// NewFrameReader returns a FrameReader reading a group's frames from r.
func NewFrameReader(r io.Reader) *FrameReader {
	return &FrameReader{r: r}
}

// ReadFrame decodes the next frame into frame, reusing its buffer. It
// returns io.EOF after the last whole frame, io.ErrUnexpectedEOF when r ends
// inside one, and [ErrTimestampOutOfRange] for a timestamp below zero.
func (fr *FrameReader) ReadFrame(frame *Frame) error {
	if frame == nil {
		panic("nil frame")
	}
	if err := frame.decode(fr.r, fr.prevTimestamp); err != nil {
		return err
	}
	fr.prevTimestamp = frame.Timestamp
	return nil
}
