package moqt

import (
	"context"
	"crypto/tls"
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
			c.Logger.Debug("client initialized")
		}
	})
}

func (c *Client) timeout() time.Duration {
	if c.Config != nil && c.Config.SetupTimeout != 0 {
		return c.Config.SetupTimeout
	}
	return 5 * time.Second // TODO: Consider appropriate timeout
}

func (c *Client) Dial(ctx context.Context, urlStr string, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		return nil, ErrClientClosed
	}
	c.init()

	if c.Logger != nil {
		c.Logger.Debug("dialing server", "url", urlStr)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("failed to parse URL", "error", err.Error(), "url", urlStr)
		}
		return nil, err
	}

	// Dial based on the scheme
	switch parsedURL.Scheme {
	case "https":
		return c.DialWebTransport(ctx, parsedURL.Hostname()+":"+parsedURL.Port(), parsedURL.Path, mux)
	case "moqt":
		return c.DialQUIC(ctx, parsedURL.Hostname()+":"+parsedURL.Port(), parsedURL.Path, mux)
	default:
		if c.Logger != nil {
			c.Logger.Error("unsupported URL scheme",
				"scheme", parsedURL.Scheme,
				"url", urlStr,
			)
		}
		return nil, ErrInvalidScheme
	}
}

func (c *Client) DialWebTransport(ctx context.Context, host, path string, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		return nil, ErrClientClosed
	}
	c.init()

	var logger *slog.Logger
	if c.Logger != nil {
		logger = c.Logger.With(
			"endpoint", "https://"+host+path,
		)
		logger.Info("dialing webtransport")
	}

	dialCtx, cancelDial := context.WithTimeout(ctx, c.timeout())
	defer cancelDial()
	var conn quic.Connection
	var err error
	if c.DialWebTransportFunc != nil {
		_, conn, err = c.DialWebTransportFunc(dialCtx, host+path, http.Header{})
	} else {
		_, conn, err = webtransportgo.Dial(dialCtx, "https://"+host+path, http.Header{})
	}
	if err != nil {
		if logger != nil {
			logger.Error("failed to dial webtransport",
				"error", err.Error(),
			)
		}
		return nil, err
	}

	extensions := func() *Parameters {
		if c.Config == nil || c.Config.ClientSetupExtensions == nil {
			if logger != nil {
				logger.Debug("no setup extensions provided, using default parameters")
			}
			return NewParameters()
		}

		params := c.Config.ClientSetupExtensions()
		if params == nil {
			if logger != nil {
				logger.Debug("client setup extensions returned nil, using default parameters")
			}
			params = NewParameters()
		}
		return params
	}

	sess, err := c.openSession(conn, path, extensions, mux)
	if err != nil {
		if logger != nil {
			logger.Error("failed to open session stream",
				"error", err.Error(),
			)
		}
		return nil, err
	}

	return sess, nil
}

// TODO: Expose this method if QUIC is supported
func (c *Client) DialQUIC(ctx context.Context, addr, path string, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		if c.Logger != nil {
			c.Logger.Error("client is shutting down")
		}
		return nil, ErrClientClosed
	}

	c.init()

	var logger *slog.Logger
	if c.Logger != nil {
		logger = c.Logger.With(
			"scheme", "moqt",
			"host", addr,
			"path", path,
		)

		logger.Debug("dialing MOQ session")
	}

	// Dial QUIC connection
	if logger != nil {
		logger.Debug("dialing QUIC connection")
	}

	dialCtx, cancelDial := context.WithTimeout(ctx, c.timeout())
	defer cancelDial()
	var conn quic.Connection
	var err error
	if c.DialQUICConn != nil {
		conn, err = c.DialQUICConn(dialCtx, addr, c.TLSConfig, c.QUICConfig)
	} else {
		conn, err = quicgo.Dial(dialCtx, addr, c.TLSConfig, c.QUICConfig)
	}
	if err != nil {
		if logger != nil {
			logger.Error("failed to dial QUIC connection",
				"error", err,
			)
		}
		return nil, err
	}

	//
	extensions := func() *Parameters {
		if c.Config == nil || c.Config.ClientSetupExtensions == nil {
			if logger != nil {
				logger.Debug("no setup extensions provided, using default parameters")
			}
			params := NewParameters()
			params.SetString(param_type_path, path)
			return params
		}

		params := c.Config.ClientSetupExtensions()
		if params == nil {
			if logger != nil {
				logger.Debug("client setup extensions returned nil, using default parameters")
			}
			params = NewParameters()
		}
		params.SetString(param_type_path, path)
		return params
	}

	return c.openSession(conn, path, extensions, mux)
}

func (c *Client) openSession(conn quic.Connection, path string, extensions func() *Parameters, mux *TrackMux) (*Session, error) {
	//
	logger := c.Logger

	// Close the session stream channel

	stream, err := conn.OpenStream()
	if err != nil {
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_session,
	}
	_, err = stm.Encode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to send a STREAM_TYPE message", "error", err)
		}

		return nil, err
	}

	clientParams := extensions()

	// Send a SESSION_CLIENT message
	scm := message.SessionClientMessage{
		SupportedVersions: internal.DefaultClientVersions,
		Parameters:        clientParams.paramMap,
	}
	_, err = scm.Encode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to send a SESSION_CLIENT message", "error", err)
		}
		return nil, err
	}

	// Receive a set-up response
	var ssm message.SessionServerMessage
	_, err = ssm.Decode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to receive a SESSION_SERVER message", "error", err)
		}

		return nil, err
	}

	version := ssm.SelectedVersion
	serverParams := &Parameters{ssm.Parameters}

	// Set the selected version and parameters

	sess := newSession(conn, version, path, clientParams, serverParams,
		stream, mux, logger)

	c.addSession(sess)
	go func() {
		<-sess.Context().Done()
		c.removeSession(sess)
	}()

	return sess, nil
}

func (s *Client) addSession(sess *Session) {
	s.sessMu.Lock()
	defer s.sessMu.Unlock()

	if sess == nil {
		return
	}
	s.activeSess[sess] = struct{}{}
}

func (s *Client) removeSession(sess *Session) {
	s.sessMu.Lock()
	defer s.sessMu.Unlock()

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

func (s *Client) shuttingDown() bool {
	return s.inShutdown.Load()
}

func (c *Client) Close() error {
	c.inShutdown.Store(true)

	for sess := range c.activeSess {
		sess.Terminate(NoError, NoError.String())
	}

	// Wait for active connections to complete if any
	if len(c.activeSess) > 0 {
		<-c.doneChan
	}

	return nil
}

func (c *Client) Shutdown(ctx context.Context) error {
	c.init() // Ensure initialization
	c.inShutdown.Store(true)

	// Go away all active sessions
	for sess := range c.activeSess {
		c.goAway(sess)
	}

	// For active connections, wait for completion or context cancellation
	if len(c.activeSess) > 0 {
		select {
		case <-c.doneChan:
		case <-ctx.Done():
			for sess := range c.activeSess {
				go sess.Terminate(GoAwayTimeoutErrorCode, GoAwayTimeoutErrorCode.String())
				c.removeSession(sess)
			}
			return ctx.Err()
		}
	}

	return nil
}

func (c *Client) goAway(sess *Session) {
	if sess == nil {
		return
	}
}
