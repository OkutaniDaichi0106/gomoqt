package moqtransport

import "github.com/quic-go/quic-go/quicvarint"

type SubscribeNamespace struct {
	TrackNamespacePrefix
	Parameters
}

func (sn SubscribeNamespace) serialize() []byte {
	b := make([]byte, 0, 1<<8)

	// Append
	b = quicvarint.Append(b, uint64(SUBSCRIBE_NAMESPACE))

	// Append Track Namespace Prefix
	b = sn.TrackNamespacePrefix.append(b)

	// Append the Parameters
	b = sn.Parameters.append(b)

	return b
}

func (sn *SubscribeNamespace) deserialize(r quicvarint.Reader) error {
	// Get Track Namespace Prefix
	if sn.TrackNamespacePrefix == nil {
		sn.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}

	err := sn.TrackNamespacePrefix.deserialize(r)
	if err != nil {
		return err
	}

	// Get Parameters
	err = sn.Parameters.deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
