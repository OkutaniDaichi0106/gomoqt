package moqtransport

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtversion"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Role byte

const (
	PUB     Role = 0x00
	SUB     Role = 0x01
	PUB_SUB Role = 0x02
)

type node struct {
	/*
	 * TLS configuration
	 */
	TLSConfig *tls.Config

	/*
	 * QUIC configuration
	 */
	QUICConfig *quic.Config

	/*
	 * Versions supported by the node
	 */
	SupportedVersions []moqtversion.Version
}

func (n node) EstablishPubSession(URL string, maxSubscribeID uint64) (*PublishingSession, error) {
	// Parse the url strings
	u, err := url.Parse(URL)

	// Connect to the server as a publisher
	trSess, err := n.connect(u)
	if err != nil {
		return nil, err
	}

	// Open bidirectional stream to send control messages
	controlStream, err := trSess.OpenStream()
	if err != nil {
		return nil, err
	}

	// Get reader of the bidirectional stream
	controlReader := quicvarint.NewReader(controlStream)

	// Set up
	// Initialize SETUP_CLIENT message
	csm := moqtmessage.ClientSetupMessage{
		Versions:   n.SupportedVersions,
		Parameters: make(moqtmessage.Parameters),
	}

	// Add the ROLE parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)

	// Add the MAX_SUBSCRIBE_ID parameter
	csm.Parameters.AddParameter(moqtmessage.MAX_SUBSCRIBE_ID, maxSubscribeID)

	// Add PATH parameter if use raw quic connection
	if u.Scheme == "moqt" {
		csm.Parameters.AddParameter(moqtmessage.PATH, u.Path)
	}

	_, err = controlStream.Write(csm.Serialize())
	if err != nil {
		return nil, err
	}

	version, _, err := n.receiveServerSetupMessage(controlReader)
	if err != nil {
		return nil, err
	}

	sessionCore := sessionCore{
		trSess:          trSess,
		controlStream:   controlStream,
		controlReader:   controlReader,
		selectedVersion: version,
	}

	// Register the session
	sessions = append(sessions, &sessionCore)

	session := PublishingSession{
		sessionCore:    sessionCore,
		maxSubscribeID: moqtmessage.SubscribeID(maxSubscribeID),
	}

	return &session, nil
}

func (n node) EstablishSubSession(URL string) (*SubscribingSession, error) {
	// Parse the url strings
	u, err := url.Parse(URL)

	// Connect to the server as a subscriber
	trSess, err := n.connect(u)
	if err != nil {
		return nil, err
	}

	// Open bidirectional stream to send control messages
	controlStream, err := trSess.OpenStream()
	if err != nil {
		return nil, err
	}

	// Get reader of the bidirectional stream
	controlReader := quicvarint.NewReader(controlStream)

	// Set up
	// Initialize SETUP_CLIENT message
	csm := moqtmessage.ClientSetupMessage{
		Versions:   n.SupportedVersions,
		Parameters: make(moqtmessage.Parameters),
	}

	// Add role parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)

	_, err = controlStream.Write(csm.Serialize())
	if err != nil {
		return nil, err
	}

	version, _, err := n.receiveServerSetupMessage(controlReader) // TODO: Handle the parameter
	if err != nil {
		return nil, err
	}

	sessionCore := sessionCore{
		trSess:          trSess,
		controlStream:   controlStream,
		controlReader:   controlReader,
		selectedVersion: version,
	}

	sessions = append(sessions, &sessionCore)

	session := SubscribingSession{
		sessionCore: sessionCore,
	}

	return &session, nil
}

func (n node) EstablishPubSubSession(URL string, maxSubscribeID uint64) (*PubSubSession, error) {
	// Parse the url strings
	u, err := url.Parse(URL)

	// Connect to the server as a publisher and a subscriber
	trSess, err := n.connect(u)
	if err != nil {
		return nil, err
	}

	// Open bidirectional stream to send control messages
	controlStream, err := trSess.OpenStream()
	if err != nil {
		return nil, err
	}

	// Get reader of the bidirectional stream
	controlReader := quicvarint.NewReader(controlStream)

	// Set up
	// Initialize SETUP_CLIENT message
	csm := moqtmessage.ClientSetupMessage{
		Versions:   n.SupportedVersions,
		Parameters: make(moqtmessage.Parameters),
	}

	// Add role parameter
	csm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB_SUB)

	// Add max subscribe id parameter
	csm.Parameters.AddParameter(moqtmessage.MAX_SUBSCRIBE_ID, maxSubscribeID)

	// Add path parameter if use raw quic connection
	if u.Scheme == "moqt" {
		csm.Parameters.AddParameter(moqtmessage.PATH, u.Path)
	}

	_, err = controlStream.Write(csm.Serialize())
	if err != nil {
		return nil, err
	}

	version, _, err := n.receiveServerSetupMessage(controlReader) // TODO: Handle the parameter
	if err != nil {
		return nil, err
	}

	sessionCore := sessionCore{
		trSess:          trSess,
		controlStream:   controlStream,
		controlReader:   controlReader,
		selectedVersion: version,
	}

	sessions = append(sessions, &sessionCore)

	session := PubSubSession{
		sessionCore: sessionCore,
	}

	return &session, nil
}

func (n node) connect(url *url.URL) (TransportSession, error) {
	// Set tls configuration
	if n.TLSConfig == nil {
		panic("no TLS configuration")
	}

	switch url.Scheme {
	case "moqt":
		/*
		 * Raw QUIC
		 */

		// Dial to the server
		conn, err := quic.DialAddr(context.TODO(), url.Host, n.TLSConfig, n.QUICConfig)
		if err != nil {
			return nil, err
		}

		return &rawQuicConnectionWrapper{innerSession: conn}, nil
	case "https":
		/*
		 * WebTransport
		 */
		// Initialize webtransport.Dialer
		d := webtransport.Dialer{ // TODO: Configure the Dialer
			TLSClientConfig: n.TLSConfig,
			QUICConfig:      n.QUICConfig,
		}

		// Set header //TODO: Handle header
		var headers http.Header

		// Dial to the server
		_, sess, err := d.Dial(context.TODO(), url.String(), headers)
		if err != nil {
			return nil, err
		}

		// Register the connection
		return &webtransportSessionWrapper{innerSession: sess}, nil
	default:
		return nil, errors.New("invalid URL scheme")
	}
}

func (n node) receiveServerSetupMessage(controlReader quicvarint.Reader) (moqtversion.Version, moqtmessage.Parameters, error) {
	// Receive SETUP_SERVER message
	id, err := moqtmessage.DeserializeMessageID(controlReader)
	if err != nil {
		return 0, nil, err
	}
	if id != moqtmessage.SERVER_SETUP {
		return 0, nil, ErrProtocolViolation
	}

	var ssm moqtmessage.ServerSetupMessage
	err = ssm.DeserializePayload(controlReader)
	if err != nil {
		return 0, nil, err
	}

	// Verify if the selected version is one of the specified versions
	err = moqtversion.Contain(ssm.SelectedVersion, n.SupportedVersions)
	if err != nil {
		return 0, nil, err
	}

	return ssm.SelectedVersion, ssm.Parameters, nil
}
