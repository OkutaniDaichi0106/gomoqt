package loc

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type Header struct {
	metadatas map[string]Metadata
}

func (h *Header) AddMetadata(m Metadata) {
	if h.metadatas == nil {
		h.metadatas = make(map[string]Metadata)
	}

	h.metadatas[m.Name] = m
}

func (h Header) Encode(w io.Writer) error {
	b := make([]byte, 0, 1<<6)

	b = quicvarint.Append(b, uint64(len(h.metadatas)))

	if _, err := w.Write(b); err != nil {
		return err
	}

	for _, m := range h.metadatas {
		if err := m.Encode(w); err != nil {
			return err
		}
	}

	return nil
}

func (h *Header) Decode(r io.Reader) error {
	reader := quicvarint.NewReader(r)

	count, err := quicvarint.Read(reader)
	if err != nil {
		return err
	}

	if h.metadatas == nil {
		h.metadatas = make(map[string]Metadata, count)
	}

	for i := uint64(0); i < count; i++ {
		var m Metadata
		if err := m.Decode(r); err != nil {
			return err
		}
		h.metadatas[m.Name] = m
	}

	return nil
}
