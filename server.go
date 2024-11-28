package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
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
	 * MOQT Versions supported by the moqt server
	 */
	SupportedVersions []Version

	//
	SetupHandler

	// Relayers running on QUIC
	quicRelayers map[string]Relayer

	wts   *webtransport.Server
	wtsMu sync.Mutex
}

func (s *Server) ListenAndServe() error {
	s.wtsMu.Lock()
	defer s.wtsMu.Unlock()

	listener, err := quic.ListenAddrEarly(s.Addr, s.TLSConfig, s.QUICConfig)
	if err != nil {
		slog.Error("failed to listen address", slog.String("error", err.Error()))
		return err
	}

	for {
		qconn, err := listener.Accept(context.Background())
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
			conn := NewMORQConnection(qconn)

			/***/
			go func() {
				/*
				 * Set up
				 */
				// Accept a Stream, which must be a Sesson Stream
				stream, err := conn.AcceptStream(context.Background())
				if err != nil {
					slog.Error("failed to accept a stream", slog.String("error", err.Error()))
					return
				}

				// Initialize a Session
				sess := ServerSession{
					session: &session{
						conn:                  conn,
						sessStr:               stream,
						subscribeWriters:      make(map[SubscribeID]*SubscribeWriter),
						receivedSubscriptions: map[SubscribeID]Subscription{},
					},
				}

				/*
				 * Set up
				 */
				// Read the first byte and get Stream Type
				buf := make([]byte, 1)
				_, err = sess.sessStr.Read(buf)
				if err != nil {
					slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
					return
				}
				// Verify if the Stream Type is the SESSION
				if StreamType(buf[0]) != SESSION {
					slog.Error("unexpected Stream Type ID", slog.Any("expected ID", SESSION), slog.Any("detected ID", StreamType(buf[0]))) // TODO
					return
				}

				// Get a set-up request
				req, err := getSetupRequest(quicvarint.NewReader(sess.sessStr))
				if err != nil {
					slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
					return
				}
				// Initialize a SetupResponceWriter{}
				srw := SetupResponceWriter{
					doneCh: make(chan struct{}, 1),
					once:   new(sync.Once),
					stream: sess.sessStr,
				}

				// Handle the Set-up
				s.HandleSetup(req, srw)

				<-srw.doneCh

				// Get a relayer from the path
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

		conn := NewMOWTConnection(wtsess)

		/*
		 * Get a Session
		 */
		// Accept a bidirectional Stream for the Sesson Stream
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to open a stream", slog.String("error", err.Error()))
			return
		}

		// Initialize a Session

		sess := ServerSession{
			session: &session{
				conn:                  conn,
				sessStr:               stream,
				subscribeWriters:      make(map[SubscribeID]*SubscribeWriter),
				receivedSubscriptions: make(map[SubscribeID]Subscription),
			},
		}

		/*
		 * Set up
		 */
		// Read the first byte and get Stream Type
		buf := make([]byte, 1)
		_, err = sess.sessStr.Read(buf)
		if err != nil {
			slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
			return
		}
		// Verify if the Stream Type is the SESSION
		if StreamType(buf[0]) != SESSION {
			slog.Error("unexpected Stream Type ID", slog.Uint64("ID", uint64(buf[0]))) // TODO
			return
		}

		// Get a set-up request
		req, err := getSetupRequest(quicvarint.NewReader(sess.sessStr))
		if err != nil {
			slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
			return
		}

		srw := SetupResponceWriter{
			doneCh: make(chan struct{}, 1),
			once:   new(sync.Once),
			stream: sess.sessStr,
		}

		s.HandleSetup(req, srw)

		<-srw.doneCh

		go relayer.run(&sess)
	})
}

func getSetupRequest(r quicvarint.Reader) (SetupRequest, error) {
	// Receive SESSION_CLIENT message
	var scm message.SessionClientMessage
	err := scm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a SESSION_CLIENT message", slog.String("error", err.Error())) // TODO
		return SetupRequest{}, err
	}

	// Get a path
	path, ok := getPath(scm.Parameters)
	if !ok {
		err := errors.New("path not found")
		slog.Error("path not found")
		return SetupRequest{}, err
	}

	req := SetupRequest{
		Path:       path,
		Parameters: Parameters(scm.Parameters),
	}
	// Set versions
	for _, v := range scm.SupportedVersions {
		req.SupportedVersions = append(req.SupportedVersions, Version(v))
	}

	return req, nil
}
