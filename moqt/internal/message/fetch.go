package message

import "github.com/quic-go/quic-go/quicvarint"

type FetchMessage struct {
	TrackNamespace     TrackNamespace
	TrackName          string
	SubscriberPriority SubscriberPriority
	GroupSequence      GroupSequence
	GroupOffset        uint64
}

func (FetchMessage) SerializePayload() []byte {
	p := make([]byte, 1<<6)
	return p
}

func (FetchMessage) DeserializePayload(r quicvarint.Reader) error {
	return nil
}
