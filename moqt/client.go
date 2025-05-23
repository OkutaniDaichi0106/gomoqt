package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
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
		if c.Logger == nil {
			c.Logger = slog.Default()
		}

		c.Logger.Debug("Client initialized")
	})
}

func (c *Client) Dial(ctx context.Context, urlStr string, mux *TrackMux) (*Session, *SetupResponse, error) {
	if c.shuttingDown() {
		return nil, nil, ErrClientClosed
	}
	c.init()

	c.Logger.Info("dialing server", "url", urlStr) // Changed to Info for better visibility
	req, err := NewSetupRequest(urlStr)
	if err != nil {
		c.Logger.Error("failed to create setup request", "error", err.Error(), "url", urlStr)
		return nil, nil, err
	}

	// Dial based on the scheme
	switch req.uri.Scheme {
	case "https":
		c.Logger.Debug("dialing using WebTransport")
		return c.DialWebTransport(ctx, req, mux)
	case "moqt":
		c.Logger.Debug("dialing using QUIC")
		return c.DialQUIC(ctx, req, mux)
	default:
		err = errors.New("invalid scheme")
		c.Logger.Error("unsupported URL scheme", "scheme", req.uri.Scheme, "url", urlStr)
		return nil, nil, err
	}
}

func (c *Client) DialWebTransport(ctx context.Context, req *SetupRequest, mux *TrackMux) (*Session, *SetupResponse, error) {
	if c.shuttingDown() {
		return nil, nil, ErrClientClosed
	}
	c.init()

	if req.uri.Scheme != "https" {
		err := errors.New("invalid scheme")
		c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		return nil, nil, err
	}

	c.Logger.Debug("dialing WebTransport", "endpoint", "https://"+req.uri.Hostname()+":"+req.uri.Port()+req.uri.Path)

	if DialWebtransportFunc == nil {
		DialWebtransportFunc = defaultDialWTFunc
	}

	_, conn, err := DialWebtransportFunc(ctx, req.uri.String(), http.Header{})
	if err != nil {
		c.Logger.Error("failed to dial WebTransport", "error", err.Error(), "endpoint", "https://"+req.uri.Hostname()+":"+req.uri.Port()+req.uri.Path)
		return nil, nil, err
	}

	// Open a session stream
	var params message.Parameters
	if c.SetupExtensions != nil {
		params = c.SetupExtensions.paramMap
		c.Logger.Debug("using setup extensions", "extensions", params)
	} else {
		c.Logger.Debug("no setup extensions provided")
	}

	sess, rsp, err := c.openSession(conn, &Parameters{params}, mux)
	if err != nil {
		c.Logger.Error("failed to open session stream", "error", err.Error())
		return nil, nil, err
	}

	c.Logger.Info("setup response received", "version", rsp.selectedVersion, "parameters", rsp.Parameters)
	return sess, rsp, nil
}

// TODO: Expose this method if QUIC is supported
func (c *Client) DialQUIC(ctx context.Context, req *SetupRequest, mux *TrackMux) (*Session, *SetupResponse, error) {
	if c.shuttingDown() {
		return nil, nil, ErrClientClosed
	}

	c.init()

	c.Logger.Debug("dialing using QUIC")

	if req.uri.Scheme != "moqt" {
		err := errors.New("invalid scheme")
		c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		return nil, nil, err
	}

	c.Logger.Debug("dialing QUIC", "host", req.uri.Hostname(), "port", req.uri.Port(), "path", req.uri.Path)

	if DialQUICFunc == nil {
		panic("no DialQUICFunc provided")
	}

	// Dial QUIC connection
	c.Logger.Debug("dialing QUIC connection", "address", req.uri.Hostname()+":"+req.uri.Port())
	conn, err := DialQUICFunc(ctx, req.uri.Hostname()+":"+req.uri.Port(), c.TLSConfig, c.QUICConfig)
	if err != nil {
		c.Logger.Error("failed to dial QUIC connection", "error", err.Error(), "address", req.uri.Hostname()+":"+req.uri.Port())
		return nil, nil, err
	}

	//
	var param *Parameters
	if c.SetupExtensions != nil {
		param = c.SetupExtensions
		c.Logger.Debug("using setup extensions", "extensions", param)
	} else {
		c.Logger.Debug("no setup extensions provided")
		param = NewParameters()
	}
	param.SetString(param_type_path, req.uri.Path)

	sess, rsp, err := c.openSession(conn, param, mux)
	if err != nil {
		c.Logger.Error("failed to open session stream", "error", err.Error())
		return nil, nil, err
	}

	c.Logger.Debug("setup response received", "version", rsp.selectedVersion, "parameters", rsp.Parameters)

	return sess, rsp, nil
}

func (c *Client) openSession(conn quic.Connection, params *Parameters, mux *TrackMux) (*Session, *SetupResponse, error) {
	sess := newSession(conn)

	err := sess.openSessionStream(internal.DefaultClientVersions, params)
	if err != nil {
		c.Logger.Error("failed to open a session stream", "error", err.Error())
		return nil, nil, err
	}

	c.Logger.Debug("session stream opened")

	if mux != nil {
		mux = DefaultMux
	}

	go sess.handleAnnounceStream(mux)
	go sess.handleSubscribeStream(mux)
	go sess.handleInfoStream(mux)

	c.addSession(sess)
	go func() {
		<-sess.Context().Done()
		c.removeSession(sess)
	}()

	return sess, &SetupResponse{
		selectedVersion: sess.sessionStream.selectedVersion,
		Parameters:      sess.sessionStream.serverParameters,
	}, nil
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
