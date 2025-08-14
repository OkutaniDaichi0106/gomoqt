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

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/quic/quicgo"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport/webtransportgo"
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

	// mu         sync.Mutex
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
	if c.Logger != nil {
		logger = c.Logger.With("session_id", sessionID, "endpoint", endpoint)
	} else {
		logger = slog.New(slog.DiscardHandler)

	}

	if c.shuttingDown() {
		logger.Warn("WebTransport dial rejected: client shutting down")
		return nil, ErrClientClosed
	}
	c.init()

	dialTimeout := c.dialTimeout()
	dialCtx, cancelDial := context.WithTimeout(ctx, dialTimeout)
	defer cancelDial()

	logger.Debug("dialing WebTransport", "timeout", dialTimeout)

	var conn quic.Connection
	var err error

	if c.DialWebTransportFunc != nil {
		_, conn, err = c.DialWebTransportFunc(dialCtx, host+path, http.Header{})
	} else {
		_, conn, err = webtransportgo.Dial(dialCtx, endpoint, http.Header{})
	}

	if err != nil {
		logger.Error("WebTransport dial failed", "error", err)
		return nil, err
	}

	logger.Info("WebTransport connection established")

	sessStream, err := openSessionStream(conn, path, c.webTransportExtensions(), logger)
	if err != nil {
		logger.Error("session establishment failed", "error", err)
		return nil, err
	}

	var sess *Session
	sess = newSession(conn, sessStream, mux, logger, func() { c.removeSession(sess) })
	c.addSession(sess)

	logger.Info("moq: established a new session over WebTransport successfully")

	return sess, nil
}

// TODO: Expose this method if QUIC is supported
func (c *Client) DialQUIC(ctx context.Context, addr, path string, mux *TrackMux) (*Session, error) {
	sessionID := generateSessionID()

	if c.shuttingDown() {
		return nil, ErrClientClosed
	}

	c.init()

	var logger *slog.Logger
	if c.Logger == nil {
		logger = slog.New(slog.DiscardHandler)
	} else {
		logger = c.Logger.With("session_id", sessionID,
			"address", addr,
			"path", path)
	}

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
		logger.Debug("using default QUIC dial function")
		conn, err = quicgo.DialAddrEarly(dialCtx, addr, c.TLSConfig, c.QUICConfig)
	}

	if err != nil {
		logger.Error("QUIC connection failed", "error", err)
		return nil, err
	}

	logger.Info("QUIC connection established")

	sessStream, err := openSessionStream(conn, path, c.nativeQUICExtensions(path), logger)
	if err != nil {
		logger.Error("failed to open session stream", "error", err)
		return nil, err
	}

	var sess *Session
	sess = newSession(conn, sessStream, mux, logger, func() { c.removeSession(sess) })
	c.addSession(sess)

	logger.Info("moq session over QUIC established successfully")

	return sess, nil
}

func (c *Client) nativeQUICExtensions(path string) *Parameters {
	if c.Config == nil || c.Config.ClientSetupExtensions == nil {
		params := NewParameters()
		params.SetString(param_type_path, path)
		return params
	}

	params := c.Config.ClientSetupExtensions()
	if params == nil {
		params = NewParameters()
	}

	params.SetString(param_type_path, path)

	return params
}

func (c *Client) webTransportExtensions() *Parameters {
	if c.Config == nil || c.Config.ClientSetupExtensions == nil {
		return NewParameters()
	}

	params := c.Config.ClientSetupExtensions()
	if params == nil {
		params = NewParameters()
	}

	return params
}

func openSessionStream(conn quic.Connection, path string, extensions *Parameters, logger *slog.Logger) (*sessionStream, error) {
	connLogger := logger.With("transport", "quic", "path", path)
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
	err = message.StreamTypeSession.Encode(stream)
	if err != nil {
		sessLogger.Error("failed to send STREAM_TYPE message",
			"error", err,
			"stream_id", stream.StreamID(),
		)
		return nil, err
	}

	sessLogger.Debug("moq: opened session stream")

	versions := DefaultClientVersions

	// Send a SESSION_CLIENT message
	scm := message.SessionClientMessage{
		SupportedVersions: versions,
		Parameters:        extensions.paramMap,
	}
	err = scm.Encode(stream)
	if err != nil {
		sessLogger.Error("failed to send SESSION_CLIENT message",
			"error", err,
		)
		return nil, err
	}

	req := &Request{
		Path:       path,
		Versions:   versions,
		Extensions: extensions,
		ctx:        stream.Context(),
	}

	sessStr := newSessionStream(stream, req)

	rsp := &response{
		sessionStream: sessStr,
	}

	err = rsp.AwaitAccepted()
	if err != nil {
		sessLogger.Error("session acceptance failed", "error", err)
		conn.CloseWithError(quic.ApplicationErrorCode(InternalSessionErrorCode), "session acceptance failed")
		return nil, err
	}

	return rsp.sessionStream, nil
}

func (c *Client) addSession(sess *Session) {
	c.sessMu.Lock()
	defer c.sessMu.Unlock()

	if sess == nil {
		if c.Logger != nil {
			c.Logger.Warn("attempted to add nil session")
		}
		return
	}

	c.activeSess[sess] = struct{}{}

	if c.Logger != nil {
		c.Logger.Debug("session added successfully",
			"total_active_sessions", len(c.activeSess),
		)
	}
}

func (c *Client) removeSession(sess *Session) {
	c.sessMu.Lock()
	defer c.sessMu.Unlock()

	_, ok := c.activeSess[sess]
	if !ok {
		return
	}

	delete(c.activeSess, sess)
	// Send completion signal if connections reach zero and server is closed
	if len(c.activeSess) == 0 && c.shuttingDown() {
		select {
		case <-c.doneChan:
			// Already closed
		default:
			close(c.doneChan)
		}
	}
}

func (c *Client) shuttingDown() bool {
	return c.inShutdown.Load()
}

func (c *Client) Close() error {
	c.inShutdown.Store(true)

	if c.Logger != nil {
		c.Logger.Info("initiating client shutdown")
	}

	c.sessMu.Lock()
	for sess := range c.activeSess {
		go sess.Terminate(NoError, NoError.String())
	}
	c.sessMu.Unlock()

	if c.Logger != nil {
		c.Logger.Debug("terminated all active sessions")
	}

	// Wait for active connections to complete if any
	if len(c.activeSess) > 0 {
		<-c.doneChan
	}

	if c.Logger != nil {
		c.Logger.Info("client shutdown completed")
	}

	return nil
}

func (c *Client) Shutdown(ctx context.Context) error {
	if c.shuttingDown() {
		return nil
	}

	c.inShutdown.Store(true)

	logger := c.Logger
	if logger != nil {
		logger.Info("shutting down client gracefully")
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
		select {
		case <-c.doneChan:
		case <-ctx.Done():
			if len(c.activeSess) > 0 {
				for sess := range c.activeSess {
					go sess.Terminate(GoAwayTimeoutErrorCode, GoAwayTimeoutErrorCode.String())
				}

				if logger != nil {
					logger.Warn("graceful shutdown timeout, forcing session termination",
						"context_error", ctx.Err(),
					)
				}
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
