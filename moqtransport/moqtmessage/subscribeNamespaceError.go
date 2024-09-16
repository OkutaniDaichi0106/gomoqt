package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeNamespaceError struct {
	TrackNamespacePrefix TrackNamespacePrefix
	Code                 SubscribeNamespaceErrorCode
	Reason               string
}

type SubscribeNamespaceErrorCode uint

func (sne SubscribeNamespaceError) Serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SUBSCRIBE_NAMESPACE_ERROR Message {
	 *  Track Namespace (tuple),
	 *  Error Code (varint),
	 *  Reason ([]byte),
	 * }
	 */
	b := make([]byte, 0, 1<<8)

	//Append message ID
	b = quicvarint.Append(b, uint64(SUBSCRIBE_NAMESPACE_ERROR))

	// Append Track Namespace Prefix
	b = sne.TrackNamespacePrefix.Append(b)

	// Append Error Code
	b = quicvarint.Append(b, uint64(sne.Code))

	// Append Error Reason
	b = quicvarint.Append(b, uint64(len(sne.Reason)))
	b = append(b, []byte(sne.Reason)...)

	return b
}

func (sne *SubscribeNamespaceError) Deserialize(r quicvarint.Reader) error {
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
