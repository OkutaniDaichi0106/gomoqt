package loc

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	capture_timestamp   MetadataID = 0x0
	video_frame_marking MetadataID = 0x1
	audio_level         MetadataID = 0x2
)

type MetadataID uint64

type Metadata struct {
	Name        string
	Description string
	ID          MetadataID
	Value       []byte
}

func (h Metadata) Encode(io.Writer) error {
	b := make([]byte, 0, 1<<6)

	b = quicvarint.Append(b, uint64(h.ID))

	b = quicvarint.Append(b, uint64(len(h.Value)))

	b = append(b, h.Value...)

	return nil
}

func (h *Metadata) Decode(r io.Reader) error {
	reader := quicvarint.NewReader(r)

	num, err := quicvarint.Read(reader)
	if err != nil {
		return err
	}
	h.ID = MetadataID(num)

	num, err = quicvarint.Read(reader)
	if err != nil {
		return err
	}

	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	h.Value = buf

	return nil
}
