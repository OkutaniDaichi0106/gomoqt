package moqt

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
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

	/*
	 *
	 */
	ServeMux *ServeMux

	//
	SetupHijackerFunc func(SetupRequest) SetupResponce
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

	s.ServeMux.mu.Lock()
	defer s.ServeMux.mu.Unlock()
	for path, op := range s.ServeMux.handlerFuncs {
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
			stream, err := acceptSessionStream(conn)
			if err != nil {
				slog.Error("failed to accept an session stream", slog.String("error", err.Error()))
				return
			}

			// Receive a set-up request
			req, err := readSetupRequest(stream)
			if err != nil {
				slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
				return
			}

			// Verify if the request contains a valid path
			if req.parsedURL.Path == "" {
				slog.Error("path not found")
				return
			}

			// Select the default version
			if !ContainVersion(Default, req.supportedVersions) {
				slog.Error("no available version", slog.Any("versions", req.supportedVersions))
				return
			}

			// Send a set-up responce
			var rsp SetupResponce
			if s.SetupHijackerFunc != nil {
				rsp = s.SetupHijackerFunc(req)
			} else {
				rsp = SetupResponce{
					SelectedVersion: Default,
					Parameters:      make(Parameters),
				}
			}
			err = sendSetupResponce(stream, rsp)
			if err != nil {
				slog.Error("failed to send a set-up responce")
				return
			}

			// Initialize a Session
			sess := &serverSession{
				session: session{
					conn:   conn,
					stream: stream,
				},
			}

			op(sess)
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
					stream, err := acceptSessionStream(conn)
					if err != nil {
						slog.Error("failed to accept an session stream", slog.String("error", err.Error()))
						return
					}

					// Receive a set-up request
					req, err := readSetupRequest(stream)
					if err != nil {
						slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
						return
					}

					// Verify if the request contains a valid path
					if req.parsedURL.Path == "" {
						slog.Error("path not found")
						return
					}

					// Select the default version
					if !ContainVersion(Default, req.supportedVersions) {
						slog.Error("no available version", slog.Any("versions", req.supportedVersions))
						return
					}

					// Send a set-up responce
					var rsp SetupResponce
					if s.SetupHijackerFunc != nil {
						rsp = s.SetupHijackerFunc(req)
					} else {
						rsp = SetupResponce{
							SelectedVersion: Default,
							Parameters:      make(Parameters),
						}
					}
					err = sendSetupResponce(stream, rsp)
					if err != nil {
						slog.Error("failed to send a set-up responce")
						return
					}

					// Initialize a Session
					sess := &serverSession{
						session: session{
							conn:   conn,
							stream: stream,
						},
					}

					handler := s.ServeMux.findHandlerFunc(req.parsedURL.Path)

					handler(sess)
				}(qconn)
			default:
				return
			}
		}(qconn)

	}

}

func acceptSessionStream(conn transport.Connection) (transport.Stream, error) {
	// Accept a Bidirectional Stream, which must be a Sesson Stream
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		slog.Error("failed to accept a stream", slog.String("error", err.Error()))
		return nil, err
	}

	// Get a Stream Type message
	var stm message.StreamTypeMessage
	err = stm.Decode(stream)
	if err != nil {
		slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
		return nil, err
	}

	// Verify if the Stream is the Session Stream
	if stm.StreamType != stream_type_session {
		slog.Error("unexpected Stream Type ID", slog.Any("ID", stm.StreamType))
		return nil, err
	}

	return stream, nil
}

func readSetupRequest(r io.Reader) (req SetupRequest, err error) {
	/*
	 * Receive a SESSION_CLIENT message
	 */

	// Decode
	var scm message.SessionClientMessage
	err = scm.Decode(r)
	if err != nil {
		slog.Error("failed to read a SESSION_CLIENT message", slog.String("error", err.Error())) // TODO
		return
	}

	// Set versions
	for _, v := range scm.SupportedVersions {
		req.supportedVersions = append(req.supportedVersions, Version(v))
	}
	// Set parameters
	req.SetupParameters = Parameters(scm.Parameters)

	// Get any PATH parameter
	path, ok := getPath(req.SetupParameters)
	if ok {
		req.parsedURL.Path = path
	}

	// Get any MAX_SUBSCRIBE_ID parameter
	maxID, ok := getMaxSubscribeID(req.SetupParameters)
	if ok {
		req.MaxSubscribeID = uint64(maxID)
	}

	return req, nil
}

func sendSetupResponce(w io.Writer, rsp SetupResponce) error {
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Version(rsp.SelectedVersion),
		Parameters:      message.Parameters(rsp.Parameters),
	}

	err := ssm.Encode(w)
	if err != nil {
		slog.Error("failed to encode a SESSION_SERVER message", slog.String("error", err.Error()))
		return err
	}

	return nil
}
