package moqtransport

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type Publisher struct {
	client
	MaxSubscribeID uint64
}

func (p *Publisher) Setup(conn Connection) error {
	/*
	 * Open a bidirectional stream
	 */
	stream, err := conn.OpenStream()
	if err != nil {
		return err
	}

	/*
	 * Set the Stream Type to the Setup
	 */
	streamType := SETUP_STREAM
	// Send the Stream Type
	_, err = stream.Write([]byte{byte(streamType)})
	if err != nil {
		return err
	}
	stream.SetType(streamType)

	/*
	 *
	 */
	p.setupStream = stream

	/*
	 * Send a CLIENT_SETUP message
	 */
	csm := moqtmessage.ClientSetupMessage{
		SupportedVersions: p.SupportedVersions,
		Parameters:        make(moqtmessage.Parameters),
	}
	// Add a role parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)
	csm.Parameters.AddParameter(moqtmessage.MAX_SUBSCRIBE_ID, p.MaxSubscribeID)
	// Add a path parameter, when using raw QUIC
	if url := p.conn.URL(); url.Scheme == "moqt" {
		csm.Parameters.AddParameter(moqtmessage.PATH, url.Path)
	}

	_, err = stream.Write(csm.Serialize())
	if err != nil {
		return err
	}

	return nil
}

func (p Publisher) SendDatagram(group moqtmessage.GroupMessage, payload []byte) error {
	// Serialize the header
	b := group.Serialize()

	// Serialize the payload
	b = quicvarint.Append(b, uint64(len(payload)))
	b = append(b, payload...)

	// Send the payload as a datagram
	return p.conn.SendDatagram(b)
}

func (p Publisher) OpenDataStream(group moqtmessage.GroupMessage) (SendStream, error) {
	stream, err := p.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}

	_, err = stream.Write(group.Serialize())
	if err != nil {
		return nil, err
	}

	return stream, nil
}
