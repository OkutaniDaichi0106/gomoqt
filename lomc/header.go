package lomc

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type Header struct {
	metadatas map[string]Metadata
}

func (h *Header) AddMetadata(m Metadata) {
	h.metadatas[m.Name] = m
}

func (h Header) Encode(w io.Writer) error {
	b := make([]byte, 0, 1<<6)

	b = quicvarint.Append(b, uint64(len(h.metadatas)))

	for _, m := range h.metadatas {
		m.Encode(w)
	}

	return nil
}

func (h *Header) Decode(r io.Reader) error {
	var m Metadata
	err := m.Decode(r)
	if err != nil {
		return err
	}

	h.metadatas[m.Name] = m

	return nil
}
