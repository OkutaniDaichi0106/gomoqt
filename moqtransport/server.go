package moqtransport

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
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
}

func (s Server) WebTransportServer() webtransport.Server {
	return webtransport.Server{
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

func (s Server) Setup(conn Connection) (*Session, error) {
	/*
	 * Accept a bidirectional stream
	 */
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}
	/*
	 *
	 */
	// Read the Stream Type
	qvReader := quicvarint.NewReader(stream)
	num, err := qvReader.ReadByte()
	if err != nil {
		return nil, err
	}
	// verify the Stream Type ID
	if StreamType(num) != SETUP_STREAM {
		log.Println(stream.Close())
		return nil, ErrUnexpectedStreamType
	}
	// Set the Stream Type to the Setup
	stream.SetType(SETUP_STREAM)

	/*
	 * Receive a CLIENT_SETUP message
	 */
	id, preader, err := moqtmessage.ReadControlMessage(qvReader)
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
	// Get a PATH parameter when using a raw quic connection
	rqconn, ok := conn.(*rawQuicConnection)
	if ok {
		path, ok := csm.Parameters.AuthorizationInfo()
		if !ok {
			return nil, ErrProtocolViolation
		}
		rqconn.url = url.URL{
			Scheme: "moqt",
			Path:   path,
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
		sessionStream:    stream,
		selectedVersion:  selectedVersion,
		trackAliasMap:    new(trackAliasMap),
		subscribeCounter: 0,
		maxSubscribeID:   maxID,
	}, nil
}
