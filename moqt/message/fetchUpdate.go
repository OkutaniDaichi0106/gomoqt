package message

import "github.com/quic-go/quic-go/quicvarint"

type FetchUpdateMessage struct{}

func (FetchUpdateMessage) SerializePayload() []byte {
	p := make([]byte, 1<<6)
	return p
}

func (FetchUpdateMessage) DeserializePayload(r quicvarint.Reader) error {
	return nil
}
