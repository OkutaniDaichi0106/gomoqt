package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
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
	TrackResolver TrackHandler

	/***/
	SetupExtensions *Parameters

	/*
	 * Logger
	 */
	Logger *slog.Logger

	once sync.Once
}

func (c *Client) init() {
	c.once.Do(func() {
		if c.Logger == nil {
			c.Logger = slog.Default()
		}

		c.Logger.Debug("Client initialized")
	})
}

func (c *Client) Dial(urlStr string, ctx context.Context) (Session, *SetupResponse, error) {
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
		return c.DialWebTransport(ctx, *req)
	case "moqt":
		c.Logger.Debug("dialing using QUIC")
		return c.dialQUIC(ctx, *req)
	default:
		err = errors.New("invalid scheme")
		c.Logger.Error("unsupported URL scheme", "scheme", req.uri.Scheme, "url", urlStr)
		return nil, nil, err
	}
}

func (c *Client) DialWebTransport(ctx context.Context, req SetupRequest) (Session, *SetupResponse, error) {
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

	sess, rsp, err := OpenSession(conn, &Parameters{params}, c.TrackResolver)
	if err != nil {
		c.Logger.Error("failed to open session stream", "error", err.Error())
		return nil, nil, err
	}

	c.Logger.Info("setup response received", "version", rsp.selectedVersion, "parameters", rsp.Parameters)
	return sess, rsp, nil
}

// TODO: Expose this method if QUIC is supported
func (c *Client) dialQUIC(ctx context.Context, req SetupRequest) (Session, *SetupResponse, error) {
	c.init()

	c.Logger.Debug("dialing using QUIC")

	if req.uri.Scheme != "moqt" {
		err := errors.New("invalid scheme")
		c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		return nil, nil, err
	}

	c.Logger.Debug("dialing QUIC", "host", req.uri.Hostname(), "port", req.uri.Port(), "path", req.uri.Path)

	// Add path parameter
	if c.SetupExtensions == nil {
		c.SetupExtensions = NewParameters()
		c.Logger.Debug("SetupExtensions initialized")
	}
	c.SetupExtensions.SetString(param_type_path, req.uri.Path)
	c.Logger.Debug("path parameter set", "path", req.uri.Path)

	if DialQUICFunc == nil {
		DialQUICFunc = defaultDialQUICFunc
		c.Logger.Debug("DialQUICFunc initialized")
	}

	// Dial QUIC connection
	c.Logger.Debug("dialing QUIC connection", "address", req.uri.Hostname()+":"+req.uri.Port())
	conn, err := DialQUICFunc(ctx, req.uri.Hostname()+":"+req.uri.Port(), c.TLSConfig, c.QUICConfig)
	if err != nil {
		c.Logger.Error("failed to dial QUIC connection", "error", err.Error(), "address", req.uri.Hostname()+":"+req.uri.Port())
		return nil, nil, err
	}

	var param *Parameters
	if c.SetupExtensions != nil {
		param = c.SetupExtensions
		c.Logger.Debug("using setup extensions", "extensions", param)
	} else {
		c.Logger.Debug("no setup extensions provided")
	}
	sess, rsp, err := OpenSession(conn, param, c.TrackResolver)
	if err != nil {
		c.Logger.Error("failed to open session stream", "error", err.Error())
		return nil, nil, err
	}

	c.Logger.Debug("setup response received", "version", rsp.selectedVersion, "parameters", rsp.Parameters)

	return sess, rsp, nil
}
