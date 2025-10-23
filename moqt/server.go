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
	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/quic/quicgo"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport/webtransportgo"
	"github.com/quic-go/quic-go/http3"
)

// Server is a MOQ server that accepts both WebTransport and raw QUIC connections.
// It handles session setup, track announcements, and subscriptions according to
// the MOQ Lite specification.
//
// The server maintains active sessions and listeners, providing graceful shutdown
// capabilities. It can serve over multiple listeners simultaneously.
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
	 * Set-up Request SetupHandler
	 */
	SetupHandler SetupHandler

	/*
	 * Listen QUIC function
	 */
	ListenFunc quic.ListenAddrFunc

	/*
	 * WebTransport Server
	 * If the server is configured with a WebTransport server, it is used to handle WebTransport sessions.
	 * If not, a default server is used.
	 */
	NewWebtransportServerFunc func(checkOrigin func(*http.Request) bool) webtransport.Server
	wtServer                  webtransport.Server

	/*
	 * Logger
	 */
	Logger *slog.Logger

	listenerMu    sync.RWMutex
	listeners     map[quic.Listener]struct{}
	listenerGroup sync.WaitGroup

	sessMu     sync.RWMutex
	activeSess map[*Session]struct{}

	initOnce sync.Once

	inShutdown atomic.Bool

	doneChan chan struct{} // Signal channel (notifies when server is completely closed)

}

func (s *Server) init() {
	s.initOnce.Do(func() {
		s.listeners = make(map[quic.Listener]struct{})
		s.doneChan = make(chan struct{})
		s.activeSess = make(map[*Session]struct{})
		// Initialize WebtransportServer

		var checkOrigin func(*http.Request) bool
		if s.Config != nil && s.Config.CheckHTTPOrigin != nil {
			checkOrigin = s.Config.CheckHTTPOrigin
		}

		if s.NewWebtransportServerFunc != nil {
			s.wtServer = s.NewWebtransportServerFunc(checkOrigin)
		} else {
			s.wtServer = webtransportgo.NewServer(checkOrigin)
		}

		if s.Logger != nil {
			s.Logger = s.Logger.With("address", s.Addr)
			s.Logger.Info("initialized server")
		}
	})
}

func (s *Server) ServeQUICListener(ln quic.Listener) error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	s.init()

	s.addListener(ln)
	defer s.removeListener(ln)

	var listenerLogger *slog.Logger
	if s.Logger != nil {
		listenerLogger = s.Logger.With("listener_address", ln.Addr())
	} else {
		listenerLogger = slog.New(slog.DiscardHandler)
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
			listenerLogger.Error("failed to accept QUIC connection",
				"error", err.Error(),
			)
			return err
		}

		// Handle connection in a goroutine
		go func(conn quic.Connection) {
			if err := s.ServeQUICConn(conn); err != nil {
				listenerLogger.Error("failed to handle connection",
					"error", err,
				)
			}
		}(conn)
	}
}

func (s *Server) ServeQUICConn(conn quic.Connection) error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	s.init()

	switch protocol := conn.ConnectionState().TLS.NegotiatedProtocol; protocol {
	case http3.NextProtoH3:
		return s.wtServer.ServeQUICConn(conn)
	case NextProtoMOQ:
		return s.handleNativeQUIC(conn)
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func (s *Server) ServeWebTransport(w http.ResponseWriter, r *http.Request) error {
	if s.shuttingDown() {
		return fmt.Errorf("server is shutting down")
	}

	s.init()

	conn, err := s.wtServer.Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	var connLogger *slog.Logger
	if s.Logger != nil {
		connLogger = s.Logger.With(
			"local_address", conn.LocalAddr(),
			"remote_address", conn.RemoteAddr(),
			"alpn", conn.ConnectionState().TLS.NegotiatedProtocol,
			"quic_version", conn.ConnectionState().Version,
		)
		// TODO: Add connection ID
	} else {
		connLogger = slog.New(slog.DiscardHandler)
	}

	connLogger.Debug("establishing a WebTransport session")

	acceptCtx, cancelAccept := context.WithTimeout(r.Context(), s.acceptTimeout())
	defer cancelAccept()
	sessStr, err := acceptSessionStream(acceptCtx, conn, connLogger)
	if err != nil {
		connLogger.Error("failed to accept session stream",
			"error", err,
		)
		return fmt.Errorf("failed to accept session stream: %w", err)
	}

	connLogger.Debug("accepted a session stream")

	// Set the path for the session
	sessStr.Path = r.URL.Path

	rsp := newResponseWriter(conn, sessStr, connLogger, s)
	req := sessStr.SetupRequest

	if s.SetupHandler != nil {
		connLogger.Debug("using custom setup handler")
		s.SetupHandler.ServeMOQ(rsp, req)
	} else {
		connLogger.Debug("no setup handler provided, using default router")
		DefaultRouter.ServeMOQ(rsp, req)
	}

	return nil
}

func (s *Server) handleNativeQUIC(conn quic.Connection) error {
	if s.shuttingDown() {
		return nil
	}

	s.init()

	var connLogger *slog.Logger
	if s.Logger != nil {
		connLogger = s.Logger.With(
			"local_address", conn.LocalAddr(),
			"remote_address", conn.RemoteAddr(),
			"alpn", conn.ConnectionState().TLS.NegotiatedProtocol,
			"quic_version", conn.ConnectionState().Version,
		)
		// TODO: Add connection ID
	} else {
		connLogger = slog.New(slog.DiscardHandler)
	}

	connLogger.Debug("moq: establishing a QUIC session")

	acceptCtx, cancelAccept := context.WithTimeout(conn.Context(), s.acceptTimeout())
	defer cancelAccept()
	sessStr, err := acceptSessionStream(acceptCtx, conn, connLogger)
	if err != nil {
		connLogger.Error("moq: failed to accept session stream",
			"error", err,
		)
		return err
	}

	rsp := newResponseWriter(conn, sessStr, connLogger, s)
	req := sessStr.SetupRequest

	if s.SetupHandler != nil {
		connLogger.Debug("moq: using custom setup handler")
		s.SetupHandler.ServeMOQ(rsp, req)
	} else {
		connLogger.Debug("moq: no setup handler provided, using default router")
		DefaultRouter.ServeMOQ(rsp, req)
	}

	return nil
}

// func (s *Server) Accept(w SetupResponseWriter, r *SetupRequest, mux *TrackMux) (*Session, error) {
// 	if w == nil {
// 		return nil, fmt.Errorf("response writer cannot be nil")
// 	}

// 	if s.shuttingDown() {
// 		w.Reject(SetupFailedErrorCode)
// 		return nil, ErrServerClosed
// 	}

// 	s.init()

// 	if r == nil {
// 		w.Reject(SetupFailedErrorCode)
// 		return nil, fmt.Errorf("request cannot be nil")
// 	}

// 	rsp, ok := w.(*responseWriter)
// 	if !ok {
// 		return nil, fmt.Errorf("response writer is not of type *response")
// 	}

// 	// Accept the setup request with default version and no extensions
// 	err := w.Accept(DefaultServerVersion, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to accept setup request: %w", err)
// 	}

// 	conn := rsp.conn
// 	if conn == nil {
// 		return nil, fmt.Errorf("quic connection cannot be nil")
// 	}

// 	var connLogger *slog.Logger
// 	if s.Logger != nil {
// 		connLogger = s.Logger.With(
// 			"local_address", conn.LocalAddr(),
// 			"remote_address", conn.RemoteAddr(),
// 			"alpn", conn.ConnectionState().TLS.NegotiatedProtocol,
// 			"quic_version", conn.ConnectionState().Version,
// 		)
// 		// TODO: Add connection ID
// 	} else {
// 		connLogger = slog.New(slog.DiscardHandler)
// 	}

// 	// var sess *Session
// 	// sess = newSession(conn, rsp.sessionStream, mux, connLogger, func() { s.removeSession(sess) })
// 	// s.addSession(sess)

// 	connLogger.Debug("accepted a new session")

// 	return sess, nil
// }

func acceptSessionStream(acceptCtx context.Context, conn quic.Connection, connLogger *slog.Logger) (*sessionStream, error) {
	sessionID := generateSessionID()

	sessLogger := connLogger.With(
		"session_id", sessionID,
	)

	sessLogger.Debug("establishing a session")

	stream, err := conn.AcceptStream(acceptCtx)
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
		} else {
			return nil, fmt.Errorf("moq: unexpected error occurred on session stream: %w", err)
		}
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

	// Get the client parameters
	clientParams := &Parameters{scm.Parameters}

	// Get the path parameter
	path, _ := clientParams.GetString(param_type_path)

	// serverParams, err := extensions(clientParams.Clone())
	// if err != nil {
	// 	sessLogger.Error("failed to process setup extensions",
	// 		"error", err,
	// 	)
	// 	return nil, err
	// }

	// version := DefaultServerVersion

	// // Send a SESSION_SERVER message
	// ssm := message.SessionServerMessage{
	// 	SelectedVersion: version,
	// 	Parameters:      serverParams.paramMap,
	// }
	// err = ssm.Encode(stream)
	// if err != nil {
	// 	sessLogger.Error("failed to send SESSION_SERVER message",
	// 		"error", err,
	// 	)

	// 	var appErr *quic.ApplicationError
	// 	if errors.As(err, &appErr) {
	// 		return nil, &SessionError{ApplicationError: appErr}
	// 	}

	// 	return nil, err
	// }

	req := &SetupRequest{
		ctx:              stream.Context(),
		Path:             path,
		Versions:         scm.SupportedVersions,
		ClientExtensions: clientParams,
	}

	return newSessionStream(stream, req), nil
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

	var ln quic.Listener
	var err error
	if s.ListenFunc != nil {
		ln, err = s.ListenFunc(s.Addr, tlsConfig, s.QUICConfig)
	} else {
		ln, err = quicgo.ListenAddrEarly(s.Addr, tlsConfig, s.QUICConfig)
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

	var ln quic.Listener
	if s.ListenFunc != nil {
		ln, err = s.ListenFunc(s.Addr, tlsConfig.Clone(), s.QUICConfig)
	} else {
		ln, err = quicgo.ListenAddrEarly(s.Addr, tlsConfig.Clone(), s.QUICConfig)
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

	s.goAway()

	// If there are no active sessions, signal done immediately so Shutdown
	// returns without waiting for the context timeout.
	s.sessMu.Lock()
	noSessions := len(s.activeSess) == 0
	s.sessMu.Unlock()
	if noSessions {
		select {
		case <-s.doneChan:
			// already closed
		default:
			close(s.doneChan)
		}
	}

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

func (s *Server) addListener(ln quic.Listener) {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()

	if s.listeners == nil {
		s.listeners = make(map[quic.Listener]struct{})
	}
	s.listeners[ln] = struct{}{}
	s.listenerGroup.Add(1)
}

func (s *Server) removeListener(ln quic.Listener) {
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

func (s *Server) goAway() {
	s.sessMu.Lock()
	defer s.sessMu.Unlock()

	for sess := range s.activeSess {
		sess.goAway("") // TODO: specify URI if needed
	}
}
