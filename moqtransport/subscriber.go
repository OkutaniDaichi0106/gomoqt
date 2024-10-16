package moqtransport

import (
	"bytes"
	"context"
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

type Subscriber struct {
	Client
}

func (s *Subscriber) SetupMORQ(qconn quic.Connection, path string) (*Session, error) {
	if !isValidPath(path) {
		panic("invalid path")
	}

	conn := newMORQConnection(qconn)
	sess, err := s.setupMORQ(conn, path)
	if err != nil {
		if terr, ok := err.(TerminateError); ok {
			conn.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
		}
	}

	return sess, nil
}

func (s *Subscriber) setupMORQ(conn Connection, path string) (*Session, error) {
	/*
	 * Open an bidirectional stream
	 */
	s.conn = conn
	stream, err := s.conn.OpenStream()
	if err != nil {
		return nil, err
	}

	/*
	 * Set the Stream Type to the Setup
	 */
	streamType := SETUP_STREAM
	// Send the Stream Type
	_, err = stream.Write([]byte{byte(streamType)})
	if err != nil {
		return nil, err
	}
	stream.SetType(streamType)

	/*
	 * Send a CLIENT_SETUP message
	 */
	csm := moqtmessage.ClientSetupMessage{
		SupportedVersions: s.SupportedVersions,
		Parameters:        make(moqtmessage.Parameters),
	}
	// Add the ROLE parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)
	// Add the path parameter
	csm.Parameters.AddParameter(moqtmessage.PATH, path)

	_, err = stream.Write(csm.Serialize())
	if err != nil {
		return nil, err
	}

	return &Session{}, nil
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
