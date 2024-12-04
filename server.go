package moqt

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
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

	//
	SetupHijackerFunc func(SetupRequest) SetupResponce
	//SetupHandler

	// Relayers running on QUIC
	quicRelayers map[string]Relayer

	wts   *webtransport.Server
	wtsMu sync.Mutex
}

func (s *Server) ListenAndServe() error {
	s.wtsMu.Lock()
	defer s.wtsMu.Unlock()

	ln, err := quic.ListenAddrEarly(s.Addr, s.TLSConfig, s.QUICConfig)
	if err != nil {
		slog.Error("failed to listen address", slog.String("error", err.Error()))
		return err
	}

	for {
		qconn, err := ln.Accept(context.Background())
		if err != nil {
			slog.Error("failed to accept", slog.String("error", err.Error()))
			return err
		}

		switch qconn.ConnectionState().TLS.NegotiatedProtocol {
		case "h3":
			/*
			 * Listen and serve on Webtransport
			 */
			if s.wts == nil {
				continue
			}

			go s.wts.ServeQUICConn(qconn)
		case "moq-00":
			/*
			 * Listen and serve on raw QUIC
			 */
			conn := moq.NewMORQConnection(qconn)

			/***/
			go func() {
				/*
				 * Set up
				 */
				// Accept a Bidirectional Stream, which must be a Sesson Stream
				stream, err := conn.AcceptStream(context.Background())
				if err != nil {
					slog.Error("failed to accept a stream", slog.String("error", err.Error()))
					return
				}

				// Get a Stream Type message
				var stm message.StreamTypeMessage
				err = stm.Decode(stream)
				if err != nil {
					slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
					return
				}

				// Verify if the Stream is the Session Stream
				if stm.StreamType != stream_type_session {
					slog.Error("unexpected Stream Type ID", slog.Any("ID", stm.StreamType))
					return
				}

				// Initialize a Session
				sess := ServerSession{
					session: &session{
						conn:                  conn,
						stream:                stream,
						subscribeWriters:      make(map[SubscribeID]*SubscribeWriter),
						receivedSubscriptions: make(map[string]Subscription),
					},
				}

				// Receive a set-up request
				req, err := readSetupRequest(sess.stream)
				if err != nil {
					slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
					return
				}

				// Verify if the request contains a valid path
				if req.Path == "" {
					slog.Error("path not found")
					return
				}

				// Select the default version
				if !ContainVersion(Default, req.SupportedVersions) {
					slog.Error("no available version", slog.Any("versions", req.SupportedVersions))
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
				err = sendSetupResponce(sess.stream, rsp)
				if err != nil {
					slog.Error("failed to send a set-up responce")
					return
				}

				relayer, ok := s.quicRelayers[req.Path]
				if !ok {
					slog.Error("relayer not found", slog.String("path", req.Path))
					return
				}

				relayer.run(&sess)
			}()
		default:
			continue
		}
	}

}

func (s *Server) RunOnQUIC(relayer Relayer) {
	if s.quicRelayers == nil {
		s.quicRelayers = make(map[string]Relayer)
	}

	if _, ok := s.quicRelayers[relayer.Path]; ok {
		panic("relayer overwrite")
	}

	s.quicRelayers[relayer.Path] = relayer
}

func (s *Server) RunOnWebTransport(relayer Relayer) {
	s.wtsMu.Lock()
	defer s.wtsMu.Unlock()

	if s.wts == nil {
		s.wts = &webtransport.Server{
			H3: http3.Server{
				Addr:       s.Addr,
				TLSConfig:  s.TLSConfig,
				QUICConfig: s.QUICConfig,
			},
		}

	}

	http.HandleFunc(relayer.Path, func(w http.ResponseWriter, r *http.Request) {
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
		conn := moq.NewMOWTConnection(wtsess)

		/*
		 * Get a Session
		 */
		var sess ServerSession
		err = sess.init(conn)
		if err != nil {
			slog.Error("failed to initialize a Server Session", slog.String("error", err.Error()))
			return
		}

		// Receive a set-up request
		req, err := readSetupRequest(sess.stream)
		if err != nil {
			slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
			return
		}

		// Select the default version
		if !ContainVersion(Default, req.SupportedVersions) {
			slog.Error("no available version", slog.Any("versions", req.SupportedVersions))
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
		err = sendSetupResponce(sess.stream, rsp)
		if err != nil {
			slog.Error("failed to send a set-up responce")
			return
		}

		/*
		 * Set up
		 */

		relayer.run(&sess)
	})
}

func readSetupRequest(r io.Reader) (SetupRequest, error) {
	/*
	 * Receive a SESSION_CLIENT message
	 */

	// Decode
	var scm message.SessionClientMessage
	err := scm.Decode(r)
	if err != nil {
		slog.Error("failed to read a SESSION_CLIENT message", slog.String("error", err.Error())) // TODO
		return SetupRequest{}, err
	}

	var req SetupRequest

	// Set versions
	for _, v := range scm.SupportedVersions {
		req.SupportedVersions = append(req.SupportedVersions, Version(v))
	}
	// Set parameters
	req.Parameters = Parameters(scm.Parameters)

	// Get any PATH parameter
	path, ok := getPath(req.Parameters)
	if ok {
		req.Path = path
	}

	// Get any MAX_SUBSCRIBE_ID parameter
	maxID, ok := getMaxSubscribeID(req.Parameters)
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
