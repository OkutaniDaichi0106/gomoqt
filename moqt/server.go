package moqt

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Server struct {
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
	SetupHandler SetupHandler

	//
	setupRR SetupRequestReader
	setupRW SetupResponceWriter

	// for QUIC
	quicRelayers map[string]Relayer

	wts *webtransport.Server
}

type Relayer struct {
	Path string

	//
	Publisher Publisher

	Subscriber Subscriber
}

func (r Relayer) run(sess Session) {
	go r.Publisher.run(sess)
	go r.Subscriber.run(sess)
}

func (s Server) ListenAndServe() error {
	/*
	 * Listen and serve on raw QUIC
	 */
	ln, err := quic.ListenAddrEarly(s.Addr, s.TLSConfig, s.QUICConfig)
	if err != nil {
		return err
	}

	go func() {
		for {
			qconn, err := ln.Accept(context.Background()) // TODO:
			if err != nil {
				slog.Error("failed to accept a connection", slog.String("error", err.Error()))
				return
			}

			conn := NewMORQConnection(qconn)

			/***/
			go func() {
				/*
				 * Set up
				 */
				// Accept a Stream which must be a Sesson Stream
				stream, err := conn.AcceptStream(context.Background())
				if err != nil {
					slog.Error("failed to open a stream", slog.String("error", err.Error()))
					return
				}

				// Read the first byte and get Stream Type
				buf := make([]byte, 1)
				_, err = stream.Read(buf)
				if err != nil {
					slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
					return
				}

				// Verify if the Stream Type is the SESSION
				if protocol.StreamType(buf[0]) != protocol.SESSION {
					slog.Error("unexpected Stream Type ID", slog.Uint64("ID", uint64(buf[0]))) // TODO
					return
				}

				// Initialize a Session
				sess := Session{
					Connection: conn,
					stream:     stream,
				}

				// Get a set-up request
				req, err := s.setupRR.Read(quicvarint.NewReader(stream))
				if err != nil {
					slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
					return
				}

				// Handle the set-up request
				terr := s.SetupHandler.HandleSetup(req, s.setupRW.New(stream))
				if terr != nil {
					slog.Error("failed to set up", slog.String("error", terr.Error()))

					err := conn.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
					if err != nil {
						slog.Error("failed to close the conncetion", slog.String("error", err.Error()))
					}

					slog.Info("close the connection")

					return
				}

				// Get a handler from the path
				relayer, ok := s.quicRelayers[req.Path]
				if !ok {
					slog.Error("relayer not found", slog.String("path", req.Path))
					return
				}

				go relayer.run(sess)
			}()
		}
	}()

	/*
	 * Listen and serve on Webtransport
	 */
	err = s.wts.ListenAndServe()
	if err != nil {
		slog.Error("failed run the webtransport server", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s Server) RunOnQUIC(relayer Relayer) {
	if _, ok := s.quicRelayers[relayer.Path]; ok {
		panic("relayer overwrite")
	}

	s.quicRelayers[relayer.Path] = relayer
}

func (s Server) RunOnWT(relayer Relayer) {
	if s.wts == nil {
		s.wts = &webtransport.Server{
			H3: http3.Server{
				Addr:       "",
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
		 * Set up
		 */
		// Accept a Stream which must be a Sesson Stream
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to open a stream", slog.String("error", err.Error()))
			return
		}

		//
		sess := Session{
			Connection: conn,
			stream:     stream,
		}

		// Read the first byte and get Stream Type
		buf := make([]byte, 1)
		_, err = stream.Read(buf)
		if err != nil {
			slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
			return
		}
		// Verify if the Stream Type is the SESSION
		if protocol.StreamType(buf[0]) != protocol.SESSION {
			slog.Error("unexpected Stream Type ID", slog.Uint64("ID", uint64(buf[0]))) // TODO
			return
		}

		// Get a set-up request
		req, err := s.setupRR.Read(quicvarint.NewReader(stream))
		if err != nil {
			slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
			return
		}

		// Handle the set-up request
		terr := s.SetupHandler.HandleSetup(req, s.setupRW.New(stream))
		if terr != nil {
			slog.Error("failed to set up", slog.String("error", terr.Error()))

			err := conn.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
			if err != nil {
				slog.Error("failed to close the conncetion", slog.String("error", err.Error()))
			}

			slog.Info("close the connection")

			return
		}

		go relayer.run(sess)
	})
}

// func handleSessionStream(SessionStream)    {}
// func handleAnnounceStream(AnnounceStream)  {}
// func handleSSubscribeStream(SessionStream) {}
// func handleFetchStream(SessionStream)      {}

// func isValidPath(pattern string) bool {
// 	// Verify the pattern starts with "/"
// 	if !strings.HasPrefix(pattern, "/") {
// 		return false
// 	}

// 	_, err := url.ParseRequestURI(pattern)

// 	return err == nil
// }

// func (s Server) SetupMORQ(qconn quic.Connection) (*Session, string, error) {
// 	conn := newMORQConnection(qconn)

// 	// Setup
// 	sess, path, err := s.setupMORQ(conn)

// 	// Terminate the connection when Terminate Error occured
// 	if err != nil {
// 		if terr, ok := err.(TerminateError); ok {
// 			qconn.CloseWithError(quic.ApplicationErrorCode(terr.TerminateErrorCode()), terr.Error())
// 		}
// 		return nil, "", err
// 	}

// 	return sess, path, nil
// }

// func (s Server) setupMORQ(conn Connection) (*Session, string, error) {
// 	/*
// 	 * Accept a bidirectional stream
// 	 */
// 	stream, err := conn.AcceptStream(context.Background())
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	err = acceptSetupStream(stream)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	/*
// 	 * Receive a CLIENT_SETUP message
// 	 */
// 	qvReader := quicvarint.NewReader(stream)
// 	id, preader, err := moqtmessage.ReadControlMessage(qvReader)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	if id != moqtmessage.CLIENT_SETUP {
// 		return nil, "", ErrProtocolViolation
// 	}
// 	var csm moqtmessage.ClientSetupMessage
// 	err = csm.DeserializePayload(preader)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	// Verify if a ROLE parameter exists
// 	role, ok := csm.Parameters.Role()
// 	if !ok {
// 		return nil, "", ErrProtocolViolation
// 	} else if role != moqtmessage.PUB && role != moqtmessage.SUB && role != moqtmessage.PUB_SUB {
// 		return nil, "", ErrProtocolViolation
// 	}
// 	// Get a MAX_SUBSCRIBE_ID parameter
// 	maxID, ok := csm.Parameters.MaxSubscribeID()
// 	if !ok {
// 		maxID = 0
// 	}
// 	// Get a PATH parameter when using raw QUIC
// 	var path string
// 	if _, ok := conn.(*rawQuicConnection); ok {
// 		path, ok = csm.Parameters.Path()
// 		if !ok {
// 			return nil, "", ErrProtocolViolation
// 		}
// 	}

// 	// Handle Parameters in a SERVER_SETUP message
// 	ssparams := make(moqtmessage.Parameters)
// 	if s.SetupHijacker != nil {
// 		ssparams, err = s.SetupHijacker(csm.Parameters)
// 		if err != nil {
// 			return nil, "", err
// 		}
// 	}

// 	/*
// 	 * Select the latest version supported by both the client and the server
// 	 */
// 	selectedVersion, err := protocol.SelectLatestVersion(getProtocolVersions(s.SupportedVersions), csm.SupportedVersions)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	/*
// 	 * Send a SERVER_SETUP message
// 	 */
// 	// Initialize a SERVER_SETUP message
// 	ssm := moqtmessage.ServerSetupMessage{
// 		SelectedVersion: selectedVersion,
// 		Parameters:      make(moqtmessage.Parameters),
// 	}
// 	// ROLE Parameter
// 	switch role {
// 	case moqtmessage.PUB:
// 		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)
// 	case moqtmessage.SUB:
// 		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)
// 	case moqtmessage.PUB_SUB:
// 		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB_SUB)
// 	}
// 	// Optional Parameters
// 	for k, v := range ssparams {
// 		ssm.Parameters.AddParameter(k, v)
// 	}
// 	// Send the message
// 	_, err = stream.Write(ssm.Serialize())
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	return &Session{
// 		Connection:       conn,
// 		selectedVersion:  selectedVersion,
// 		trackAliasMap:    new(trackAliasMap),
// 		subscribeCounter: 0,
// 		maxSubscribeID:   &maxID,
// 	}, path, nil
// }

// func (s Server) SetupMOWT(wtconn *webtransport.Session) (*Session, error) {
// 	conn := newMOWTConnection(wtconn)
// 	sess, err := s.setupMOWT(conn)
// 	if err != nil {
// 		// Terminate the connection if the error is a Terminate Error
// 		if terr, ok := err.(TerminateError); ok {
// 			conn.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
// 		}
// 	}

// 	return sess, nil
// }
// func (s Server) setupMOWT(conn Connection) (*Session, error) {

// 	/*
// 	 * Accept a bidirectional stream
// 	 */
// 	stream, err := conn.AcceptStream(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = acceptSetupStream(stream)
// 	if err != nil {
// 		return nil, err
// 	}

// 	/*
// 	 * Receive a CLIENT_SETUP message
// 	 */
// 	qvReader := quicvarint.NewReader(stream)
// 	id, preader, err := moqtmessage.ReadControlMessage(qvReader)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if id != moqtmessage.CLIENT_SETUP {
// 		return nil, ErrProtocolViolation
// 	}
// 	var csm moqtmessage.ClientSetupMessage
// 	err = csm.DeserializePayload(preader)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Verify if a ROLE parameter exists
// 	role, ok := csm.Parameters.Role()
// 	if !ok {
// 		return nil, ErrProtocolViolation
// 	} else if role != moqtmessage.PUB && role != moqtmessage.SUB && role != moqtmessage.PUB_SUB {
// 		return nil, ErrProtocolViolation
// 	}
// 	// Get a MAX_SUBSCRIBE_ID parameter
// 	maxID, ok := csm.Parameters.MaxSubscribeID()
// 	if !ok {
// 		maxID = 0
// 	}
// 	// Get a PATH parameter and close the connection
// 	if _, ok := conn.(*rawQuicConnection); ok {
// 		_, ok = csm.Parameters.Path()
// 		if ok {
// 			return nil, ErrProtocolViolation
// 		}
// 	}

// 	// Handle Parameters in a SERVER_SETUP message
// 	ssparams := make(moqtmessage.Parameters)
// 	if s.SetupHijacker != nil {
// 		ssparams, err = s.SetupHijacker(csm.Parameters)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	/*
// 	 * Select the latest version supported by both the client and the server
// 	 */
// 	selectedVersion, err := protocol.SelectLatestVersion(getProtocolVersions(s.SupportedVersions), csm.SupportedVersions)
// 	if err != nil {
// 		return nil, err
// 	}

// 	/*
// 	 * Send a SERVER_SETUP message
// 	 */
// 	// Initialize a SERVER_SETUP message
// 	ssm := moqtmessage.ServerSetupMessage{
// 		SelectedVersion: selectedVersion,
// 		Parameters:      make(moqtmessage.Parameters),
// 	}
// 	// ROLE Parameter
// 	switch role {
// 	case moqtmessage.PUB:
// 		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)
// 	case moqtmessage.SUB:
// 		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)
// 	case moqtmessage.PUB_SUB:
// 		ssm.Parameters.AddParameter(moqtmessage.ROLE, moqtmessage.PUB_SUB)
// 	}
// 	// Optional Parameters
// 	for k, v := range ssparams {
// 		ssm.Parameters.AddParameter(k, v)
// 	}
// 	// Send the message
// 	_, err = stream.Write(ssm.Serialize())
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Session{
// 		Connection:       conn,
// 		selectedVersion:  selectedVersion,
// 		trackAliasMap:    new(trackAliasMap),
// 		subscribeCounter: 0,
// 		maxSubscribeID:   &maxID,
// 	}, nil
// }

// func acceptSetupStream(stream Stream) error {
// 	/*
// 	 *
// 	 */
// 	// Read the Stream Type
// 	qvReader := quicvarint.NewReader(stream)
// 	num, err := qvReader.ReadByte()
// 	if err != nil {
// 		return err
// 	}
// 	// verify the Stream Type ID
// 	if StreamType(num) != protocol.SESSION {
// 		log.Println(stream.Close())
// 		return ErrUnexpectedStreamType
// 	}

// 	return nil
// }

func (s *Server) SetCertFiles(certFile, keyFile string) error {
	var err error
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	s.TLSConfig = &tls.Config{
		Certificates: certs,
	}

	return nil
}
