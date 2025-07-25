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
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic/quicgo"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport/webtransportgo"
	"github.com/quic-go/quic-go/http3"
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

	ListenFunc func(addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.EarlyListener, error)

	/*
	 * Logger
	 */
	Logger *slog.Logger

	/*
	 * WebTransport Server
	 * If the server is configured with a WebTransport server, it is used to handle WebTransport sessions.
	 * If not, a default server is used.
	 */
	ServeWebtransportFunc func(addr string, tlsConfig *tls.Config, quicConfig *quic.Config,
		checkOrigin func(*http.Request) bool) webtransport.Server
	wtServer webtransport.Server

	listenerMu    sync.RWMutex
	listeners     map[quic.EarlyListener]struct{}
	listenerGroup sync.WaitGroup

	sessMu     sync.RWMutex
	activeSess map[*Session]struct{}

	initOnce sync.Once

	inShutdown atomic.Bool

	nativeQUICCh chan quic.Connection

	doneChan chan struct{} // Signal channel (notifies when server is completely closed)
}

func (s *Server) init() {
	s.initOnce.Do(func() {
		s.listeners = make(map[quic.EarlyListener]struct{})
		s.doneChan = make(chan struct{})
		s.activeSess = make(map[*Session]struct{})
		s.nativeQUICCh = make(chan quic.Connection, 1<<4)
		// Initialize WebtransportServer

		if s.ServeWebtransportFunc == nil {
			// If not set, create a default WebTransport server
			s.wtServer = webtransportgo.NewDefaultServer(s.Addr, s.TLSConfig, s.QUICConfig, s.Config.CheckHTTPOrigin)
		} else {
			s.wtServer = s.ServeWebtransportFunc(s.Addr, s.TLSConfig, s.QUICConfig, s.Config.CheckHTTPOrigin)
		}

		if s.Logger != nil {
			s.Logger = s.Logger.With("address", s.Addr)
			s.Logger.Debug("initialized server")
		}
	})
}

func (s *Server) ServeQUICListener(ln quic.EarlyListener) error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	s.init()

	s.addListener(ln)
	defer s.removeListener(ln)

	logger := s.Logger
	if logger != nil {
		logger.Debug("listening for QUIC connections")
	}

	// Create context for listener's Accept operation
	// This context will be canceled when the server is shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		if s.shuttingDown() {
			return ErrServerClosed
		}

		// Listen for new QUIC connections
		conn, err := ln.Accept(ctx)
		if err != nil {
			if logger != nil {
				logger.Error("failed to accept QUIC connection",
					"error", err.Error(),
				)
			}
			return err
		}

		if logger != nil {
			logger = logger.With(
				"remote_address", conn.RemoteAddr(),
			)
			logger.Debug("accepted a new QUIC connection")
		}

		// Handle connection in a goroutine
		go func(conn quic.Connection) {
			if err := s.ServeQUICConn(conn); err != nil {
				if logger != nil {
					logger.Debug("failed to handle connection",
						"error", err,
					)
				}
			}
		}(conn)
	}
}

func (s *Server) ServeQUICConn(conn quic.Connection) error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	s.init()

	logger := s.Logger
	if logger != nil {
		logger = logger.With(
			"remote_address", conn.RemoteAddr(),
		)
	}

	switch protocol := conn.ConnectionState().TLS.NegotiatedProtocol; protocol {
	case http3.NextProtoH3:
		if logger != nil {
			logger.Debug("handling webtransport session",
				"remote_address", conn.RemoteAddr(),
			)
		}

		return s.wtServer.ServeQUICConn(conn)
	case NextProtoMOQ:
		select {
		case s.nativeQUICCh <- conn:
		default:
			conn.CloseWithError(quic.ConnectionErrorCode(quic.ConnectionRefused), "")
		}
		return nil
	default:
		if logger != nil {
			logger.Error("unsupported negotiated protocol",
				"remote_address", conn.RemoteAddr(),
				"protocol", protocol,
			)
		}
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func (s *Server) AcceptQUIC(ctx context.Context, mux *TrackMux) (*Session, error) {
	if s.shuttingDown() {
		return nil, ErrServerClosed
	}

	s.init()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-s.nativeQUICCh:
		if s.shuttingDown() {
			return nil, ErrServerClosed
		}

		var connLogger *slog.Logger
		if s.Logger != nil {
			connLogger = s.Logger.With(
				"remote_address", conn.RemoteAddr().String(),
				"alpn", conn.ConnectionState().TLS.NegotiatedProtocol,
				"quic_version", conn.ConnectionState().Version,
			)
		} else {
			connLogger = slog.New(slog.DiscardHandler)
		}

		connLogger.Debug("establishing a session over QUIC connection")

		var path *string

		// Listen the session stream
		extensions := func(clientParams *Parameters) (*Parameters, error) {
			var err error

			// Get the path parameter
			*path, err = clientParams.GetString(param_type_path)
			if err != nil {
				connLogger.Error("failed to get 'path' parameter", "error", err)
				return nil, err
			}

			// Get any setup extensions
			if s.Config == nil || s.Config.ServerSetupExtensions == nil {
				connLogger.Debug("no setup extensions provided, using default parameters")
				return NewParameters(), nil
			}

			params, err := s.Config.ServerSetupExtensions(clientParams)
			if err != nil {
				connLogger.Error("failed to get setup extensions", "error", err)
				return nil, err
			}

			if params == nil {
				connLogger.Debug("server setup extensions returned nil, using default parameters")
				return NewParameters(), nil
			}

			return params, nil
		}

		acceptCtx, cancelAccept := context.WithTimeout(ctx, s.acceptTimeout())
		defer cancelAccept()
		return s.acceptSession(acceptCtx, path, conn, extensions, mux, connLogger)
	}
}

// ServeWebTransport serves a WebTransport session.
// It upgrades the HTTP/3 connection to a WebTransport session and calls the session handler.
// If the server is not configured with a WebTransport server, it creates a default server.
func (s *Server) AcceptWebTransport(w http.ResponseWriter, r *http.Request, mux *TrackMux) (*Session, error) {
	if s.shuttingDown() {
		return nil, ErrServerClosed
	}

	s.init()

	var logger *slog.Logger
	if s.Logger != nil {
		var protocol string
		if r.TLS == nil {
			protocol = "none"
		} else {
			protocol = r.TLS.NegotiatedProtocol
		}
		logger = s.Logger.With(
			"remote_address", r.RemoteAddr,
			"alpn", protocol,
			"url_path", r.URL.Path,
		)
	} else {
		logger = slog.New(slog.DiscardHandler)
	}

	conn, err := s.wtServer.Upgrade(w, r)
	if err != nil {
		logger.Error("failed to upgrade HTTP to WebTransport",
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	connLogger := logger.With(
		"quic_version", conn.ConnectionState().Version,
	)

	connLogger.Debug("WebTransport connection established")

	extensions := func(clientParams *Parameters) (*Parameters, error) {
		if s.Config == nil || s.Config.ServerSetupExtensions == nil {
			connLogger.Debug("no setup extensions provided, using default parameters")
			return NewParameters(), nil
		}

		params, err := s.Config.ServerSetupExtensions(clientParams)
		if err != nil {
			connLogger.Error("failed to get setup extensions",
				"error", err,
			)
			return nil, err
		}

		if params == nil {
			connLogger.Debug("server setup extensions returned nil, using default parameters")
			return NewParameters(), nil
		}

		return params.Clone(), nil
	}

	setupCtx, cancelSetup := context.WithTimeout(r.Context(), s.acceptTimeout())
	defer cancelSetup()
	return s.acceptSession(setupCtx, &r.URL.Path, conn, extensions, mux, connLogger)
}

func (s *Server) acceptSession(setupCtx context.Context, path *string, conn quic.Connection,
	extensions func(*Parameters) (*Parameters, error), mux *TrackMux, connLogger *slog.Logger) (*Session, error) {

	sessionID := generateSessionID()

	sessLogger := connLogger.With(
		"session_id", sessionID,
		"path", path,
	)

	sessLogger.Debug("establishing a session")

	stream, err := conn.AcceptStream(setupCtx)
	if err != nil {
		sessLogger.Error("failed to accept a session stream",
			"error", err,
		)

		return nil, fmt.Errorf("failed to accept a session stream: %w", err)
	}

	var streamType message.StreamType
	err = streamType.Decode(stream)
	if err != nil {
		sessLogger.Error("failed to receive STREAM_TYPE message",
			"error", err,
		)

		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			return nil, &SessionError{ApplicationError: appErr}
		}

		return nil, err
	}

	streamLogger := sessLogger.With(
		"stream_id", stream.StreamID(),
	)

	streamLogger.Debug("accepted a session stream")

	var scm message.SessionClientMessage
	err = scm.Decode(stream)
	if err != nil {
		streamLogger.Error("failed to receive SESSION_CLIENT message",
			"error", err,
		)

		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			return nil, &SessionError{ApplicationError: appErr}
		}

		return nil, err
	}

	clientParams := &Parameters{scm.Parameters}

	serverParams, err := extensions(clientParams.Clone())
	if err != nil {
		sessLogger.Error("failed to process setup extensions",
			"error", err,
		)
		return nil, err
	}

	// Use default server version
	version := DefaultServerVersion

	// Send a SESSION_SERVER message
	ssm := message.SessionServerMessage{
		SelectedVersion: version,
		Parameters:      serverParams.paramMap,
	}
	err = ssm.Encode(stream)
	if err != nil {
		sessLogger.Error("failed to send SESSION_SERVER message",
			"error", err,
		)

		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			return nil, &SessionError{ApplicationError: appErr}
		}

		return nil, err
	}

	// Create session
	sess := newSession(conn, version, *path, clientParams, serverParams,
		stream, mux, sessLogger)

	s.addSession(sess)

	go func() {
		<-sess.Context().Done()
		s.removeSession(sess)
	}()

	sessLogger.Debug("moq: session established")

	return sess, nil
}

func (s *Server) ListenAndServe() error {
	s.init()

	// Configure TLS for QUIC
	if s.TLSConfig == nil {
		return errors.New("configuration for TLS is required for QUIC")
	}

	// Clone the TLS config to avoid modifying the original
	tlsConfig := s.TLSConfig.Clone()

	// Make sure we have NextProtos set for ALPN negotiation
	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{NextProtoMOQ}
	}

	var ln quic.EarlyListener
	var err error
	if s.ListenFunc != nil {
		ln, err = s.ListenFunc(s.Addr, tlsConfig, s.QUICConfig)
	} else {
		ln, err = quicgo.Listen(s.Addr, tlsConfig, s.QUICConfig)
	}
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("failed to start QUIC listener", "address", s.Addr, "error", err.Error())
		}
		return err
	}

	return s.ServeQUICListener(ln)
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	if s.shuttingDown() {
		return ErrServerClosed
	}
	s.init()

	var err error
	// Generate TLS configuration
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("failed to load X509 key pair", "certFile", certFile, "keyFile", keyFile, "error", err.Error())
		}
		return err
	}

	// Create TLS config with certificates
	tlsConfig := &tls.Config{
		Certificates: certs,
		NextProtos:   []string{NextProtoMOQ, http3.NextProtoH3},
	}

	var ln quic.EarlyListener
	if s.ListenFunc != nil {
		ln, err = s.ListenFunc(s.Addr, tlsConfig.Clone(), s.QUICConfig)
	} else {
		ln, err = quicgo.Listen(s.Addr, tlsConfig.Clone(), s.QUICConfig)
	}
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("failed to start QUIC listener for TLS", "address", s.Addr, "error", err.Error())
		}
		return err
	}

	return s.ServeQUICListener(ln)
}

func (s *Server) Close() error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	// Set the shutdown flag
	s.inShutdown.Store(true)

	// Ensure that the server is initialized
	s.init()

	if s.Logger != nil {
		s.Logger.Info("closing server", "address", s.Addr)
	}
	// Terminate all active sessions
	s.sessMu.Lock()
	if len(s.activeSess) > 0 {
		for sess := range s.activeSess {
			go sess.Terminate(NoError, NoError.String())
		}

		s.sessMu.Unlock()

		<-s.doneChan
	} else {
		s.sessMu.Unlock()

		// No active sessions, close doneChan immediately
		select {
		case <-s.doneChan:
			// Already closed
		default:
			close(s.doneChan)
		}
	} // Close all listeners
	s.listenerMu.Lock()
	if len(s.listeners) > 0 {
		if s.Logger != nil {
			s.Logger.Info("closing QUIC listeners", "address", s.Addr)
		}
		for ln := range s.listeners {
			ln.Close()
		}
	}
	s.listenerMu.Unlock()

	// Wait for all listeners to close
	s.listenerGroup.Wait()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	// Set the shutdown flag
	s.inShutdown.Store(true)

	// Go away all active sessions
	s.sessMu.Lock()
	for sess := range s.activeSess {
		s.goAway(sess)
	}
	s.sessMu.Unlock()

	select {
	case <-s.doneChan:
		// Already closed
	case <-ctx.Done():
		// Context canceled, terminate all sessions forcefully
		s.sessMu.Lock()
		if len(s.activeSess) > 0 {
			for sess := range s.activeSess {
				go sess.Terminate(GoAwayTimeoutErrorCode, GoAwayTimeoutErrorCode.String())
			}
			s.sessMu.Unlock()

			// Wait for all sessions to close
			<-s.doneChan
		}
	}

	// Close all listeners
	s.listenerMu.Lock()
	for ln := range s.listeners {
		ln.Close()
	}
	s.listeners = nil
	s.listenerMu.Unlock()

	// Wait for all listeners to close
	s.listenerGroup.Wait()

	return nil
}

func (s *Server) addListener(ln quic.EarlyListener) {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()

	if s.listeners == nil {
		s.listeners = make(map[quic.EarlyListener]struct{})
	}
	s.listeners[ln] = struct{}{}
	s.listenerGroup.Add(1)
}

func (s *Server) removeListener(ln quic.EarlyListener) {
	s.listenerMu.Lock()

	_, ok := s.listeners[ln]
	if ok {
		delete(s.listeners, ln)
	}

	s.listenerMu.Unlock()

	if ok {
		s.listenerGroup.Done()
	}
}

func (s *Server) addSession(sess *Session) {
	s.sessMu.Lock()
	defer s.sessMu.Unlock()

	if sess == nil {
		return
	}
	s.activeSess[sess] = struct{}{}
}

func (s *Server) removeSession(sess *Session) {
	s.sessMu.Lock()
	defer s.sessMu.Unlock()

	_, ok := s.activeSess[sess]
	if !ok {
		return
	}

	delete(s.activeSess, sess)

	// Send completion signal if connections reach zero and server is closed
	if len(s.activeSess) == 0 && s.shuttingDown() {
		// Close the done channel to signal server is done
		select {
		case <-s.doneChan:
			// Already closed
		default:
			close(s.doneChan)
		}
	}
}

func (s *Server) shuttingDown() bool {
	return s.inShutdown.Load()
}

func (s *Server) acceptTimeout() time.Duration {
	if s.Config != nil && s.Config.SetupTimeout != 0 {
		return s.Config.SetupTimeout
	}
	return 5 * time.Second
}

func (s *Server) goAway(sess *Session) {
	// TODO: Implement go away
}
