package moqtransport

import (
	"bytes"
	"context"
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/protocol"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Subscriber struct {
	conn              Connection
	SupportedVersions []Version
}

func (s *Subscriber) SetupMORQ(qconn quic.Connection, path string) (*Session, error) {
	// Verify the string follows the path style
	if !isValidPath(path) {
		panic("invalid path")
	}

	//
	conn := newMORQConnection(qconn)
	//
	sess, err := s.setupMORQ(conn, path)
	// Handle any error
	if err != nil {
		// Terminate the Session if the error is a Terminate Error
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
	streamType := setup_stream
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
		SupportedVersions: getProtocolVersions(s.SupportedVersions),
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

func (s Subscriber) SetupMOWT(wtconn *webtransport.Session) (*Session, error) {
	// Get a Connection from the webtransport.Session
	conn := newMOWTConnection(wtconn)

	// Get a Session through a set-up negotiation
	sess, err := s.setupMOWT(conn)

	// Handle any error
	if err != nil {
		// Terminate if the error is a Terminate Error
		if terr, ok := err.(TerminateError); ok {
			wtconn.CloseWithError(webtransport.SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
		}
		return nil, err
	}

	return sess, nil
}

func (s Subscriber) setupMOWT(conn Connection) (*Session, error) {
	/*
	 * Open a bidirectional setupStream
	 */
	stream, err := conn.OpenStream()
	if err != nil {
		return nil, err
	}

	/*
	 * Set the Stream Type to the Setup
	 */
	streamType := setup_stream
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
		SupportedVersions: getProtocolVersions(s.SupportedVersions),
		Parameters:        make(moqtmessage.Parameters),
	}
	// Add a role parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)

	_, err = stream.Write(csm.Serialize())
	if err != nil {
		return nil, err
	}
	/*
	 * Receive a SERVER_SETUP message
	 */
	qvReader := quicvarint.NewReader(stream)
	id, preader, err := moqtmessage.ReadControlMessage(qvReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.SERVER_SETUP {
		return nil, ErrProtocolViolation
	}
	var ssm moqtmessage.ServerSetupMessage
	err = ssm.DeserializePayload(preader)
	if err != nil {
		return nil, err
	}

	// Verify the selected version is valid
	ok := protocol.ContainVersion(ssm.SelectedVersion, getProtocolVersions(s.SupportedVersions))
	if !ok {
		return nil, ErrProtocolViolation
	}

	return &Session{
		Connection:       conn,
		setupStream:      stream,
		selectedVersion:  ssm.SelectedVersion,
		trackAliasMap:    new(trackAliasMap),
		subscribeCounter: 0,
	}, nil
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
