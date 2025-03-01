package moqt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

type Server struct {
	/*
	 * Server's Address
	 */
	Addr string

	/*
	 * TLS configuration
	 */
	TLSConfig *tls.Config

	/*
	 * QUIC configuration
	 */
	QUICConfig *quic.Config

	Config *Config

	/*
	 *
	 */
	ServeMux *ServeMux

	/*
	 * Session Handler
	 * This function is called when a session is established
	 */
	SessionHandlerFunc func(path string, sess Session)

	// Webtransport Server
	wts *webtransport.Server

	// QUIC Listener
	mutex     sync.RWMutex
	listeners map[*quic.EarlyListener]int
}

func (s *Server) ServeQUICListener(lis quic.Listener) error {

	/*
	 * Listen and serve on raw QUIC
	 */
	for {
		qconn, err := lis.Accept(context.Background())
		if err != nil {
			slog.Error("failed to accept", slog.String("error", err.Error()))
			return err
		}

		s.ServeQUIC(qconn)
	}
}

func (s *Server) ServeQUICConn(qconn quic.Connection) error {
	/*
	 * Listen and serve on raw QUIC
	 */
	conn := transport.NewMORQConnection(qconn)

	/*
	 * Set up
	 */
	stream, scm, err := internal.AcceptSessionStream(conn, context.Background())
	if err != nil {
		slog.Error("failed to accept a session stream", slog.String("error", err.Error()))
		return err
	}

	// Verify if the request contains a valid path
	path, err := Parameters{scm.Parameters}.GetString(param_type_path)
	if err != nil {
		slog.Error("failed to get a path", slog.String("error", err.Error()))
		return err
	}

	if path == "" {
		slog.Error("path not found")
		return err
	}

	// Send a set-up responce
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Version(internal.DefaultServerVersion),
	}

	if s.Config != nil {
		if s.Config.SetupExtensions.paramMap != nil {
			ssm.Parameters = message.Parameters(s.Config.SetupExtensions.paramMap)
		}
	}

	_, err = ssm.Encode(stream.Stream)
	if err != nil {
		slog.Error("failed to send a set-up responce", slog.String("error", err.Error()))
		return err
	}

	sess := &session{internalSession: internal.NewSession(conn, stream), extensions: Parameters{scm.Parameters}}

	s.SessionHandlerFunc(path, sess)

	return nil
}

func (s *Server) ServeWebTransport(w http.ResponseWriter, r *http.Request) error {
	wtsess, err := s.wts.Upgrade(w, r)
	if err != nil {
		slog.Error("failed to upgrade WebTransport session", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// Get a Connection
	conn := transport.NewMOWTConnection(wtsess)

	/*
	 * Set up
	 */
	stream, scm, err := internal.AcceptSessionStream(conn, context.Background())
	if err != nil {
		slog.Error("failed to accept an session stream", slog.String("error", err.Error()))
		return err
	}

	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Version(internal.DefaultServerVersion),
	}

	if s.Config != nil {
		if s.Config.SetupExtensions.paramMap != nil {
			ssm.Parameters = message.Parameters(s.Config.SetupExtensions.paramMap)
		}
	}

	_, err = ssm.Encode(stream.Stream)

	if err != nil {
		slog.Error("failed to send a set-up response", slog.String("error", err.Error()))
		return err
	}

	// Initialize a Session
	sess := &session{internalSession: internal.NewSession(conn, stream), extensions: Parameters{scm.Parameters}}

	s.SessionHandlerFunc(r.URL.Path, sess)

	return nil
}

func (s *Server) init() (err error) {
	/*
	 * Raw QUIC
	 */
	s.quicListener, err = quic.ListenAddrEarly(s.Addr, s.TLSConfig, s.QUICConfig)
	if err != nil {
		slog.Error("failed to listen address", slog.String("error", err.Error()))
		return err
	}

	if s.ServeMux == nil {
		s.ServeMux = DefaultMux
	}

	return nil
}

func (s *Server) ListenAndServe() error {
	err := s.init()
	if err != nil {
		slog.Error("failed to initialize a server")
		return err
	}

	/***/
	for {
		qconn, err := s.quicListener.Accept(context.Background())
		if err != nil {
			slog.Error("failed to accept", slog.String("error", err.Error()))
			return err
		}

		switch qconn.ConnectionState().TLS.NegotiatedProtocol {
		case http3.NextProtoH3:
			/*
			 * Listen and serve on webtransport
			 */
			// If the server is not initialized, initialize it
			if s.wts == nil {
				s.wts = &webtransport.Server{
					H3: http3.Server{
						Addr:       s.Addr,
						TLSConfig:  s.TLSConfig,
						QUICConfig: s.QUICConfig,
					},
				}
			}

			/*
			 * Serve the QUIC Connection
			 */
			go s.wts.ServeQUICConn(qconn)
		case NextProtoMOQ:
			/*
			 * Serve the QUIC Connection
			 */

			go s.ServeQUIC(qconn)
		default:
			slog.Error("unsupported protocol", slog.String("protocol", qconn.ConnectionState().TLS.NegotiatedProtocol))
			return fmt.Errorf("unsupported protocol: %s", qconn.ConnectionState().TLS.NegotiatedProtocol)
		}

	}

}

const NextProtoMOQ = "moq-00"
