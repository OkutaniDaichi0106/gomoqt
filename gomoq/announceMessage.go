package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceMessage struct {
	/*
	 * Track Namespace
	 */
	TrackNamespace string

	/*
	 * Announce Parameters
	 * Parameters should include track authorization information
	 */
	Parameters Parameters
}

func (a AnnounceMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * ANNOUNCE Payload {
	 *   Track Namespace ([]byte),
	 *   Number of Parameters (),
	 *   Announce Parameters(..)
	 * }
	 */

	// TODO?: Chech track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(a.TrackNamespace)))
	b = append(b, []byte(a.TrackNamespace)...)

	// Serialize the parameters and append it
	/*
	 * Announce Parameters {
	 *   [Authorization Info Parameter (stirng)],
	 *   [Optional Patameters(..)],
	 * }
	 */
	b = a.Parameters.append(b)

	return b
}

func (a *AnnounceMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != ANNOUNCE {
		return errors.New("unexpected message")
	}

	// Get length of the track namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get track namespace
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	a.TrackNamespace = string(buf)

	err = a.Parameters.parse(r)
	if err != nil {
		return err
	}

	return nil
}

type AnnounceOkMessage struct {
	TrackNamespace string
}

func (ao AnnounceOkMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * ANNOUNCE_OK Payload {
	 *   Track Namespace ([]byte),
	 * }
	 */

	// TODO?: Chech track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_OK))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ao.TrackNamespace)))
	b = append(b, []byte(ao.TrackNamespace)...)

	return b
}

func (ao *AnnounceOkMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != ANNOUNCE_OK {
		return errors.New("unexpected message")
	}

	// Get length of the string of the track namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get track namespace
	buf := make([]byte, num)

	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	ao.TrackNamespace = string(buf)

	// Just one track namespace supposed to be detected
	// Over one track namespace will not be detected

	return nil
}

type UnannounceMessage struct {
	TrackNamespace string
}

func (ua UnannounceMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * UNANNOUNCE Payload {
	 *   Track Namespace ([]byte),
	 * }
	 */

	// TODO?: Chech track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(UNANNOUNCE))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ua.TrackNamespace)))
	b = append(b, []byte(ua.TrackNamespace)...)

	return b
}

func (ua *UnannounceMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != UNANNOUNCE {
		return errors.New("unexpected message")
	}

	// Get length of the string of the track namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get track namespace
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	ua.TrackNamespace = string(buf)

	// Just one track namespace supposed to be detected
	// Over one track namespace will not be detected

	return nil
}

type AnnounceCancelMessage struct {
	TrackNamespace string
}

func (ua AnnounceCancelMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * ANNOUNCE_CANCEL Payload {
	 *   Track Namespace ([]byte),
	 * }
	 */

	// TODO?: Check track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_CANCEL))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ua.TrackNamespace)))
	b = append(b, []byte(ua.TrackNamespace)...)

	return b
}

func (ac *AnnounceCancelMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != ANNOUNCE_CANCEL {
		return errors.New("unexpected message")
	}

	// Get length of the string of the namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	buf := make([]byte, num)
	// Get track namespace
	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	ac.TrackNamespace = string(buf)

	// Just one track namespace supposed to be detected
	// Over one track namespace will not be detected

	return nil
}
