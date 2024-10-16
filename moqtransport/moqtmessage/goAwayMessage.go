package moqtmessage

import "github.com/quic-go/quic-go/quicvarint"

type GoAwayMessage struct {
	/*
	 * New session URI
	 * If this is 0 byte, this should be set to current session URI
	 */
	NewSessionURI string
}

func (ga GoAwayMessage) Serialize() []byte {
	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<6)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(ga.NewSessionURI)))
	p = append(p, []byte(ga.NewSessionURI)...)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	b = quicvarint.Append(b, uint64(GOAWAY))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (ga *GoAwayMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get length of the URI
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// if num == 0 {
	// 	// TODO: Reuse currenct URI
	// }

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
