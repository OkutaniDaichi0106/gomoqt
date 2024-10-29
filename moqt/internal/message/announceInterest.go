package message

import "github.com/quic-go/quic-go/quicvarint"

type AnnounceInterestMessage struct {
	TrackPrefix TrackPrefix
	Parameters  Parameters
}

func (aim AnnounceInterestMessage) SerializePayload() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE_NAMESPACE Message Payload {
	 *   Track Namespace prefix (tuple),
	 *   Number of Parameters (varint),
	 *   Subscribe Parameters (..),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace Prefix
	p = AppendTrackNamespacePrefix(p, aim.TrackPrefix)

	// Append the Parameters
	p = aim.Parameters.Append(p)

	return p
}

func (aim *AnnounceInterestMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get a Track Namespace Prefix
	tnsp, err := ReadTrackNamespacePrefix(r)
	if err != nil {
		return err
	}
	aim.TrackPrefix = tnsp

	// Get Parameters
	err = aim.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
