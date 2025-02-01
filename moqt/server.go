package moqt

import (
	"context"
	"crypto/tls"
	"log"
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

	//
	// SetupHijackerFunc func(SetupRequest) SetupResponce
	// TODO:

	//elayManager *RelayManager

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
				log.Printf("upgrading failed: %s", err)
				w.WriteHeader(500)
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
				Parameters:      message.Parameters(s.Config.SetupExtension.paramMap),
			}

			_, err = ssm.Encode(stream.Stream)
			if err != nil {
				slog.Error("failed to send a set-up responce")
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
					path, err := Parameters{scm.Parameters}.GetString(path)
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
						Parameters:      message.Parameters(s.Config.SetupExtension.paramMap),
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

// func readSetupRequest(r io.Reader) (req SetupRequest, err error) {
// 	slog.Debug("reading a set-up request")
// 	/*
// 	 * Receive a SESSION_CLIENT message
// 	 */
// 	// Decode
// 	var scm message.SessionClientMessage
// 	_, err = scm.Decode(r)
// 	if err != nil {
// 		slog.Error("failed to read a SESSION_CLIENT message", slog.String("error", err.Error())) // TODO
// 		return
// 	}

// 	// Set versions
// 	for _, v := range scm.SupportedVersions {
// 		req.supportedVersions = append(req.supportedVersions, Version(v))
// 	}
// 	// Set parameters
// 	req.SetupParameters = Parameters{scm.Parameters}

// 	slog.Debug("read a set-up request", slog.Any("request", req))

// 	return req, nil
// }

// func writeSetupResponce(w io.Writer, rsp SetupResponce) error {
// 	slog.Debug("writing a set-up responce", slog.Any("responce", rsp))

// 	ssm := message.SessionServerMessage{
// 		SelectedVersion: protocol.Version(rsp.SelectedVersion),
// 		Parameters:      message.Parameters(rsp.Parameters.paramMap),
// 	}

// 	_, err := ssm.Encode(w)
// 	if err != nil {
// 		slog.Error("failed to encode a SESSION_SERVER message", slog.String("error", err.Error()))
// 		return err
// 	}

// 	slog.Debug("wrote a set-up responce")

// 	return nil
// }
