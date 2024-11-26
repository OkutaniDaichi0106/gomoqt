package message

import "github.com/quic-go/quic-go/quicvarint"

type GoAwayMessage struct {
	/*
	 * New session URI
	 * If this is 0 byte, this should be set to current session URI
	 */
	NewSessionURI string
}

func (ga GoAwayMessage) SerializePayload() []byte {
	/*
	 * Serialize the payload in the following format
	 *
	 * GOAWAY Message Payload {
	 *   New Session URI (string),
	 * }
	 */
	p := make([]byte, 0, 1<<6)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(ga.NewSessionURI)))
	p = append(p, []byte(ga.NewSessionURI)...)

	return p
}

func (ga *GoAwayMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get length of the URI
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get URI
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ga.NewSessionURI = string(buf)

	// Just one URI supposed to be detected
	// Over one URI will not be detected

	return nil
}
