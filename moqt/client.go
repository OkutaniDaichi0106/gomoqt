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

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/quic/quicgo"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport/webtransportgo"
)

// Client is a MOQ client that can establish sessions with MOQ servers.
// It supports both WebTransport (for browser compatibility) and raw QUIC connections.
//
// A Client can dial multiple servers and maintain multiple active sessions.
// Sessions are tracked and managed automatically. When the client shuts down,
// all active sessions are terminated gracefully.
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

	/*
	 * Dial QUIC function
	 */
	DialQUICFunc quic.DialAddrFunc

	/*
	 * Dial WebTransport function
	 */
	DialWebTransportFunc webtransport.DialAddrFunc

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
	var logger *slog.Logger
	if c.Logger != nil {
		logger = c.Logger
	} else {
		logger = slog.New(slog.DiscardHandler)
	}

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
		return c.DialWebTransport(ctx, parsedURL.Hostname()+":"+parsedURL.Port(), parsedURL.Path, mux)
	case "moqt":
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
	var clientLogger *slog.Logger
	if c.Logger != nil {
		clientLogger = c.Logger.With(
			"host", host,
		)
	} else {
		clientLogger = slog.New(slog.DiscardHandler)

	}

	if c.shuttingDown() {
		clientLogger.Warn("WebTransport dial rejected: client shutting down")
		return nil, ErrClientClosed
	}

	c.init()

	dialTimeout := c.dialTimeout()
	dialCtx, cancelDial := context.WithTimeout(ctx, dialTimeout)
	defer cancelDial()

	clientLogger.Debug("dialing WebTransport")

	var conn quic.Connection
	var err error

	if c.DialWebTransportFunc != nil {
		_, conn, err = c.DialWebTransportFunc(dialCtx, host+path, http.Header{})
	} else {
		_, conn, err = webtransportgo.Dial(dialCtx, "https://"+host+path, http.Header{})
	}

	if err != nil {
		clientLogger.Error("WebTransport dial failed", "error", err)
		return nil, err
	}

	connLogger := clientLogger.With(
		"transport", "webtransport",
		"local_address", conn.LocalAddr(),
		"remote_address", conn.RemoteAddr(),
		"quic_version", conn.ConnectionState().Version,
		"alpn", conn.ConnectionState().TLS.NegotiatedProtocol,
	)

	connLogger.Info("WebTransport connection established")

	sessStream, err := openSessionStream(conn, path, webTransportExtensions(), connLogger)
	if err != nil {
		connLogger.Error("session establishment failed", "error", err)
		return nil, err
	}

	var sess *Session
	sess = newSession(conn, sessStream, mux, connLogger, func() { c.removeSession(sess) })
	c.addSession(sess)

	connLogger.Info("moq: established a new session over WebTransport successfully")

	return sess, nil
}

// TODO: Expose this method if QUIC is supported
func (c *Client) DialQUIC(ctx context.Context, addr, path string, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		return nil, ErrClientClosed
	}

	c.init()

	var clientLogger *slog.Logger
	if c.Logger == nil {
		clientLogger = slog.New(slog.DiscardHandler)
	} else {
		clientLogger = c.Logger
	}

	dialTimeout := c.dialTimeout()
	dialCtx, cancelDial := context.WithTimeout(ctx, dialTimeout)
	defer cancelDial()

	var conn quic.Connection
	var err error

	if c.DialQUICFunc != nil {
		clientLogger.Debug("using custom QUIC dial function")
		conn, err = c.DialQUICFunc(dialCtx, addr, c.TLSConfig, c.QUICConfig)
	} else {
		clientLogger.Debug("using default QUIC dial function")
		conn, err = quicgo.DialAddrEarly(dialCtx, addr, c.TLSConfig, c.QUICConfig)
	}

	if err != nil {
		clientLogger.Error("QUIC connection failed", "error", err)
		return nil, err
	}

	connLogger := clientLogger.With(
		"transport", "quic",
		"local_address", conn.LocalAddr(),
		"remote_address", conn.RemoteAddr(),
		"quic_version", conn.ConnectionState().Version,
		"alpn", conn.ConnectionState().TLS.NegotiatedProtocol,
	)
	// TODO: Add connection ID

	connLogger.Info("QUIC connection established")

	sessStream, err := openSessionStream(conn, path, quicExtensions(path), connLogger)
	if err != nil {
		connLogger.Error("failed to open session stream", "error", err)
		return nil, err
	}

	var sess *Session
	sess = newSession(conn, sessStream, mux, connLogger, func() { c.removeSession(sess) })
	c.addSession(sess)

	return sess, nil
}

func quicExtensions(path string) *Parameters {
	params := NewParameters()

	params.SetString(param_type_path, path)

	return params
}

func webTransportExtensions() *Parameters {
	params := NewParameters()

	return params
}

func openSessionStream(conn quic.Connection, path string, extensions *Parameters, connLogger *slog.Logger) (*sessionStream, error) {
	connLogger.Debug("moq: opening session stream")

	stream, err := conn.OpenStream()
	if err != nil {
		connLogger.Error("moq: failed to open session stream", "error", err)
		return nil, err
	}

	streamLogger := connLogger.With(
		"stream_id", stream.StreamID(),
	)

	// Send STREAM_TYPE message
	err = message.StreamTypeSession.Encode(stream)
	if err != nil {
		streamLogger.Error("moq: failed to send STREAM_TYPE message",
			"error", err,
		)
		return nil, err
	}

	streamLogger.Debug("moq: opened session stream")

	versions := DefaultClientVersions

	// Send a SESSION_CLIENT message
	scm := message.SessionClientMessage{
		SupportedVersions: versions,
		Parameters:        extensions.paramMap,
	}
	err = scm.Encode(stream)
	if err != nil {
		streamLogger.Error("moq: failed to send SESSION_CLIENT message",
			"error", err,
		)
		return nil, err
	}

	req := &SetupRequest{
		Path:             path,
		Versions:         versions,
		ClientExtensions: extensions,
		ctx:              stream.Context(),
	}

	sessStr := newSessionStream(stream, req)

	rsp := newResponse(sessStr)

	err = rsp.AwaitAccepted()
	if err != nil {
		streamLogger.Error("moq: failed to set up session", "error", err)
		conn.CloseWithError(quic.ApplicationErrorCode(InternalSessionErrorCode), "moq: failed to set up session")
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
		c.Logger.Info("session added successfully",
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

	c.goAway()

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

func (c *Client) goAway() {
	for sess := range c.activeSess {
		if sess == nil {
			continue
		}
		sess.goAway("")
	}
}
