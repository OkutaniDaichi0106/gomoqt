package moqt

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"

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

	// QUIC Listener
	quicListener *quic.EarlyListener

	// Webtransport Server
	wts *webtransport.Server
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

	/*
	 * WebTransport
	 */
	s.wts = &webtransport.Server{
		H3: http3.Server{
			Addr:       s.Addr,
			TLSConfig:  s.TLSConfig,
			QUICConfig: s.QUICConfig,
		},
	}

	if s.ServeMux == nil {
		s.ServeMux = DefaultHandler
	}

	s.ServeMux.mu.Lock()
	defer s.ServeMux.mu.Unlock()
	for path, handler := range s.ServeMux.handlerFuncs {
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			/*
			 *
			 */
			wtsess, err := s.wts.Upgrade(w, r)
			if err != nil {
				slog.Error("upgrading failed", slog.String("error", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Get a Connection
			conn := transport.NewMOWTConnection(wtsess)

			/*
			 * Set up
			 */
			stream, scm, err := internal.AcceptSessionStream(conn, context.Background())
			if err != nil {
				slog.Error("failed to accept an session stream", slog.String("error", err.Error()))
				return
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
				return
			}

			// Initialize a Session
			sess := internal.NewSession(conn, stream)

			handler(&session{internalSession: sess, extensions: Parameters{scm.Parameters}})
		})
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

		go func(qconn quic.Connection) {
			switch qconn.ConnectionState().TLS.NegotiatedProtocol {
			case "h3":
				/*
				 * Listen and serve on Webtransport
				 */
				if s.wts == nil {
					panic("webtranport.Server is nil")
				}

				/*
				 * Serve the QUIC Connection
				 */
				go s.wts.ServeQUICConn(qconn)

				/*
				 * The quic.Connection will be handled in the ServeQUICConn()
				 * and will be served as a webtransport.Session in a http.HandlerFunc()
				 * In the HandlerFunc() the ServerSession will be handled
				 */
			case "moq-00":
				/*
				 * Serve the QUIC Connection
				 */
				go func(qconn quic.Connection) {
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
						return
					}

					// Verify if the request contains a valid path
					path, err := Parameters{scm.Parameters}.GetString(param_type_path)
					if err != nil {
						slog.Error("failed to get a path", slog.String("error", err.Error()))
						return
					}

					if path == "" {
						slog.Error("path not found")
						return
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
						return
					}

					handler := s.ServeMux.findHandlerFunc("/" + path)

					handler(&session{internalSession: internal.NewSession(conn, stream), extensions: Parameters{scm.Parameters}})
				}(qconn)
			default:
				return
			}
		}(qconn)

	}

}
