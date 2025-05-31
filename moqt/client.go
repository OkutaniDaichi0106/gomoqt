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
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
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

	/*
	 * Logger
	 */
	Logger *slog.Logger

	//
	initOnce sync.Once

	mu         sync.RWMutex
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
	if c.Config != nil && c.Config.Timeout != 0 {
		return c.Config.Timeout
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
		return c.DialWebTransport(ctx, parsedURL, mux)
	case "moqt":
		return c.DialQUIC(ctx, parsedURL, mux)
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

func (c *Client) DialWebTransport(ctx context.Context, uri *url.URL, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		return nil, ErrClientClosed
	}
	c.init()

	if uri.Scheme != "https" {
		if c.Logger != nil {
			c.Logger.Error("unsupported url scheme",
				"scheme", uri.Scheme,
				"url", uri.String(),
			)
		}
		return nil, ErrInvalidScheme
	}

	var logger *slog.Logger
	if c.Logger != nil {
		logger = c.Logger.With(
			"endpoint", "https://"+uri.Hostname()+":"+uri.Port()+uri.Path,
		)
		logger.Info("dialing webtransport")
	}

	dialCtx, cancelDial := context.WithTimeout(ctx, c.timeout())
	defer cancelDial()
	_, conn, err := webtransport.DialWebtransportFunc(dialCtx, uri.String(), http.Header{})
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

	sess, err := c.openSession(conn, extensions, mux)
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
func (c *Client) DialQUIC(ctx context.Context, uri *url.URL, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		if c.Logger != nil {
			c.Logger.Error("client is shutting down")
		}
		return nil, ErrClientClosed
	}

	c.init()

	path := uri.Path

	var logger *slog.Logger
	if c.Logger != nil {
		logger = c.Logger.With(
			"scheme", uri.Scheme,
			"host", uri.Hostname(),
			"port", uri.Port(),
			"path", path,
		)

		logger.Debug("dialing MOQ session")
	}

	if uri.Scheme != "moqt" {
		if logger != nil {
			logger.Error("unsupported url scheme")
		}
		return nil, ErrInvalidScheme
	}

	if quic.DialFunc == nil {
		panic("no quic.DialFunc provided")
	}

	// Dial QUIC connection
	if logger != nil {
		logger.Debug("dialing QUIC connection")
	}

	dialCtx, cancelDial := context.WithTimeout(ctx, c.timeout())
	defer cancelDial()
	conn, err := quic.DialFunc(dialCtx, uri.Hostname()+":"+uri.Port(), c.TLSConfig, c.QUICConfig)
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

	return c.openSession(conn, extensions, mux)
}

func (c *Client) openSession(conn quic.Connection, extensions func() *Parameters, mux *TrackMux) (*Session, error) {
	//
	logger := c.Logger

	//
	var sessTracer *moqtrace.SessionTracer
	if c.Config != nil && c.Config.Tracer != nil {
		sessTracer = c.Config.Tracer()
	} else {
		sessTracer = moqtrace.DefaultSessionTracer()
	}

	// Close the session stream channel

	stream, err := conn.OpenStream()
	if err != nil {
		return nil, err
	}

	var streamTracer *moqtrace.StreamTracer
	if sessTracer.QUICStreamOpened != nil {
		streamTracer = sessTracer.QUICStreamOpened(stream.StreamID())
		moqtrace.InitStreamTracer(streamTracer)
	} else {
		streamTracer = moqtrace.DefaultQUICStreamOpened(stream.StreamID())
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
	streamTracer.StreamTypeMessageSent(stm)

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
	streamTracer.SessionClientMessageSent(scm)

	// Receive a set-up response
	var ssm message.SessionServerMessage
	_, err = ssm.Decode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to receive a SESSION_SERVER message", "error", err)
		}
		return nil, err
	}
	streamTracer.SessionServerMessageReceived(ssm)

	version := ssm.SelectedVersion
	serverParams := &Parameters{ssm.Parameters}
	path, err := serverParams.GetString(param_type_path)
	if err != nil {
		if logger != nil {
			logger.Error("failed to get path parameter from server parameters", "error", err)
		}
		return nil, err
	}

	sessCtx := newSessionContext(conn.Context(), version, path, clientParams, serverParams, logger, sessTracer)

	// Set the selected version and parameters
	sessstr := newSessionStream(
		sessCtx,
		stream,
		streamTracer,
	)

	if logger != nil {
		logger.Debug("opened a session stream")
	}

	sess := newSession(sessCtx, sessstr, conn, mux)

	c.addSession(sess)
	go func() {
		<-sessCtx.Done()
		c.removeSession(sess)
	}()

	return sess, nil
}

func (s *Client) addSession(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess == nil {
		return
	}
	s.activeSess[sess] = struct{}{}
}

func (s *Client) removeSession(sess *Session) {
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

func (s *Client) shuttingDown() bool {
	return s.inShutdown.Load()
}

func (c *Client) Close() error {
	c.inShutdown.Store(true)

	c.mu.Lock()
	defer c.mu.Unlock()

	for sess := range c.activeSess {
		(*sess).Terminate(NoErrTerminate)
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

	c.mu.Lock()
	defer c.mu.Unlock()

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
				go sess.Terminate(ErrGoAwayTimeout)
				c.removeSession(sess)
			}
			return ctx.Err() // Return context error
		}
	}

	return nil
}

func (c *Client) goAway(sess *Session) {
	if sess == nil {
		return
	}
}
