package moqtransport

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtversion"

	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

var pathes map[string]struct{}

var DefaultPublishingHandler map[string]func(*PublishingSession)
var DefaultSubscribingHandler map[string]func(*SubscribingSession)
var DefaultPubSubHandler map[string]func(*PubSubSession)

//var relayMapping map[subscribingSessionID][]PublishingSession

func HandlePublishingFunc(pattern string, op func(*PublishingSession)) {
	_, ok := DefaultPublishingHandler[pattern]
	if ok {
		panic("the path is already in use")
	}

	DefaultPublishingHandler[pattern] = op
}

func HandleSubscribingFunc(pattern string, op func(*SubscribingSession)) {
	_, ok := DefaultSubscribingHandler[pattern]
	if ok {
		panic("the path is already in use")
	}

	DefaultSubscribingHandler[pattern] = op
}

func HandleRelayFunc(pattern string, op func(*PubSubSession)) {
	_, ok := DefaultPubSubHandler[pattern]
	if ok {
		panic("the path is already in use")
	}

	DefaultPubSubHandler[pattern] = op
}

var sessions []*sessionCore

/*
 * Server
 */
type Server struct {
	Addr string

	Port int

	TLSConfig *tls.Config

	QUICConfig *quic.Config

	Versions []moqtversion.Version

	WTConfig struct {
		ReorderingTimeout time.Duration

		CheckOrigin func(r *http.Request) bool

		EnableDatagrams bool
	}

	HijackSetup func(moqtmessage.ClientSetupMessage) (moqtmessage.ServerSetupMessage, error)
}

func (s Server) ListenAndServe() error {
	errCh := make(chan error, 1)

	if s.TLSConfig == nil {
		panic("TLS configuration not found")
	}

	go func(srv Server) {
		err := srv.listenWebTransport()
		if err != nil {
			errCh <- err
		}
	}(s)

	go func(srv Server) {
		err := s.listenRawQUIC()
		if err != nil {
			errCh <- err
		}
	}(s)

	return <-errCh
}

func (s Server) listenWebTransport() error {
	wtServer := webtransport.Server{
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

	for path := range pathes {
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			/*
			 * Establish WebTransport session
			 */
			// Upgrade the HTTP session to a WebTransport session
			sess, err := wtServer.Upgrade(w, r)
			if err != nil {
				log.Printf("upgrading failed: %s", err)
				w.WriteHeader(500)
				return
			}

			err = s.handleWebTransport(path, sess)

			switch e := err.(type) {
			case TerminateError:
				sess.CloseWithError(webtransport.SessionErrorCode(e.Code()), e.Error())
			default:
				sess.CloseWithError(webtransport.SessionErrorCode(TERMINATE_INTERNAL_ERROR), e.Error())
			}
		})
	}

	return wtServer.ListenAndServe()
}

func (s Server) listenRawQUIC() error {
	ln, err := quic.ListenAddrEarly(s.Addr, s.TLSConfig, s.QUICConfig)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := ln.Accept(context.TODO()) // TODO:
			if err != nil {
				log.Println(err)
			}

			go func(conn quic.Connection) {
				err := s.handleRawQUIC(conn)
				switch e := err.(type) {
				case TerminateError:
					conn.CloseWithError(quic.ApplicationErrorCode(e.Code()), e.Error())
				default:
					conn.CloseWithError(quic.ApplicationErrorCode(TERMINATE_INTERNAL_ERROR), e.Error())
				}
			}(conn)
		}
	}()

	return nil
}

func (s Server) handleWebTransport(path string, sess *webtransport.Session) error {
	// Accept bidirectional stream to exchange control messages
	controlStream, err := sess.AcceptStream(context.TODO())
	if err != nil {
		return err
	}

	controlReader := quicvarint.NewReader(controlStream)

	// Receive a SETUP_CLIENT message
	var csm moqtmessage.ClientSetupMessage
	err = csm.DeserializePayload(controlReader)
	if err != nil {
		return err
	}

	// Select the later version of the moqtransport
	version, err := moqtversion.SelectLaterVersion(s.Versions, csm.Versions)

	// Get the ROLE parameter
	role, ok := csm.Parameters.Role()
	if !ok {
		return ErrProtocolViolation
	}

	// Get the MAX_SUBSCRIBE_ID parameter
	maxSubscribeID, ok := csm.Parameters.MaxSubscribeID()

	// Send a SERVER_SETUP message
	var ssm moqtmessage.ServerSetupMessage

	if s.HijackSetup != nil {
		// Get a costimized SERVER_SETUP
		ssm, err = s.HijackSetup(csm)
		if err != nil {
			return err
		}
	} else {
		// Get the default SERVER_SETUP
		ssm = moqtmessage.ServerSetupMessage{
			SelectedVersion: version,
			Parameters:      make(moqtmessage.Parameters),
		}
	}

	_, err = controlStream.Write(ssm.Serialize())
	if err != nil {
		return err
	}

	sessionCore := sessionCore{
		sessionID:       getNextSessionID(),
		trSess:          &webtransportSessionWrapper{innerSession: sess},
		controlStream:   webtransportStreamWrapper{innerStream: controlStream},
		controlReader:   controlReader,
		selectedVersion: version,
	}

	sessions = append(sessions, &sessionCore)

	switch role {
	case moqtmessage.PUB:
		sess := SubscribingSession{
			sessionCore: sessionCore,
		}

		op, ok := DefaultSubscribingHandler[path]
		if !ok {
			return ErrInternalError
		}

		op(&sess)

		return nil
	case moqtmessage.SUB:
		sess := PublishingSession{
			sessionCore:    sessionCore,
			maxSubscribeID: maxSubscribeID,
		}

		op, ok := DefaultPublishingHandler[path]
		if !ok {
			return ErrInternalError
		}

		op(&sess)

		return nil
	case moqtmessage.PUB_SUB:
		sess := PubSubSession{
			sessionCore:    sessionCore,
			maxSubscribeID: maxSubscribeID,
		}

		op, ok := DefaultPubSubHandler[path]
		if !ok {
			return ErrInternalError
		}

		op(&sess)

		return nil
	default:
		return ErrInternalError
	}
}

func (s Server) handleRawQUIC(conn quic.Connection) error {
	controlStream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return err
	}

	controlReader := quicvarint.NewReader(controlStream)

	id, err := moqtmessage.DeserializeMessageID(controlReader)
	if err != nil {
		return err
	}

	// Verify the received message is a CLIENT_SETUP message
	if id != moqtmessage.CLIENT_SETUP {
		return err
	}

	var csm moqtmessage.ClientSetupMessage
	err = csm.DeserializePayload(controlReader)
	if err != nil {
		return err
	}

	// Select the later version of the moqtransport
	version, err := moqtversion.SelectLaterVersion(s.Versions, csm.Versions)

	// Get the Role parameter
	role, ok := csm.Parameters.Role()
	if ok {
		return ErrProtocolViolation
	}

	// Get the MAX_SUBSCRIBE_ID parameter
	maxSubscribeID, ok := csm.Parameters.MaxSubscribeID()

	// Get the Path parameter
	path, ok := csm.Parameters.Path()
	if !ok {
		return ErrProtocolViolation
	}

	var ssm moqtmessage.ServerSetupMessage

	if s.HijackSetup != nil {
		ssm, err = s.HijackSetup(csm)
		if err != nil {
			return err
		}
	}

	_, err = controlStream.Write(ssm.Serialize())
	if err != nil {
		return err
	}

	sessionCore := sessionCore{
		sessionID:       getNextSessionID(),
		trSess:          &rawQuicConnectionWrapper{innerSession: conn},
		controlStream:   rawQuicStreamWrapper{innerStream: controlStream},
		controlReader:   controlReader,
		selectedVersion: version,
	}

	sessions = append(sessions, &sessionCore)

	switch role {
	case moqtmessage.PUB:
		sess := SubscribingSession{
			sessionCore: sessionCore,
		}

		op, ok := DefaultSubscribingHandler[path]
		if !ok {
			return ErrInternalError
		}

		op(&sess)

		return nil
	case moqtmessage.SUB:
		sess := PublishingSession{
			sessionCore:    sessionCore,
			maxSubscribeID: maxSubscribeID,
		}

		op, ok := DefaultPublishingHandler[path]
		if !ok {
			return ErrInternalError
		}

		op(&sess)

		return nil
	case moqtmessage.PUB_SUB:
		sess := PubSubSession{
			sessionCore:    sessionCore,
			maxSubscribeID: maxSubscribeID,
		}

		op, ok := DefaultPubSubHandler[path]
		if !ok {
			return ErrInternalError
		}

		op(&sess)

		return nil
	default:
		return ErrInternalError
	}
}

func (s Server) GoAway(url string, duration time.Duration) {
	gm := moqtmessage.GoAwayMessage{
		NewSessionURI: url,
	}

	for _, session := range sessions {
		go func(session *sessionCore) {
			// Send the GoAway message
			_, err := session.controlStream.Write(gm.Serialize())
			if err != nil {
				log.Println(err)
			}

			time.Sleep(duration)

			session.Terminate(ErrGoAwayTimeout)
		}(session)
	}
}

func isValidPath(pattern string) bool {
	// Verify the pattern starts with "/"
	if !strings.HasPrefix(pattern, "/") {
		return false
	}

	_, err := url.ParseRequestURI(pattern)

	return err == nil
}

var ErrUnsuitableRole = errors.New("the role cannot perform the operation ")
var ErrInvalidRole = errors.New("given role is invalid")
var ErrDuplicatedNamespace = errors.New("given namespace is already registered")
var ErrNoAgent = errors.New("no agent")
