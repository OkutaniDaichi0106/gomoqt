package moqtransport

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Server struct {
	Addr string

	Port int

	TLSConfig *tls.Config

	QUICConfig *quic.Config

	SupportedVersions []moqtmessage.Version

	WTConfig struct {
		ReorderingTimeout time.Duration

		CheckOrigin func(r *http.Request) bool

		EnableDatagrams bool
	}

	SetupHijacker func(moqtmessage.Parameters) (moqtmessage.Parameters, error)

	wts *webtransport.Server
}

func (s Server) WebtransportUpgrade(w http.ResponseWriter, r *http.Request) (*webtransport.Session, error) {
	if s.wts == nil {
		s.wts = &webtransport.Server{
			H3: http3.Server{
				Addr:            s.Addr,
				Port:            s.Port,
				TLSConfig:       s.TLSConfig,
				QUICConfig:      s.QUICConfig,
				EnableDatagrams: s.WTConfig.EnableDatagrams,
			},
			ReorderingTimeout: s.WTConfig.ReorderingTimeout,
			CheckOrigin:       s.WTConfig.CheckOrigin,
		}
	}

	wtsess, err := s.wts.Upgrade(w, r)

	if err != nil {
		return nil, err
	}

	return wtsess, nil
}

func (s Server) SetupMORQ(qconn quic.Connection) (*Session, string, error) {
	conn := newMORQConnection(qconn)

	// Setup
	sess, path, err := s.setupMORQ(conn)

	// Terminate the connection when Terminate Error occured
	if err != nil {
		if terr, ok := err.(TerminateError); ok {
			qconn.CloseWithError(quic.ApplicationErrorCode(terr.TerminateErrorCode()), terr.Error())
		}
		return nil, "", err
	}

	return sess, path, nil
}

func (s Server) setupMORQ(conn Connection) (*Session, string, error) {
	/*
	 * Accept a bidirectional stream
	 */
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, "", err
	}

	err = acceptSetupStream(stream)
	if err != nil {
		return nil, "", err
	}

	/*
	 * Receive a CLIENT_SETUP message
	 */
	qvReader := quicvarint.NewReader(stream)
	id, preader, err := moqtmessage.ReadControlMessage(qvReader)
	if err != nil {
		return nil, "", err
	}
	if id != moqtmessage.CLIENT_SETUP {
		return nil, "", ErrProtocolViolation
	}
	var csm moqtmessage.ClientSetupMessage
	err = csm.DeserializePayload(preader)
	if err != nil {
		return nil, "", err
	}
	// Verify if a ROLE parameter exists
	role, ok := csm.Parameters.Role()
	if !ok {
		return nil, "", ErrProtocolViolation
	} else if role != moqtmessage.PUB && role != moqtmessage.SUB && role != moqtmessage.PUB_SUB {
		return nil, "", ErrProtocolViolation
	}
	// Get a MAX_SUBSCRIBE_ID parameter
	maxID, ok := csm.Parameters.MaxSubscribeID()
	if !ok {
		maxID = 0
	}
	// Get a PATH parameter when using raw QUIC
	var path string
	if _, ok := conn.(*rawQuicConnection); ok {
		path, ok = csm.Parameters.Path()
		if !ok {
			return nil, "", ErrProtocolViolation
		}
	}

	// Handle Parameters in a SERVER_SETUP message
	ssparams := make(moqtmessage.Parameters)
	if s.SetupHijacker != nil {
		ssparams, err = s.SetupHijacker(csm.Parameters)
		if err != nil {
			return nil, "", err
		}
	}

	/*
	 * Select the latest version supported by both the client and the server
	 */
	selectedVersion, err := moqtmessage.SelectLatestVersion(s.SupportedVersions, csm.SupportedVersions)
	if err != nil {
		return nil, "", err
	}

	/*
	 * Send a SERVER_SETUP message
	 */
	// Initialize a SERVER_SETUP message
	ssm := moqtmessage.ServerSetupMessage{
		SelectedVersion: selectedVersion,
		Parameters:      make(moqtmessage.Parameters),
	}
	// ROLE Parameter
	switch role {
	case moqtmessage.PUB:
		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)
	case moqtmessage.SUB:
		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)
	case moqtmessage.PUB_SUB:
		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB_SUB)
	}
	// Optional Parameters
	for k, v := range ssparams {
		ssm.Parameters.AddParameter(k, v)
	}
	// Send the message
	_, err = stream.Write(ssm.Serialize())
	if err != nil {
		return nil, "", err
	}

	return &Session{
		Connection:       conn,
		setupStream:      stream,
		selectedVersion:  selectedVersion,
		trackAliasMap:    new(trackAliasMap),
		subscribeCounter: 0,
		maxSubscribeID:   &maxID,
	}, path, nil
}

func (s Server) SetupMOWT(wtconn *webtransport.Session) (*Session, error) {
	conn := newMOWTConnection(wtconn)
	sess, err := s.setupMOWT(conn)
	if err != nil {
		// Terminate the connection if the error is a Terminate Error
		if terr, ok := err.(TerminateError); ok {
			conn.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
		}
	}

	return sess, nil
}
func (s Server) setupMOWT(conn Connection) (*Session, error) {

	/*
	 * Accept a bidirectional stream
	 */
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}

	err = acceptSetupStream(stream)
	if err != nil {
		return nil, err
	}

	/*
	 * Receive a CLIENT_SETUP message
	 */
	qvReader := quicvarint.NewReader(stream)
	id, preader, err := moqtmessage.ReadControlMessage(qvReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.CLIENT_SETUP {
		return nil, ErrProtocolViolation
	}
	var csm moqtmessage.ClientSetupMessage
	err = csm.DeserializePayload(preader)
	if err != nil {
		return nil, err
	}
	// Verify if a ROLE parameter exists
	role, ok := csm.Parameters.Role()
	if !ok {
		return nil, ErrProtocolViolation
	} else if role != moqtmessage.PUB && role != moqtmessage.SUB && role != moqtmessage.PUB_SUB {
		return nil, ErrProtocolViolation
	}
	// Get a MAX_SUBSCRIBE_ID parameter
	maxID, ok := csm.Parameters.MaxSubscribeID()
	if !ok {
		maxID = 0
	}
	// Get a PATH parameter and close the connection
	if _, ok := conn.(*rawQuicConnection); ok {
		_, ok = csm.Parameters.Path()
		if ok {
			return nil, ErrProtocolViolation
		}
	}

	// Handle Parameters in a SERVER_SETUP message
	ssparams := make(moqtmessage.Parameters)
	if s.SetupHijacker != nil {
		ssparams, err = s.SetupHijacker(csm.Parameters)
		if err != nil {
			return nil, err
		}
	}

	/*
	 * Select the latest version supported by both the client and the server
	 */
	selectedVersion, err := moqtmessage.SelectLatestVersion(s.SupportedVersions, csm.SupportedVersions)
	if err != nil {
		return nil, err
	}

	/*
	 * Send a SERVER_SETUP message
	 */
	// Initialize a SERVER_SETUP message
	ssm := moqtmessage.ServerSetupMessage{
		SelectedVersion: selectedVersion,
		Parameters:      make(moqtmessage.Parameters),
	}
	// ROLE Parameter
	switch role {
	case moqtmessage.PUB:
		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)
	case moqtmessage.SUB:
		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)
	case moqtmessage.PUB_SUB:
		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB_SUB)
	}
	// Optional Parameters
	for k, v := range ssparams {
		ssm.Parameters.AddParameter(k, v)
	}
	// Send the message
	_, err = stream.Write(ssm.Serialize())
	if err != nil {
		return nil, err
	}

	return &Session{
		Connection:       conn,
		setupStream:      stream,
		selectedVersion:  selectedVersion,
		trackAliasMap:    new(trackAliasMap),
		subscribeCounter: 0,
		maxSubscribeID:   &maxID,
	}, nil
}

func acceptSetupStream(stream Stream) error {
	/*
	 *
	 */
	// Read the Stream Type
	qvReader := quicvarint.NewReader(stream)
	num, err := qvReader.ReadByte()
	if err != nil {
		return err
	}
	// verify the Stream Type ID
	if StreamType(num) != SETUP_STREAM {
		log.Println(stream.Close())
		return ErrUnexpectedStreamType
	}
	// Set the Stream Type to the Setup
	stream.SetType(SETUP_STREAM)

	return nil
}
