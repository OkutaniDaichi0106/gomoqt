package moqtransport

import "github.com/quic-go/quic-go/quicvarint"

type GoAwayMessage struct {
	/*
	 * New session URI
	 * If this is 0 byte, this should be set to current session URI
	 */
	NewSessionURI string
}

func (ga GoAwayMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * GOAWAY Payload {
	 *   New Session URI ([]byte),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(GOAWAY))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ga.NewSessionURI)))
	b = append(b, []byte(ga.NewSessionURI)...)

	return b
}

func (ga *GoAwayMessage) deserializeBody(r quicvarint.Reader) error {
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
