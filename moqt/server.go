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

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
	quicgo "github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	webtransportgo "github.com/quic-go/webtransport-go"
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
	QUICConfig *quicgo.Config

	/*
	 * MOQ Configuration
	 */
	Config *Config

	/*
	 * Setup Extensions
	 * This function is called when a session is established
	 */
	SetupExtensions func(req *Parameters) (rsp *Parameters, err error)

	/*
	 * Handlers
	 * - TrackHandler:
	 * - AnnouncementHandler:
	 */

	//
	AcceptTimeout time.Duration // TODO: Rename

	/*
	 * Logger
	 */
	Logger *slog.Logger

	/*
	 * WebTransport Server
	 * If the server is configured with a WebTransport server, it is used to handle WebTransport sessions.
	 * If not, a default server is used.
	 */
	WebtransportServer webtransport.Server

	mu            sync.RWMutex
	listeners     map[*quic.EarlyListener]struct{}
	listenerGroup sync.WaitGroup
	//
	activeSess map[*Session]struct{}
	// onShutdown []func() // TODO: Implement if needed

	// cancelFuncs []context.CancelFunc

	initOnce sync.Once

	inShutdown atomic.Bool

	nativeQUICCh chan quic.Connection

	doneChan chan struct{} // Signal channel (notifies when server is completely closed)
	// connCount    atomic.Int64  // Active connection counter
}

func (s *Server) init() {
	s.initOnce.Do(func() {
		if s.Logger == nil {
			s.Logger = slog.Default()
		}
		s.listeners = make(map[*quic.EarlyListener]struct{})

		// Initialize signal channels
		s.doneChan = make(chan struct{})

		// Initialize WebtransportServer
		if s.WebtransportServer == nil {
			s.setDefaultWebtransportServer()
		}

		if s.nativeQUICCh == nil {
			s.nativeQUICCh = make(chan quic.Connection, 1<<4)
		}

		s.Logger.Debug("initialized server", "address", s.Addr)
	})
}

func (s *Server) ServeQUICListener(ln quic.EarlyListener) error {
	if s.shuttingDown() {
		return errors.New("server is shutting down")
	}

	s.init()

	s.addListener(&ln)
	defer s.removeListener(&ln)

	// Create context for listener's Accept operation
	// This context will be canceled when the server is shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Logger.Debug("listening for QUIC connections", "listener", ln)

	for {
		if s.shuttingDown() {
			return errors.New("server is shutting down")
		}

		// Listen for new QUIC connections
		conn, err := ln.Accept(ctx)
		if err != nil {
			return err
		}

		s.Logger.Info("Accepted new QUIC connection", "remote_address", conn.RemoteAddr())

		// Handle connection in a goroutine
		go func(conn quic.Connection) {
			if err := s.ServeQUICConn(conn); err != nil {
				s.Logger.Debug("handling connection failed", "error", err)
			}
		}(conn)
	}
}

func (s *Server) ServeQUICConn(conn quic.Connection) error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	s.init()

	protocol := conn.ConnectionState().TLS.NegotiatedProtocol

	s.Logger.Info("Negotiated protocol", "remote_address", conn.RemoteAddr(), "protocol", protocol)

	switch protocol {
	case http3.NextProtoH3:
		s.Logger.Debug("handling webtransport session", "remote_address", conn.RemoteAddr())
		if s.WebtransportServer == nil {
			s.setDefaultWebtransportServer()
		}
		return s.WebtransportServer.ServeQUICConn(conn)
	case NextProtoMOQ:
		select {
		case s.nativeQUICCh <- conn:
		default:
			conn.CloseWithError(quic.ConnectionErrorCode(quicgo.ConnectionRefused), "")
		}
		return nil
	default:
		s.Logger.Error("unsupported negotiated protocol", "remote_address", conn.RemoteAddr(), "protocol", protocol) // Updated to use conn.RemoteAddr()
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func (s *Server) AcceptQUIC(ctx context.Context, mux *TrackMux) (string, *Session, error) {
	if s.shuttingDown() {
		return "", nil, ErrServerClosed
	}
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	case conn := <-s.nativeQUICCh:
		s.Logger.Debug("handling quic connection", "remote_address", conn.RemoteAddr())

		var path string
		// Listen the session stream
		params := func(reqParam *Parameters) (*Parameters, error) {
			var err error

			// Get the path parameter
			path, err = reqParam.GetString(param_type_path)
			if err != nil {
				s.Logger.Error("failed to get 'path' parameter", "remote_address", conn.RemoteAddr(), "error", err.Error())
				return nil, err
			}

			// Get any setup extensions
			rspParam, err := s.SetupExtensions(reqParam)
			if err != nil {
				s.Logger.Error("failed to get setup extensions", "remote_address", conn.RemoteAddr(), "error", err.Error())
				return nil, err
			}

			return rspParam, nil
		}

		sess, err := s.acceptSession(conn, params, mux)
		if err != nil {
			return "", nil, err
		}

		return path, sess, nil
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

	if s.WebtransportServer == nil {
		s.setDefaultWebtransportServer()
	}

	conn, err := s.WebtransportServer.Upgrade(w, r)
	if err != nil {
		s.Logger.Error("WebTransport upgrade failed", "remote_address", r.RemoteAddr, "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	s.Logger.Debug("WebTransport session established", "remote_address", r.RemoteAddr)

	params := s.SetupExtensions

	return s.acceptSession(conn, params, mux)
}

func (s *Server) acceptSession(conn quic.Connection, params func(req *Parameters) (rsp *Parameters, err error), mux *TrackMux) (*Session, error) {
	sessCtx := newSessionContext(conn.Context(), s.Logger)
	sess := newSession(sessCtx, conn, mux)

	ctxAccept, cancelAccept := context.WithTimeout(sessCtx, s.acceptTimeout())
	defer cancelAccept()
	err := sess.acceptSessionStream(ctxAccept, params)
	if err != nil {
		slog.Error("failed to accept session stream", "error", err)
		return nil, err
	}

	s.addSession(sess)
	go func() {
		<-sessCtx.Done()
		s.removeSession(sess)
	}()

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

	if quic.ListenQUICFunc == nil {
		panic("ListenQUICFunc is nil")
	}
	// Start listener with configured TLS
	ln, err := quic.ListenQUICFunc(s.Addr, tlsConfig, s.QUICConfig)
	if err != nil {
		s.Logger.Error("failed to start QUIC listener", "address", s.Addr, "error", err.Error())
		return err
	}

	return s.ServeQUICListener(ln)
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
	s.TLSConfig = tlsConfig.Clone()

	ln, err := quic.ListenQUICFunc(s.Addr, tlsConfig, s.QUICConfig)
	if err != nil {
		s.Logger.Error("failed to start QUIC listener for TLS", "address", s.Addr, "error", err.Error())
		return err
	}

	return s.ServeQUICListener(ln)
}

func (s *Server) Close() error {
	s.inShutdown.Store(true)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Logger.Info("closing server", "address", s.Addr)

	// Close all listeners
	if s.listeners != nil {
		s.Logger.Info("closing QUIC listeners", "address", s.Addr)
		for ln := range s.listeners {
			(*ln).Close()
		}
	}

	for sess := range s.activeSess {
		(*sess).Terminate(NoErrTerminate)
		s.removeSession(sess)
	}

	// Wait for active connections to complete if any
	if len(s.activeSess) > 0 {
		<-s.doneChan
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.inShutdown.Store(true)

	s.mu.Lock()
	defer s.mu.Unlock()

	for ln := range s.listeners {
		(*ln).Close()
	}
	s.listeners = nil

	// Wait
	s.mu.Unlock()
	s.listenerGroup.Wait()
	s.mu.Lock()

	// Go away all active sessions
	for sess := range s.activeSess {
		s.goAway(sess)
	}

	// For active connections, wait for completion or context cancellation
	if len(s.activeSess) > 0 {
		select {
		case <-s.doneChan:
			return nil
		case <-ctx.Done():
			for sess := range s.activeSess {
				(*sess).Terminate(ErrGoAwayTimeout)
				s.removeSession(sess)
			}
			return ctx.Err()
		}
	}

	return nil
}

func (s *Server) setDefaultWebtransportServer() {
	wtserver := &webtransportgo.Server{
		H3: http3.Server{
			Addr: s.Addr,
		},
	}

	// Wrap the WebTransport server
	s.WebtransportServer = webtransport.WrapWebTransportServer(wtserver)

	s.Logger.Debug("set default WebTransport server", "address", s.Addr)
}

func (s *Server) addListener(ln *quic.EarlyListener) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listeners == nil {
		s.listeners = make(map[*quic.EarlyListener]struct{})
	}
	s.listeners[ln] = struct{}{}
	s.listenerGroup.Add(1)
}

func (s *Server) removeListener(ln *quic.EarlyListener) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listeners == nil {
		return
	}
	delete(s.listeners, ln)
	s.listenerGroup.Done()
}

func (s *Server) addSession(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess == nil {
		return
	}
	s.activeSess[sess] = struct{}{}
}

func (s *Server) removeSession(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeSess, sess)

	// Send completion signal if connections reach zero and server is closed
	if len(s.activeSess) == 0 && s.shuttingDown() {
		select {
		case s.doneChan <- struct{}{}:
		default:
			// Channel might already be closed
		}
	}
}

func (s *Server) shuttingDown() bool {
	return s.inShutdown.Load()
}

func (s *Server) acceptTimeout() time.Duration {
	if s.AcceptTimeout != 0 {
		return s.AcceptTimeout
	}
	return 5 * time.Second
}

func (s *Server) goAway(sess *Session) {
	// TODO: Implement go away
	// sess.goAway("")
}

const NextProtoMOQ = "moq-00"
