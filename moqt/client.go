package moqt

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic/quicgo"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport/webtransportgo"
)

type Client struct {
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

	/***/
	DialQUICConn func(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.Connection, error)

	DialWebTransportFunc func(ctx context.Context, addr string, header http.Header) (*http.Response, quic.Connection, error)

	/*
	 * Logger
	 */
	Logger *slog.Logger

	//
	initOnce sync.Once

	sessMu     sync.RWMutex
	activeSess map[*Session]struct{}

	inShutdown atomic.Bool
	doneChan   chan struct{}
}

func (c *Client) init() {
	c.initOnce.Do(func() {
		c.activeSess = make(map[*Session]struct{})
		c.doneChan = make(chan struct{}, 1)

		if c.Logger != nil {
			c.Logger.Info("client initialized")
		}
	})
}

func (c *Client) dialTimeout() time.Duration {
	if c.Config != nil && c.Config.SetupTimeout != 0 {
		return c.Config.SetupTimeout
	}

	return 5 * time.Second
}

func (c *Client) Dial(ctx context.Context, urlStr string, mux *TrackMux) (*Session, error) {
	sessionID := generateSessionID()
	var logger *slog.Logger
	if c.Logger == nil {
		logger = slog.New(slog.DiscardHandler)
	} else {
		logger = c.Logger.With("session_id", sessionID, "url", urlStr)
	}

	logger.Info("dial started")

	if c.shuttingDown() {
		logger.Warn("dial rejected: client shutting down")
		return nil, ErrClientClosed
	}
	c.init()

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Error("URL parsing failed", "error", err)
		return nil, err
	}

	// Dial based on the scheme
	switch parsedURL.Scheme {
	case "https":
		logger.Debug("using WebTransport protocol",
			"scheme", "https",
			"host", parsedURL.Hostname(),
			"port", parsedURL.Port(),
			"path", parsedURL.Path,
		)
		return c.DialWebTransport(ctx, parsedURL.Hostname()+":"+parsedURL.Port(), parsedURL.Path, mux)
	case "moqt":
		logger.Debug("using QUIC protocol",
			"scheme", "moqt",
			"host", parsedURL.Hostname(),
			"port", parsedURL.Port(),
			"path", parsedURL.Path,
		)
		return c.DialQUIC(ctx, parsedURL.Hostname()+":"+parsedURL.Port(), parsedURL.Path, mux)
	default:
		logger.Error("unsupported URL scheme", "scheme", parsedURL.Scheme)
		return nil, ErrInvalidScheme
	}
}

// generateSessionID creates a unique session identifier for logging
func generateSessionID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (c *Client) DialWebTransport(ctx context.Context, host, path string, mux *TrackMux) (*Session, error) {
	sessionID := generateSessionID()
	endpoint := "https://" + host + path

	var logger *slog.Logger
	if c.Logger == nil {
		// Return a logger that discards all output
		logger = slog.New(slog.DiscardHandler)
	} else {
		// Create logger with session context
		logger = c.Logger.With("session_id", sessionID, "endpoint", endpoint)
	}

	logger.Info("initiating WebTransport connection")

	if c.shuttingDown() {
		logger.Warn("WebTransport dial rejected: client shutting down")
		return nil, ErrClientClosed
	}
	c.init()

	dialTimeout := c.dialTimeout()
	dialCtx, cancelDial := context.WithTimeout(ctx, dialTimeout)
	defer cancelDial()

	logger.Debug("starting WebTransport dial", "timeout", dialTimeout)

	var conn quic.Connection
	var err error

	if c.DialWebTransportFunc != nil {
		logger.Debug("using custom WebTransport dial function")
		_, conn, err = c.DialWebTransportFunc(dialCtx, host+path, http.Header{})
	} else {
		logger.Debug("using default WebTransport dial")
		_, conn, err = webtransportgo.Dial(dialCtx, endpoint, http.Header{})
	}

	if err != nil {
		logger.Error("WebTransport dial failed", "error", err)
		return nil, err
	}

	logger.Info("WebTransport connection established")

	extensions := func() *Parameters {
		if c.Config == nil || c.Config.ClientSetupExtensions == nil {
			logger.Debug("no setup extensions provided, using default parameters")
			return NewParameters()
		}

		params := c.Config.ClientSetupExtensions()
		if params == nil {
			logger.Debug("client setup extensions returned nil, using default parameters")
			params = NewParameters()
		}

		logger.Debug("setup extensions configured")
		return params
	}

	sess, err := c.openSession(conn, path, extensions, mux, logger)
	if err != nil {
		logger.Error("session establishment failed", "error", err)
		return nil, err
	}

	logger.Info("moq session over webTransport established successfully")

	return sess, nil
}

// TODO: Expose this method if QUIC is supported
func (c *Client) DialQUIC(ctx context.Context, addr, path string, mux *TrackMux) (*Session, error) {
	sessionID := generateSessionID()

	var logger *slog.Logger
	if c.Logger == nil {
		logger = slog.New(slog.DiscardHandler)
	} else {
		logger = c.Logger.With("session_id", sessionID,
			"address", addr,
			"path", path)
	}
	logger.Info("initiating QUIC MOQ session")

	if c.shuttingDown() {
		logger.Warn("QUIC dial rejected: client shutting down")
		return nil, ErrClientClosed
	}

	c.init()

	logger.Debug("starting QUIC connection establishment")

	dialTimeout := c.dialTimeout()
	dialCtx, cancelDial := context.WithTimeout(ctx, dialTimeout)
	defer cancelDial()

	var conn quic.Connection
	var err error

	if c.DialQUICConn != nil {
		logger.Debug("using custom QUIC dial function")
		conn, err = c.DialQUICConn(dialCtx, addr, c.TLSConfig, c.QUICConfig)
	} else {
		logger.Debug("using default QUIC dial")
		conn, err = quicgo.Dial(dialCtx, addr, c.TLSConfig, c.QUICConfig)
	}

	if err != nil {
		logger.Error("QUIC connection failed", "error", err)
		return nil, err
	}

	logger.Info("QUIC connection established")

	extensions := func() *Parameters {
		if c.Config == nil || c.Config.ClientSetupExtensions == nil {
			logger.Debug("no setup extensions provided, using default parameters")
			params := NewParameters()
			params.SetString(param_type_path, path)
			return params
		}

		params := c.Config.ClientSetupExtensions()
		if params == nil {
			logger.Debug("client setup extensions returned nil, using default parameters")
			params = NewParameters()
		}
		params.SetString(param_type_path, path)

		logger.Debug("setup extensions configured with path parameter")

		return params
	}

	sess, err := c.openSession(conn, path, extensions, mux, logger)
	if err != nil {
		logger.Error("QUIC session establishment failed", "error", err)
		return nil, err
	}

	return sess, nil
}

func (c *Client) openSession(conn quic.Connection, path string, extensions func() *Parameters, mux *TrackMux, connLogger *slog.Logger) (*Session, error) {
	connLogger.Debug("opening session stream")

	sessionID := generateSessionID()

	sessLogger := connLogger.With(
		"session_id", sessionID,
	)

	stream, err := conn.OpenStream()
	if err != nil {
		sessLogger.Error("failed to open session stream", "error", err)
		return nil, err
	}

	sessLogger.Debug("session stream opened", "stream_id", stream.StreamID())

	// Send STREAM_TYPE message
	stm := message.StreamTypeMessage{
		StreamType: stream_type_session,
	}
	_, err = stm.Encode(stream)
	if err != nil {
		sessLogger.Error("failed to send STREAM_TYPE message",
			"error", err,
			"stream_id", stream.StreamID(),
		)
		return nil, err
	}

	sessLogger.Debug("moq: opened session stream")

	clientParams := extensions()

	// Send a SESSION_CLIENT message
	scm := message.SessionClientMessage{
		SupportedVersions: internal.DefaultClientVersions,
		Parameters:        clientParams.paramMap,
	}
	_, err = scm.Encode(stream)
	if err != nil {
		sessLogger.Error("failed to send SESSION_CLIENT message",
			"error", err,
		)
		return nil, err
	}

	var ssm message.SessionServerMessage
	_, err = ssm.Decode(stream)
	if err != nil {
		sessLogger.Error("failed to receive SESSION_SERVER message",
			"error", err,
		)
		return nil, err
	}

	version := ssm.SelectedVersion

	serverParams := &Parameters{ssm.Parameters}

	// Create session
	sess := newSession(conn, version, path, clientParams, serverParams,
		stream, mux, sessLogger)

	c.addSession(sess)

	go func() {
		<-sess.Context().Done()
		c.removeSession(sess)
	}()

	sessLogger.Info("established a moq session")

	return sess, nil
}

func (s *Client) addSession(sess *Session) {
	s.sessMu.Lock()
	defer s.sessMu.Unlock()

	if sess == nil {
		if s.Logger != nil {
			s.Logger.Warn("attempted to add nil session")
		}
		return
	}

	s.activeSess[sess] = struct{}{}

	if s.Logger != nil {
		s.Logger.Debug("session added successfully",
			"total_active_sessions", len(s.activeSess),
		)
	}
}

func (s *Client) removeSession(sess *Session) {
	s.sessMu.Lock()
	defer s.sessMu.Unlock()

	if sess == nil {
		if s.Logger != nil {
			s.Logger.Warn("attempted to remove nil session")
		}
		return
	}

	delete(s.activeSess, sess)

	if s.Logger != nil {
		s.Logger.Debug("session removed successfully",
			"remaining_active_sessions", len(s.activeSess),
		)
	}

	// Send completion signal if connections reach zero and server is closed
	if len(s.activeSess) == 0 && s.shuttingDown() {
		if s.Logger != nil {
			s.Logger.Info("all sessions closed, sending completion signal")
		}
		select {
		case s.doneChan <- struct{}{}:
		default:
			// Channel might already be closed
		}
	}
}

func (s *Client) shuttingDown() bool {
	return s.inShutdown.Load()
}

func (c *Client) Close() error {
	start := time.Now()
	c.inShutdown.Store(true)

	if c.Logger != nil {
		c.Logger.Info("initiating client shutdown",
			"active_sessions", len(c.activeSess),
		)
	}

	sessionCount := 0
	for sess := range c.activeSess {
		sess.Terminate(NoError, NoError.String())
		sessionCount++
	}

	if c.Logger != nil {
		c.Logger.Debug("terminated all active sessions",
			"terminated_sessions", sessionCount,
		)
	}

	// Wait for active connections to complete if any
	if len(c.activeSess) > 0 {
		if c.Logger != nil {
			c.Logger.Debug("waiting for sessions to complete cleanup")
		}
		<-c.doneChan
	}

	if c.Logger != nil {
		c.Logger.Info("client shutdown completed",
			"duration", time.Since(start),
		)
	}

	return nil
}

func (c *Client) Shutdown(ctx context.Context) error {
	c.init() // Ensure initialization
	c.inShutdown.Store(true)
	logger := c.Logger
	if logger != nil {
		logger.Info("shutting down client gracefully",
			"active_sessions", len(c.activeSess),
		)
	}

	// Go away all active sessions
	for sess := range c.activeSess {
		c.goAway(sess)
	}

	if logger != nil {
		logger.Debug("sent go-away to all active sessions")
	}

	// For active connections, wait for completion or context cancellation
	if len(c.activeSess) > 0 {
		if logger != nil {
			logger.Debug("waiting for graceful session termination")
		}

		select {
		case <-c.doneChan:
			if logger != nil {
				logger.Info("all sessions terminated gracefully")
			}
		case <-ctx.Done():
			for sess := range c.activeSess {
				go sess.Terminate(GoAwayTimeoutErrorCode, GoAwayTimeoutErrorCode.String())
				c.removeSession(sess)
			}
			if logger != nil {
				logger.Warn("graceful shutdown timeout, forcing session termination",
					"context_error", ctx.Err(),
				)
			}
			return ctx.Err()
		}
	}

	if logger != nil {
		logger.Info("graceful client shutdown completed")
	}

	return nil
}

func (c *Client) goAway(sess *Session) {
	// TODO: Implement actual go-away logic
}
