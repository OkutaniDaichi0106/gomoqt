package moqt

import (
	"io"
)

type BufferStream struct {
	group  Group
	buffer []byte
	src    Stream
	ended  bool
}

func (stream BufferStream) Read(buf []byte) (int, error) {
	if stream.buffer == nil {
		stream.buffer = make([]byte, 0, 1<<10)
	}

	n, err := stream.src.Read(buf)
	if err != nil && err != io.EOF {
		return 0, err
	} else if err == io.EOF {
		stream.ended = true
	}

	stream.buffer = append(stream.buffer, buf[:n]...)

	return n, err
}

func (stream BufferStream) ReadOffset(buf []byte, offset uint64) (int, error) {
	if len(stream.buffer[offset:]) < 1 {
		if stream.ended {
			return 0, io.EOF
		}

		sub := make([]byte, len(buf))
		n, err := stream.Read(sub)
		if err != nil {
			return n, err
		}

		return stream.ReadOffset(buf, offset)
	}

	n := copy(buf, stream.buffer)

	return n, nil
}
