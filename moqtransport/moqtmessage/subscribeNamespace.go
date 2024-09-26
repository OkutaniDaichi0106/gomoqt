package moqtmessage

import "github.com/quic-go/quic-go/quicvarint"

type SubscribeNamespaceMessage struct {
	TrackNamespacePrefix
	Parameters Parameters
}

func (sn SubscribeNamespaceMessage) Serialize() []byte {
	b := make([]byte, 0, 1<<8)

	// Append
	b = quicvarint.Append(b, uint64(SUBSCRIBE_NAMESPACE))

	// Append Track Namespace Prefix
	b = sn.TrackNamespacePrefix.Append(b)

	// Append the Parameters
	b = sn.Parameters.append(b)

	return b
}

func (sn *SubscribeNamespaceMessage) Deserialize(r quicvarint.Reader) error {
	// Get Track Namespace Prefix
	if sn.TrackNamespacePrefix == nil {
		sn.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}

	err := sn.TrackNamespacePrefix.Deserialize(r)
	if err != nil {
		return err
	}

	// Get Parameters
	err = sn.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
