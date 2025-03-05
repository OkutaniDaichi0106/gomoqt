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

	/*
	 * MOQ Configuration
	 */
	Config *Config

	/*
	 * Setup Extensions
	 * This function is called when a session is established
	 */
	SetupExtensions func(req Parameters) (rsp Parameters, err error)

	/*
	 * TrackMux for routing requests
	 */
	TrackMux *TrackMux

	/*
	 * Session Handler
	 * This function is called when a session is established
	 */
	SessionHandlerFunc func(path string, sess Session)

	/*
	 * Logger
	 */
	Logger slog.Logger

	/*
	 * WebTransport Server
	 * If the server is configured with a WebTransport server, it is used to handle WebTransport sessions.
	 * If not, a default server is used.
	 */
	WebtransportServer *webtransport.Server

	lnMu            sync.RWMutex
	rawQUICListners map[*QUICEarlyListener]struct{}

	closed bool
}

func (s *Server) ServeQUICListener(ln QUICEarlyListener) error {
	for {
		qconn, err := ln.Accept(context.Background())
		if err != nil {
			s.Logger.Error("failed to accept", "error", err.Error())
			return err
		}

		return s.ServeQUICConn(qconn)
	}
}

func (s *Server) ServeQUICConn(qconn quic.Connection) error {
	// Debug: Log negotiated protocol.
	s.Logger.Debug("Negotiated protocol", "protocol", qconn.ConnectionState().TLS.NegotiatedProtocol)

	switch qconn.ConnectionState().TLS.NegotiatedProtocol {
	case http3.NextProtoH3:
		slog.Debug("serving MOQ over WebTransport session")

		if s.WebtransportServer == nil {
			s.setDefaultWebtransportServer()
		}

		return s.WebtransportServer.ServeQUICConn(qconn)
	case NextProtoMOQ:
		s.Logger.Debug("serving MOQ over QUIC connection")

		conn := transport.NewMORQConnection(qconn)

		var path string
		handler := func(reqParam message.Parameters) (message.Parameters, error) {
			var err error
			param := Parameters{reqParam}
			path, err = param.GetString(param_type_path)
			if err != nil {
				s.Logger.Error("failed to get a path", "error", err.Error())
				return nil, err
			}

			rspParam, err := s.SetupExtensions(param)
			if err != nil {
				s.Logger.Error("failed to set up", "error", err.Error())
				return nil, err
			}

			return rspParam.paramMap, nil
		}

		internalSess, err := internal.AcceptSession(context.Background(), conn, handler)
		if err != nil {
			s.Logger.Error("failed to open a session stream", "error", err.Error())
			internalSess.Terminate(err)
			return err
		}
		s.Logger.Debug("accepted a session")

		s.Logger.Debug("extracted session path", "path", path)

		if path == "" {
			s.Logger.Error("invalid session path", "path", path)
			return fmt.Errorf("invalid session path")
		}

		sess := &session{internalSession: internalSess}

		s.SessionHandlerFunc(path, sess)

		return nil
	default:
		s.Logger.Error("unsupported protocol", "protocol", qconn.ConnectionState().TLS.NegotiatedProtocol)
		return fmt.Errorf("unsupported protocol: %s", qconn.ConnectionState().TLS.NegotiatedProtocol)
	}
}

// ServeWebTransport serves a WebTransport session.
// It upgrades the HTTP/3 connection to a WebTransport session and calls the session handler.
// If the server is not configured with a WebTransport server, it creates a default server.
func (s *Server) ServeWebTransport(w http.ResponseWriter, r *http.Request) error {
	if s.WebtransportServer == nil {
		s.setDefaultWebtransportServer()
	}

	wtsess, err := s.WebtransportServer.Upgrade(w, r)
	if err != nil {
		s.Logger.Error("failed to upgrade to WebTransport session", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	s.Logger.Debug("Upgraded to WebTransport session", "remoteAddr", r.RemoteAddr)

	// Get a Connection
	conn := transport.NewMOWTConnection(wtsess)
	s.Logger.Debug("WebTransport connection created")

	var rspParam Parameters
	handler := func(reqParam message.Parameters) (message.Parameters, error) {
		var err error
		rspParam, err = s.SetupExtensions(Parameters{reqParam})
		if err != nil {
			s.Logger.Error("failed to set up", "error", err.Error())
			return nil, err
		}
		return rspParam.paramMap, nil
	}

	internalSess, err := internal.AcceptSession(context.Background(), conn, handler)
	if err != nil {
		s.Logger.Error("failed to open a session stream", "error", err.Error())
		internalSess.Terminate(err)
		return err
	}
	s.Logger.Debug("session started")

	// Initialize a Session
	sess := &session{
		internalSession: internalSess,
	}

	s.Logger.Debug("Invoking session handler", "path", r.URL.Path)

	s.SessionHandlerFunc(r.URL.Path, sess)

	return nil
}

func (s *Server) ListenAndServe() error {
	s.Logger.Debug("server listening", "address", s.Addr)
	ln, err := quic.ListenAddrEarly(s.Addr, s.TLSConfig, s.QUICConfig)
	if err != nil {
		s.Logger.Error("failed to listen", "error", err.Error())
		return err
	}

	return s.ServeQUICListener(ln)
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) (err error) {
	s.Logger.Debug("server (TLS) listening", "address", s.Addr)
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		s.Logger.Error("failed to load X509 key pair", "error", err.Error())
		return err
	}
	ln, err := quic.ListenAddrEarly(s.Addr, &tls.Config{Certificates: certs}, s.QUICConfig)
	if err != nil {
		s.Logger.Error("failed to listen", "error", err.Error())
		return err
	}

	return s.ServeQUICListener(ln)
}

func (s *Server) Close() error {
	s.lnMu.Lock()
	defer s.lnMu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	if s.rawQUICListners != nil {
		for ln := range s.rawQUICListners {
			(*ln).Close()
		}
	}

	if s.WebtransportServer != nil {
		s.WebtransportServer.Close()
	}

	return nil
}

func (s *Server) setDefaultWebtransportServer() {
	s.lnMu.Lock()
	defer s.lnMu.Unlock()

	if s.WebtransportServer == nil {
		s.WebtransportServer = &webtransport.Server{
			H3: http3.Server{
				Addr:       s.Addr,
				TLSConfig:  s.TLSConfig.Clone(),
				QUICConfig: s.QUICConfig.Clone(),
			},
		}
		s.Logger.Debug("default WebTransport server initialized", "address", s.Addr)
	}
}

const NextProtoMOQ = "moq-00"
