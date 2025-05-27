package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
	quicgo "github.com/quic-go/quic-go"
)

type Client struct {
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

	/***/
	SetupExtensions *Parameters

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
		if c.Logger != nil {
			c.Logger.Debug("Client initialized")
		}
	})
}

func (c *Client) timeout() time.Duration {
	if c.Config.Timeout != 0 {
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
		c.Logger.Info("dialing server", "url", urlStr) // Changed to Info for better visibility
	}

	req, err := NewSetupRequest(urlStr)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("failed to create setup request", "error", err.Error(), "url", urlStr)
		}
		return nil, err
	}

	// Dial based on the scheme
	switch req.uri.Scheme {
	case "https":
		// c.Logger.Debug("dialing using WebTransport")
		return c.DialWebTransport(req, mux)
	case "moqt":
		// c.Logger.Debug("dialing using QUIC")
		return c.DialQUIC(req, mux)
	default:
		if c.Logger != nil {
			c.Logger.Error("unsupported URL scheme",
				"scheme", req.uri.Scheme,
				"url", urlStr,
			)
		}
		return nil, ErrInvalidScheme
	}
}

func (c *Client) DialWebTransport(req *SetupRequest, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		return nil, ErrClientClosed
	}
	c.init()

	if req.uri.Scheme != "https" {
		err := errors.New("invalid scheme")
		if c.Logger != nil {
			c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		}
		return nil, err
	}

	if c.Logger != nil {
		c.Logger.Info("dialing WebTransport", "endpoint", "https://"+req.uri.Hostname()+":"+req.uri.Port()+req.uri.Path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout())
	defer cancel()
	_, conn, err := webtransport.DialWebtransportFunc(ctx, req.uri.String(), http.Header{})
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("failed to dial WebTransport", "error", err.Error(), "endpoint", "https://"+req.uri.Hostname()+":"+req.uri.Port()+req.uri.Path)
		}
		return nil, err
	}

	// Open a session stream
	var params message.Parameters
	if c.SetupExtensions != nil {
		params = c.SetupExtensions.paramMap
		if c.Logger != nil {
			c.Logger.Debug("using setup extensions", "extensions", params)
		}
	} else {
		if c.Logger != nil {
			c.Logger.Debug("no setup extensions provided")
		}
	}

	sess, err := c.openSession(conn, &Parameters{params}, mux)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("failed to open session stream", "error", err.Error())
		}
		return nil, err
	}

	return sess, nil
}

// TODO: Expose this method if QUIC is supported
func (c *Client) DialQUIC(req *SetupRequest, mux *TrackMux) (*Session, error) {
	if c.shuttingDown() {
		if c.Logger != nil {
			c.Logger.Error("client is shutting down")
		}
		return nil, ErrClientClosed
	}

	c.init()

	if c.Logger != nil {
		c.Logger.Debug("dialing MOQ",
			"scheme", req.uri.Scheme,
			"host", req.uri.Hostname(),
			"port", req.uri.Port(),
			"path", req.uri.Path,
		)
	}

	if req.uri.Scheme != "moqt" {
		if c.Logger != nil {
			c.Logger.Error("unsupported url scheme",
				"scheme", req.uri.Scheme,
				"url", req.uri.String(),
			)
		}
		return nil, ErrInvalidScheme
	}

	if quic.DialFunc == nil {
		panic("no DialQUICFunc provided")
	}

	// Dial QUIC connection
	if c.Logger != nil {
		c.Logger.Debug("dialing QUIC connection",
			"address", req.uri.Hostname()+":"+req.uri.Port(),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout())
	defer cancel()
	conn, err := quic.DialFunc(ctx, req.uri.Hostname()+":"+req.uri.Port(), c.TLSConfig, c.QUICConfig)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("failed to dial QUIC connection",
				"error", err,
				"address", req.uri.Hostname()+":"+req.uri.Port(),
			)
		}
		return nil, err
	}

	//
	var param *Parameters
	if c.SetupExtensions != nil {
		param = c.SetupExtensions
		if c.Logger != nil {
			c.Logger.Debug("using setup extensions", "extensions", param)
		}
	} else {
		if c.Logger != nil {
			c.Logger.Debug("no setup extensions provided")
		}
		param = NewParameters()
	}
	param.SetString(param_type_path, req.uri.Path)

	sess, err := c.openSession(conn, param, mux)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("failed to open session stream", "error", err.Error())
		}
		return nil, err
	}

	return sess, nil
}

func (c *Client) openSession(conn quic.Connection, params *Parameters, mux *TrackMux) (*Session, error) {
	sessCtx := newSessionContext(conn.Context(), c.Logger)
	sess := newSession(sessCtx, conn, mux)

	err := sess.openSessionStream(internal.DefaultClientVersions, params)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("failed to open a session stream", "error", err.Error())
		}
		return nil, err
	}

	if c.Logger != nil {
		c.Logger.Debug("session stream opened")
	}

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
			return nil
		case <-ctx.Done():
			for sess := range c.activeSess {
				(*sess).Terminate(ErrGoAwayTimeout)
			}
			return ctx.Err()
		}
	}

	return nil
}

func (c *Client) goAway(sess *Session) {}
