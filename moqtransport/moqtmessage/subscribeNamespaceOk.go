package moqtmessage

import "github.com/quic-go/quic-go/quicvarint"

type SubscribeNamespaceOk struct {
	TrackNamespacePrefix TrackNamespacePrefix
}

func (sno SubscribeNamespaceOk) Serialize() []byte {
	b := make([]byte, 0, 1<<8)

	// Append message ID
	b = quicvarint.Append(b, uint64(SUBSCRIBE_NAMESPACE_OK))

	// Append Track Namespace Prefix
	b = sno.TrackNamespacePrefix.Append(b)

	return b
}

func (sno *SubscribeNamespaceOk) Deserialize(r quicvarint.Reader) error {
	if sno.TrackNamespacePrefix == nil {
		sno.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}
	err := sno.TrackNamespacePrefix.Deserialize(r)

	return err
}

type UnsubscribeNamespace struct {
	TrackNamespacePrefix TrackNamespacePrefix
}

func (usn UnsubscribeNamespace) Serialize() []byte {
	b := make([]byte, 0, 1<<8)

	// Append message ID
	b = quicvarint.Append(b, uint64(UNSUBSCRIBE_NAMESPACE))

	// Append Track Namespace Prefix
	b = usn.TrackNamespacePrefix.Append(b)

	return b
}

func (usn *UnsubscribeNamespace) Deserialize(r quicvarint.Reader) error {
	if usn.TrackNamespacePrefix == nil {
		usn.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}
	err := usn.TrackNamespacePrefix.Deserialize(r)

	return err
}
