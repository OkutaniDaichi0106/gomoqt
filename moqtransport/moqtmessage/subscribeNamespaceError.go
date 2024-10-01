package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeNamespaceErrorMessage struct {
	TrackNamespacePrefix TrackNamespacePrefix
	Code                 SubscribeNamespaceErrorCode
	Reason               string
}

type SubscribeNamespaceErrorCode uint

func (sne SubscribeNamespaceErrorMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE_NAMESPACE_ERROR Message {
	 *  Track Namespace Prefix (tuple),
	 *  Error Code (varint),
	 *  Reason ([]byte),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace Prefix
	p = sne.TrackNamespacePrefix.Append(p)

	// Append the Error Code
	p = quicvarint.Append(p, uint64(sne.Code))

	// Append the Error Reason
	p = quicvarint.Append(p, uint64(len(sne.Reason)))
	p = append(p, []byte(sne.Reason)...)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	b = quicvarint.Append(b, uint64(SUBSCRIBE_NAMESPACE_ERROR))

	// Append the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (sne *SubscribeNamespaceErrorMessage) DeserializePayload(r quicvarint.Reader) error {
	if sne.TrackNamespacePrefix == nil {
		sne.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}
	err := sne.TrackNamespacePrefix.Deserialize(r)
	if err != nil {
		return err
	}

	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	sne.Code = SubscribeNamespaceErrorCode(num)

	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)

	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	sne.Reason = string(buf)

	return nil
}
