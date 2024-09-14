package moqtransport

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

// Group ID
type groupID uint64

// Peep ID
type peepID uint64

// Object ID
type objectID uint64

/*
 * Track Alias
 * This must be more than 1
 * If This is 0, throw an error
 */
type TrackAlias uint64

// Publisher Priority
type PublisherPriority byte

/*
 * Forwarding Preference
 * Following type are defined in the official document
 * TRACK, GROUP, OBJECT, DATAGRAM
 */
type ForwardingPreference uint64

const (
	TRACK ForwardingPreference = iota
	//GROUP
	//OBJECT
	PEEP
	DATAGRAM
)

// Object Status
type ObjectStatusCode uint64

const (
	NOMAL_OBJECT       ObjectStatusCode = 0x00
	NONEXISTENT_OBJECT ObjectStatusCode = 0x01
	NONEXISTENT_GROUP  ObjectStatusCode = 0x02
	END_OF_GROUP       ObjectStatusCode = 0x03
	END_OF_TRACK       ObjectStatusCode = 0x04
	END_OF_PEEP        ObjectStatusCode = 0x05
)

type TrackNamespace []string

func (tns TrackNamespace) append(b []byte) []byte {
	// Append the number of the elements of the Track Namespace
	b = quicvarint.Append(b, uint64(len(tns)))

	for _, v := range tns {
		// Append the length of the data
		b = quicvarint.Append(b, uint64(len(v)))

		// Append the data
		b = append(b, []byte(v)...)
	}

	return b
}

func (tns TrackNamespace) deserialize(r quicvarint.Reader) error {
	// Get the number of the elements of the track namespace
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < num; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}

		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		tns = append(tns, string(buf))
	}

	return nil
}

type TrackNamespacePrefix []string

func (tns TrackNamespacePrefix) append(b []byte) []byte {
	// Append the number of the elements of the Track Namespace
	b = quicvarint.Append(b, uint64(len(tns)))

	for _, v := range tns {
		// Append the length of the data
		b = quicvarint.Append(b, uint64(len(v)))

		// Append the data
		b = append(b, []byte(v)...)
	}

	return b
}

func (tns TrackNamespacePrefix) deserialize(r quicvarint.Reader) error {
	// Get the number of the elements of the track namespace
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < num; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}

		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		tns = append(tns, string(buf))
	}

	return nil
}
