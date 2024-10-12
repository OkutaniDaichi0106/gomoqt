package moqtransport

import (
	"bytes"
	"context"
	"io"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type Subscriber struct {
	client
	QUICPath url.URL
}

func (s *Subscriber) Setup(conn Connection) error {
	/*
	 * Open an bidirectional stream
	 */
	stream, err := s.conn.OpenStream()
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
	s.setupStream = stream

	/*
	 * Send a CLIENT_SETUP message
	 */
	csm := moqtmessage.ClientSetupMessage{
		SupportedVersions: s.SupportedVersions,
		Parameters:        make(moqtmessage.Parameters),
	}
	// Add a role parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)
	// Add a path parameter, when using raw QUIC
	if url := s.conn.URL(); url.Scheme == "moqt" {
		csm.Parameters.AddParameter(moqtmessage.PATH, url.Path)
	}

	_, err = stream.Write(csm.Serialize())
	if err != nil {
		return err
	}

	return nil
}

func (s Subscriber) ReceiveDatagram(ctx context.Context) (*moqtmessage.GroupMessage, io.Reader, error) {
	/*
	 * Receive a datagram
	 */
	b, err := s.conn.ReceiveDatagram(ctx)
	if err != nil {
		return nil, nil, err
	}

	/*
	 * Read a Group message
	 */
	byteReader := bytes.NewReader(b)
	qvReader := quicvarint.NewReader(byteReader)
	var gm moqtmessage.GroupMessage
	err = gm.DeserializePayload(qvReader)
	if err != nil {
		return nil, nil, err
	}

	return &gm, byteReader, nil
}

func (s Subscriber) AcceptDataStream(ctx context.Context) (*moqtmessage.GroupMessage, ReceiveStream, error) {
	/*
	 * Accept an unidirectional stream
	 */
	stream, err := s.conn.AcceptUniStream(ctx)
	if err != nil {
		return nil, nil, err
	}

	/*
	 * Read a Group message
	 */
	qvReader := quicvarint.NewReader(stream)
	var gm moqtmessage.GroupMessage
	err = gm.DeserializePayload(qvReader)
	if err != nil {
		return nil, nil, err
	}

	return &gm, stream, nil
}
