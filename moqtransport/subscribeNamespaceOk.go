package moqtransport

import "github.com/quic-go/quic-go/quicvarint"

type SubscribeNamespaceOk struct {
	TrackNamespacePrefix TrackNamespacePrefix
}

func (sno SubscribeNamespaceOk) serialize() []byte {
	b := make([]byte, 0, 1<<8)

	// Append message ID
	b = quicvarint.Append(b, uint64(SUBSCRIBE_NAMESPACE_OK))

	// Append Track Namespace Prefix
	b = sno.TrackNamespacePrefix.append(b)

	return b
}

func (sno *SubscribeNamespaceOk) deserialize(r quicvarint.Reader) error {
	if sno.TrackNamespacePrefix == nil {
		sno.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}
	err := sno.TrackNamespacePrefix.deserialize(r)

	return err
}

type SubscribeNamespaceError struct {
	TrackNamespacePrefix TrackNamespacePrefix
	Code                 SubscribeNamespaceErrorCode
	Reason               string
}
type SubscribeNamespaceErrorCode int

func (sne SubscribeNamespaceError) serialize() []byte {
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
	b = sne.TrackNamespacePrefix.append(b)

	// Append Error Code
	b = quicvarint.Append(b, uint64(sne.Code))

	// Append Error Reason
	b = quicvarint.Append(b, uint64(len(sne.Reason)))
	b = append(b, []byte(sne.Reason)...)

	return b
}

func (sne *SubscribeNamespaceError) deserialize(r quicvarint.Reader) error {
	if sne.TrackNamespacePrefix == nil {
		sne.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}
	err := sne.TrackNamespacePrefix.deserialize(r)
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

type UnsubscribeNamespace struct {
	TrackNamespacePrefix TrackNamespacePrefix
}

func (usn UnsubscribeNamespace) serialize() []byte {
	b := make([]byte, 0, 1<<8)

	// Append message ID
	b = quicvarint.Append(b, uint64(UNSUBSCRIBE_NAMESPACE))

	// Append Track Namespace Prefix
	b = usn.TrackNamespacePrefix.append(b)

	return b
}

func (usn *UnsubscribeNamespace) deserialize(r quicvarint.Reader) error {
	if usn.TrackNamespacePrefix == nil {
		usn.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}
	err := usn.TrackNamespacePrefix.deserialize(r)

	return err
}
