package moqtransport

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Publisher struct {
	Client
	MaxSubscribeID uint64
}

func (p *Publisher) SetupMORQ(qconn quic.Connection, path string) (*Session, error) {
	if !isValidPath(path) {
		panic("invalid path")
	}

	conn := newMORQConnection(qconn)
	sess, err := p.setupMORQ(conn, path)
	if err != nil {
		if terr, ok := err.(TerminateError); ok {
			conn.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
		}
		return nil, err
	}

	return sess, nil
}
func (p *Publisher) setupMORQ(conn Connection, path string) (*Session, error) {
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
		SupportedVersions: p.SupportedVersions,
		Parameters:        make(moqtmessage.Parameters),
	}
	// Add a role parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)
	csm.Parameters.AddParameter(moqtmessage.MAX_SUBSCRIBE_ID, p.MaxSubscribeID)
	// Add a path parameter
	if !isValidPath(path) {
		panic("invalid path")
	}
	csm.Parameters.AddParameter(moqtmessage.PATH, path)

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
	ok := moqtmessage.ContainVersion(ssm.SelectedVersion, p.SupportedVersions)
	if !ok {
		return nil, ErrProtocolViolation
	}

	return &Session{
		Connection:       conn,
		setupStream:      stream,
		selectedVersion:  ssm.SelectedVersion,
		trackAliasMap:    new(trackAliasMap),
		subscribeCounter: 0,
		maxSubscribeID:   (*moqtmessage.SubscribeID)(&p.MaxSubscribeID),
	}, nil
}

func (p *Publisher) SetupMOWT(wtconn *webtransport.Session) (*Session, error) {
	conn := newMOWTConnection(wtconn)
	sess, err := p.setupMOWT(conn)
	if err != nil {
		if terr, ok := err.(TerminateError); ok {
			wtconn.CloseWithError(webtransport.SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
		}
		return nil, err
	}

	return sess, nil
}

func (p *Publisher) setupMOWT(conn Connection) (*Session, error) {
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
		SupportedVersions: p.SupportedVersions,
		Parameters:        make(moqtmessage.Parameters),
	}
	// Add a role parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)
	csm.Parameters.AddParameter(moqtmessage.MAX_SUBSCRIBE_ID, p.MaxSubscribeID)

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
	ok := moqtmessage.ContainVersion(ssm.SelectedVersion, p.SupportedVersions)
	if !ok {
		return nil, ErrProtocolViolation
	}

	return &Session{
		Connection:       conn,
		setupStream:      stream,
		selectedVersion:  ssm.SelectedVersion,
		trackAliasMap:    new(trackAliasMap),
		subscribeCounter: 0,
		maxSubscribeID:   (*moqtmessage.SubscribeID)(&p.MaxSubscribeID),
	}, nil
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
