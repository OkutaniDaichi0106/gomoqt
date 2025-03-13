package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

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
	 * Handler
	 */
	Handler TrackResolver

	/*
	 * Session Handler
	 * This function is called when a session is established
	 */
	SessionHandlerFunc func(path string, sess Session)

	/*
	 * Logger
	 */
	Logger *slog.Logger

	/*
	 * WebTransport Server
	 * If the server is configured with a WebTransport server, it is used to handle WebTransport sessions.
	 * If not, a default server is used.
	 */
	WebtransportServer *webtransport.Server

	lnMu            sync.RWMutex
	rawQUICListners map[*QUICEarlyListener]struct{}

	once   sync.Once
	mu     sync.Mutex
	closed atomic.Bool

	doneChan     chan struct{} // Signal channel (notifies when server is completely closed)
	shutdownChan chan struct{} // Shutdown notification channel
	connCount    atomic.Int64  // Active connection counter
}

func (s *Server) init() {
	s.once.Do(func() {
		if s.Logger == nil {
			s.Logger = slog.Default()
		}
		s.rawQUICListners = make(map[*QUICEarlyListener]struct{})

		// Initialize signal channels
		s.doneChan = make(chan struct{})
		s.shutdownChan = make(chan struct{})

		// TODO: Initialize TrackMux
		// TODO: Initialize SessionHandlerFunc
		// TODO: Initialize SetupExtensions
		if s.SetupExtensions == nil {
			s.setDefaultSetupExtensions()
		}

		// Initialize WebtransportServer
		if s.WebtransportServer == nil {
			s.setDefaultWebtransportServer()
		}

		s.Logger.Debug("initialized server", "address", s.Addr)
	})
}

func (s *Server) serveQUICListener(ln QUICEarlyListener) error {
	s.init()

	s.addListener(&ln)
	defer s.removeListener(&ln)

	// Create context for listener's Accept operation
	// This context will be canceled when the server is shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor shutdown notification and cancel context
	go func() {
		select {
		case <-s.shutdownChan:
			cancel()
		case <-ctx.Done():
			// Context canceled for other reasons
		}
	}()

	s.Logger.Debug("listening for QUIC connections", "listener", ln)

	for {
		// Listen for new QUIC connections
		qconn, err := ln.Accept(ctx)
		if err != nil {
			// Error due to shutdown
			if errors.Is(err, quic.ErrServerClosed) || ctx.Err() != nil || s.closed.Load() {
				return http.ErrServerClosed
			}
			s.Logger.Error("failed to accept new QUIC connection", "listener", ln, "error", err.Error())
			return err
		}

		s.Logger.Info("Accepted new QUIC connection", "remote_address", qconn.RemoteAddr())

		// Increment connection counter
		s.incrementConnCount()

		// Handle connection in a goroutine
		go func() {
			defer s.decrementConnCount()
			if err := s.serveQUICConn(qconn); err != nil {
				s.Logger.Debug("handling connection failed", "error", err)
			}
		}()
	}
}

func (s *Server) serveQUICConn(qconn quic.Connection) error {
	s.init()

	protocol := qconn.ConnectionState().TLS.NegotiatedProtocol

	s.Logger.Info("Negotiated protocol", "remote_address", qconn.RemoteAddr(), "protocol", protocol)

	switch protocol {
	case http3.NextProtoH3:
		s.Logger.Debug("handling webtransport session", "remote_address", qconn.RemoteAddr())
		if s.WebtransportServer == nil {
			s.setDefaultWebtransportServer()
		}
		return s.WebtransportServer.ServeQUICConn(qconn)
	case NextProtoMOQ:
		s.Logger.Debug("handling quic connection", "remote_address", qconn.RemoteAddr())
		conn := transport.NewMORQConnection(qconn)
		var path string
		handler := func(reqParam message.Parameters) (message.Parameters, error) {
			var err error
			param := Parameters{reqParam}

			// Get the path parameter
			path, err = param.GetString(param_type_path)
			if err != nil {
				s.Logger.Error("failed to get 'path' parameter", "remote_address", qconn.RemoteAddr(), "error", err.Error())
				return nil, err
			}

			// Get any setup extensions
			rspParam, err := s.SetupExtensions(param)
			if err != nil {
				s.Logger.Error("failed to get setup extensions", "remote_address", qconn.RemoteAddr(), "error", err.Error())
				return nil, err
			}

			return rspParam.paramMap, nil
		}

		internalSess, err := internal.AcceptSession(context.Background(), conn, handler)
		if err != nil {
			s.Logger.Error("failed to accept session", "remote_address", qconn.RemoteAddr(), "error", err.Error())
			return err
		}

		s.Logger.Info("established moq session over quic", "remote_address", qconn.RemoteAddr())

		if path == "" {
			s.Logger.Error("Invalid session path", "remote_address", qconn.RemoteAddr(), "path", path)
			err := fmt.Errorf("invalid session path")
			// Terminate the session
			internalSess.Terminate(err)
			return err
		}

		sess := newSession(internalSess, s.Handler)

		s.Logger.Debug("handle session", "remote_address", qconn.RemoteAddr(), "path", path)

		s.SessionHandlerFunc(path, sess)

		s.Logger.Debug("completed session handling", "remote_address", qconn.RemoteAddr(), "path", path)

		return nil
	default:
		s.Logger.Error("unsupported negotiated protocol", "remote_address", qconn.RemoteAddr(), "protocol", protocol)
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// ServeWebTransport serves a WebTransport session.
// It upgrades the HTTP/3 connection to a WebTransport session and calls the session handler.
// If the server is not configured with a WebTransport server, it creates a default server.
func (s *Server) ServeWebTransport(w http.ResponseWriter, r *http.Request) error {
	s.init()

	if s.WebtransportServer == nil {
		s.setDefaultWebtransportServer()
	}
	wtsess, err := s.WebtransportServer.Upgrade(w, r)
	if err != nil {
		s.Logger.Error("WebTransport upgrade failed", "remote_address", r.RemoteAddr, "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	s.Logger.Debug("WebTransport session established", "remote_address", r.RemoteAddr)

	// Create a MOQ connection
	conn := transport.NewMOWTConnection(wtsess)

	var rspParam Parameters
	handler := func(reqParam message.Parameters) (message.Parameters, error) {
		if s.SetupExtensions == nil {
			s.setDefaultSetupExtensions()
		}

		// Get any setup extensions
		rspParam, err = s.SetupExtensions(Parameters{reqParam})
		if err != nil {
			s.Logger.Error("SetupExtensions error during WebTransport", "remote_address", r.RemoteAddr, "error", err.Error())
			return nil, err
		}

		return rspParam.paramMap, nil
	}

	internalSess, err := internal.AcceptSession(context.Background(), conn, handler)
	if err != nil {
		s.Logger.Error("failed to create internal WebTransport session", "remote_address", r.RemoteAddr, "error", err.Error())
		return err
	}

	sess := newSession(internalSess, s.Handler)

	s.Logger.Debug("MOQ session established", "remote_address", r.RemoteAddr)

	s.SessionHandlerFunc(r.URL.Path, sess)

	s.Logger.Debug("session handling finished", "urlPath", r.URL.Path, "remote_address", r.RemoteAddr)

	return nil
}

func (s *Server) ListenAndServe() error {
	s.init()

	// Configure TLS for QUIC
	tlsConfig := s.TLSConfig
	if tlsConfig == nil {
		return errors.New("configuration for TLS is required for QUIC")
	}

	// Clone the TLS config to avoid modifying the original
	tlsConfig = tlsConfig.Clone()

	// Make sure we have NextProtos set for ALPN negotiation
	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{NextProtoMOQ}
	}

	// Start listener with configured TLS
	ln, err := quic.ListenAddrEarly(s.Addr, tlsConfig, s.QUICConfig)
	if err != nil {
		s.Logger.Error("failed to start QUIC listener", "address", s.Addr, "error", err.Error())
		return err
	}

	return s.serveQUICListener(ln)
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) (err error) {
	s.init()

	// Generate TLS configuration
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		s.Logger.Error("failed to load X509 key pair", "certFile", certFile, "keyFile", keyFile, "error", err.Error())
		return err
	}

	// Create TLS config with certificates
	tlsConfig := &tls.Config{
		Certificates: certs,
		NextProtos:   []string{NextProtoMOQ, http3.NextProtoH3},
	}
	s.TLSConfig = tlsConfig

	ln, err := quic.ListenAddrEarly(s.Addr, tlsConfig, s.QUICConfig)
	if err != nil {
		s.Logger.Error("failed to start QUIC listener for TLS", "address", s.Addr, "error", err.Error())
		return err
	}
	return s.serveQUICListener(ln)
}

func (s *Server) Close() error {
	s.lnMu.Lock()
	defer s.lnMu.Unlock()

	s.Logger.Info("closing server", "address", s.Addr)

	// Early return if already closed
	if s.closed.Load() {
		return nil
	}

	// Mark the server as closed
	s.closed.Store(true)

	// Send shutdown notification (cancels listener's Accept)
	close(s.shutdownChan)

	// Close all listeners
	if s.rawQUICListners != nil {
		s.Logger.Info("closing QUIC listeners", "address", s.Addr)
		for ln := range s.rawQUICListners {
			(*ln).Close()
		}
	}

	if s.WebtransportServer != nil {
		s.Logger.Info("closing WebTransport server", "address", s.Addr)
		s.WebtransportServer.Close()
	}

	// Wait for active connections to complete if any
	if s.connCount.Load() > 0 {
		<-s.doneChan
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.lnMu.Lock()

	// Early return if already closed
	if s.closed.Load() {
		s.lnMu.Unlock()
		return nil
	}

	// Mark the server as closed
	s.closed.Store(true)

	// Send shutdown notification (cancels listener's Accept)
	close(s.shutdownChan)

	// Close all listeners
	if s.rawQUICListners != nil {
		for ln := range s.rawQUICListners {
			(*ln).Close()
		}
	}
	s.lnMu.Unlock()

	// Use WebTransport server's shutdown if available
	if s.WebtransportServer != nil {
		return s.WebtransportServer.H3.Shutdown(ctx)
	}

	// For active connections, wait for completion or context cancellation
	if s.connCount.Load() > 0 {
		select {
		case <-s.doneChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (s *Server) setDefaultWebtransportServer() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.WebtransportServer = &webtransport.Server{
		H3: http3.Server{
			Addr: s.Addr,
		},
	}
	if s.TLSConfig != nil {
		s.WebtransportServer.H3.TLSConfig = s.TLSConfig.Clone()
	}
	if s.QUICConfig != nil {
		s.WebtransportServer.H3.QUICConfig = s.QUICConfig.Clone()
	}
	s.Logger.Debug("set default WebTransport server", "address", s.Addr)
}

func (s *Server) setDefaultSetupExtensions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SetupExtensions = NoSetupExtensions

	s.Logger.Debug("set Default setup extensions")
}

func (s *Server) addListener(ln *QUICEarlyListener) {
	s.lnMu.Lock()
	defer s.lnMu.Unlock()

	s.rawQUICListners[ln] = struct{}{}
}

func (s *Server) removeListener(ln *QUICEarlyListener) {
	s.lnMu.Lock()
	defer s.lnMu.Unlock()

	delete(s.rawQUICListners, ln)
}

func (s *Server) incrementConnCount() {
	s.connCount.Add(1)
}

func (s *Server) decrementConnCount() {
	newCount := s.connCount.Add(-1)

	// Send completion signal if connections reach zero and server is closed
	if newCount == 0 && s.closed.Load() {
		select {
		case s.doneChan <- struct{}{}:
		default:
			// Channel might already be closed
		}
	}
}

const NextProtoMOQ = "moq-00"

var NoSetupExtensions = func(req Parameters) (rsp Parameters, err error) {
	return Parameters{}, nil
}
