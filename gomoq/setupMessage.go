package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type ClientSetupMessage struct {
	/*
	 * Versions supported by the client
	 */
	Versions []Version

	/*
	 * Setup Parameters
	 */
	Parameters
}

func (cs ClientSetupMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * CLIENT_SETUP Payload {
	 *   Number of Supported Versions (varint),
	 *   Supported Versions (varint),
	 *   Number of Parameters (varint),
	 *   Setup Parameters (..),
	 * }
	 */

	// Check the condition of the parameters
	// 1. At least one version is required
	if len(cs.Versions) == 0 {
		panic("no version is specifyed")
	}
	// 2. Parameters should conclude role parameter
	if !cs.Parameters.Contain(role) {
		panic("no role is specifyed")
	}

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(CLIENT_SETUP))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(cs.Versions)))
	var version Version
	for _, version = range cs.Versions {
		b = quicvarint.Append(b, uint64(version))
	}

	// Serialize the parameters and append it
	/*
	 * Setup Parameters {
	 *   Role Parameter (varint),
	 *   [Path Parameter (stirng),]
	 *   [Optional Patameters(..)],
	 * }
	 */
	b = cs.Parameters.append(b)

	return b
}

func (cs *ClientSetupMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != CLIENT_SETUP {
		return errors.New("unexpected message")
	}

	// Get number of supported versions
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	count := num

	// Get supported versions
	for i := uint64(0); i < count; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		cs.Versions = append(cs.Versions, Version(num))
	}

	err = cs.Parameters.parse(r)
	if err != nil {
		return err
	}
	return nil
}

type ServerSetupMessage struct {
	/*
	 * Versions selected by the server
	 */
	SelectedVersion Version

	/*
	 * Setup Parameters
	 * Keys of the maps should not be duplicated
	 */
	Parameters
}

func (ss ServerSetupMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SERVER_SETUP Payload {
	 *   Selected Version (varint),
	 *   Number of Parameters (varint),
	 *   Setup Parameters (..),
	 * }
	 */
	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<8) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(SERVER_SETUP))
	// Append the selected version
	b = quicvarint.Append(b, uint64(ss.SelectedVersion))

	// Serialize the parameters and append it
	/*
	 * Setup Parameters {
	 *   [Optional Patameters(..)],
	 * }
	 */
	b = ss.Parameters.append(b)

	return b
}

func (ss *ServerSetupMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != SERVER_SETUP {
		return errors.New("unexpected message")
	}

	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ss.SelectedVersion = Version(num)

	err = ss.Parameters.parse(r)
	if err != nil {
		return err
	}
	return nil

	// PATH parameter must not be included in the parameters
	// Keys of the parameters must not be duplicate
}
