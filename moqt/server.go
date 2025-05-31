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

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
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

	/*
	 * Setup Extensions
	 * This function is called when a session is established
	 */

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
	listeners     map[quic.EarlyListener]struct{}
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
		s.listeners = make(map[quic.EarlyListener]struct{})
		s.doneChan = make(chan struct{})
		s.activeSess = make(map[*Session]struct{})
		s.nativeQUICCh = make(chan quic.Connection, 1<<4)

		// Initialize WebtransportServer

		if s.WebtransportServer == nil {
			s.WebtransportServer = webtransport.NewDefaultServer(s.Addr)
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
		if s.WebtransportServer == nil {
			s.WebtransportServer = webtransport.NewDefaultServer(s.Addr)
		}

		if logger != nil {
			logger.Debug("handling webtransport session",
				"remote_address", conn.RemoteAddr(),
			)
		}

		return s.WebtransportServer.ServeQUICConn(conn)
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
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-s.nativeQUICCh:
		logger := s.Logger
		if logger != nil {
			logger = logger.With(
				"remote_address", conn.RemoteAddr(),
			)
			logger.Debug("handling quic connection", "remote_address", conn.RemoteAddr())
		}

		var path string
		// Listen the session stream
		extensions := func(clientParams *Parameters) (*Parameters, error) {
			var err error

			// Get the path parameter
			path, err = clientParams.GetString(param_type_path)
			if err != nil {
				if logger != nil {
					logger.Error("failed to get 'path' parameter",
						"error", err.Error(),
					)
				}
				return nil, err
			}

			// Get any setup extensions
			if s.Config == nil || s.Config.ServerSetupExtensions == nil {
				if logger != nil {
					logger.Debug("no setup extensions provided, using default parameters")
				}
				return NewParameters(), nil
			}

			params, err := s.Config.ServerSetupExtensions(clientParams)
			if err != nil {
				if logger != nil {
					logger.Error("failed to get setup extensions",
						"error", err.Error(),
					)
				}
				return nil, err
			}
			if params == nil {
				if logger != nil {
					logger.Debug("server setup extensions returned nil, using default parameters")
				}
				return NewParameters(), nil
			}

			return params, nil
		}

		acceptCtx, cancelAccept := context.WithTimeout(ctx, s.acceptTimeout())
		defer cancelAccept()
		return s.acceptSession(acceptCtx, path, conn, extensions, mux)
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

	logger := s.Logger
	if logger != nil {
		logger = logger.With(
			"remote_address", r.RemoteAddr,
		)
		logger.Debug("accepting webtransport session")
	}

	if s.WebtransportServer == nil {
		s.WebtransportServer = webtransport.NewDefaultServer(s.Addr)
		if logger != nil {
			logger.Debug("using default WebTransport server")
		}
	}
	conn, err := s.WebtransportServer.Upgrade(w, r)
	if err != nil {
		if logger != nil {
			logger.Error("failed to upgrade http to webtransport",
				"error", err.Error(),
			)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	extensions := func(clientParams *Parameters) (*Parameters, error) {
		if s.Config == nil || s.Config.ServerSetupExtensions == nil {
			if logger != nil {
				logger.Debug("no setup extensions provided, using default parameters")
			}
			return NewParameters(), nil
		}

		params, err := s.Config.ServerSetupExtensions(clientParams)
		if err != nil {
			if logger != nil {
				logger.Error("failed to get setup extensions",
					"error", err.Error(),
				)
			}
			return nil, err
		}
		if params == nil {
			if logger != nil {
				logger.Debug("server setup extensions returned nil, using default parameters")
			}
			return NewParameters(), nil
		}

		return params.Clone(), nil
	}

	if logger != nil {
		logger.Debug("WebTransport session established", "remote_address", r.RemoteAddr)
	}

	acceptCtx, cancelAccept := context.WithTimeout(r.Context(), s.acceptTimeout())
	defer cancelAccept()
	return s.acceptSession(acceptCtx, r.URL.Path, conn, extensions, mux)
}

func (s *Server) acceptSession(acceptCtx context.Context, path string, conn quic.Connection, extensions func(*Parameters) (*Parameters, error), mux *TrackMux) (*Session, error) {
	logger := s.Logger
	var sessTracer *moqtrace.SessionTracer
	if s.Config != nil && s.Config.Tracer != nil {
		sessTracer = s.Config.Tracer()
		// This should not be nil, and if it is, panic occurs
		moqtrace.InitSessionTracer(sessTracer)
	} else {
		sessTracer = moqtrace.DefaultSessionTracer()
	}

	stream, err := conn.AcceptStream(acceptCtx)
	if err != nil {
		if logger != nil {
			logger.Error("failed to accept a session stream",
				"error", err,
			)
		}
		return nil, fmt.Errorf("failed to accept a session stream: %w", err)
	}
	streamTracer := sessTracer.QUICStreamAccepted(stream.StreamID())

	var stm message.StreamTypeMessage
	_, err = stm.Decode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to get a STREAM_TYPE message",
				"error", err,
			)
		}
	}
	streamTracer.StreamTypeMessageReceived(stm)

	var scm message.SessionClientMessage
	_, err = scm.Decode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to get a SESSION_CLIENT message",
				"error", err,
			)
		}

		stream.CancelRead(ErrInternalError.StreamErrorCode())
		stream.CancelWrite(ErrInternalError.StreamErrorCode())
		return nil, fmt.Errorf("failed to get a SESSION_CLIENT message: %w", err)
	}
	streamTracer.SessionClientMessageReceived(scm)

	clientParams := &Parameters{scm.Parameters}
	serverParams, err := extensions(clientParams.Clone())
	if err != nil {
		return nil, err
	}

	// Set the selected version and parameters
	version := internal.DefaultServerVersion

	//

	// Send a SESSION_SERVER message
	ssm := message.SessionServerMessage{
		SelectedVersion: version,
		Parameters:      serverParams.paramMap,
	}
	_, err = ssm.Encode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to send a SESSION_SERVER message", "error", err)
		}
		return nil, err
	}
	streamTracer.SessionServerMessageSent(ssm)

	sessCtx := newSessionContext(conn.Context(), version, path, clientParams, serverParams, logger, sessTracer)

	sessstr := newSessionStream(sessCtx, stream, streamTracer)

	sess := newSession(sessCtx, sessstr, conn, mux)

	s.addSession(sess)
	go func() {
		<-sess.ctx.Done()
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
		if s.Logger != nil {
			s.Logger.Error("failed to start QUIC listener", "address", s.Addr, "error", err.Error())
		}
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
	s.TLSConfig = tlsConfig.Clone()
	ln, err := quic.ListenQUICFunc(s.Addr, tlsConfig, s.QUICConfig)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("failed to start QUIC listener for TLS", "address", s.Addr, "error", err.Error())
		}
		return err
	}

	return s.ServeQUICListener(ln)
}

func (s *Server) Close() error {
	s.inShutdown.Store(true)
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Logger != nil {
		s.Logger.Info("closing server", "address", s.Addr)
	}
	// Close all listeners
	if s.listeners != nil {
		if s.Logger != nil {
			s.Logger.Info("closing QUIC listeners", "address", s.Addr)
		}
		for ln := range s.listeners {
			ln.Close()
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
		ln.Close()
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
				go sess.Terminate(ErrGoAwayTimeout)
				s.removeSession(sess)
			}
			return ctx.Err()
		}
	}

	return nil
}

func (s *Server) addListener(ln quic.EarlyListener) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listeners == nil {
		s.listeners = make(map[quic.EarlyListener]struct{})
	}
	s.listeners[ln] = struct{}{}
	s.listenerGroup.Add(1)
}

func (s *Server) removeListener(ln quic.EarlyListener) {
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
	if s.Config != nil && s.Config.Timeout != 0 {
		return s.Config.Timeout
	}
	return 5 * time.Second
}

func (s *Server) goAway(sess *Session) {
	// TODO: Implement go away
}

const NextProtoMOQ = "moq-00"
