package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
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

func (c *Client) Dial(urlStr string, ctx context.Context) (Session, SetupResponse, error) {
	c.init()

	c.Logger.Info("dialing server", "url", urlStr) // Changed to Info for better visibility
	req, err := NewSetupRequest(urlStr)
	if err != nil {
		c.Logger.Error("failed to create setup request", "error", err.Error(), "url", urlStr)
		return nil, SetupResponse{}, err
	}

	// Dial based on the scheme
	switch req.uri.Scheme {
	case "https":
		c.Logger.Debug("dialing using WebTransport")
		return c.DialWebTransport(*req, ctx)
	case "moqt":
		c.Logger.Debug("dialing using QUIC")
		return c.DialQUIC(*req, ctx)
	default:
		err = errors.New("invalid scheme")
		c.Logger.Error("unsupported URL scheme", "scheme", req.uri.Scheme, "url", urlStr)
		return nil, SetupResponse{}, err
	}
}

func (c *Client) DialWebTransport(req SetupRequest, ctx context.Context) (Session, SetupResponse, error) {
	c.init()

	if req.uri.Scheme != "https" {
		err := errors.New("invalid scheme")
		c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("dialing WebTransport", "endpoint", "https://"+req.uri.Hostname()+":"+req.uri.Port()+req.uri.Path)

	// Dial on webtransport
	var d webtransport.Dialer
	_, wtsess, err := d.Dial(ctx, "https://"+req.uri.Hostname()+":"+req.uri.Port()+req.uri.Path, http.Header{}) // TODO: configure the header
	if err != nil {
		c.Logger.Error("failed to dial WebTransport", "error", err.Error(), "endpoint", "https://"+req.uri.Hostname()+":"+req.uri.Port()+req.uri.Path)
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("webtransport session established", "remote_address", wtsess.RemoteAddr(), "path", req.uri.Path)

	// Ensure wtsess is not nil before proceeding
	if wtsess == nil {
		err := errors.New("webtransport session is nil after dial")
		c.Logger.Error("WebTransport session is nil", "error", err.Error())
		return nil, SetupResponse{}, err
	}

	// Open a session stream
	var param message.Parameters
	if c.SetupExtensions != nil {
		param = c.SetupExtensions.paramMap
		c.Logger.Debug("using setup extensions", "extensions", param)
	} else {
		c.Logger.Debug("no setup extensions provided")
	}
	sess, stream, err := internal.OpenSession(transport.NewMOWTConnection(wtsess), param)
	if err != nil {
		c.Logger.Error("failed to open session stream", "error", err.Error())
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("session established over WebTransport", "selectedVersion", stream.SessionServerMessage.SelectedVersion)

	rsp := SetupResponse{
		selectedVersion: stream.SessionServerMessage.SelectedVersion,
		Parameters:      Parameters{stream.SessionServerMessage.Parameters},
	}

	c.Logger.Info("setup response received", "version", rsp.selectedVersion, "parameters", rsp.Parameters)

	return &session{internalSession: sess}, rsp, nil
}

func (c *Client) DialQUIC(req SetupRequest, ctx context.Context) (Session, SetupResponse, error) {
	c.init()

	c.Logger.Debug("dialing using QUIC")

	if req.uri.Scheme != "moqt" {
		err := errors.New("invalid scheme")
		c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("dialing QUIC", "host", req.uri.Hostname(), "port", req.uri.Port(), "path", req.uri.Path)

	// Add path parameter
	if c.SetupExtensions == nil {
		c.SetupExtensions = NewParameters()
		c.Logger.Debug("SetupExtensions initialized")
	}
	c.SetupExtensions.SetString(param_type_path, req.uri.Path)
	c.Logger.Debug("path parameter set", "path", req.uri.Path)

	// Look up the IP address
	var ips []net.IP
	ips, err := net.LookupIP(req.uri.Hostname())
	if err != nil {
		c.Logger.Error("failed to look up IP address", "error", err.Error(), "host", req.uri.Hostname())
		return nil, SetupResponse{}, err
	}
	c.Logger.Debug("resolved IPs", "ips", ips)

	var qconn quic.Connection

	// Try all IPs

	for i, ip := range ips {
		// Get Address
		addr := ip.String()
		if strings.Contains(addr, ":") && !strings.HasPrefix(addr, "[") {
			addr = "[" + addr + "]"
		}
		addr += ":" + req.uri.Port()

		// Dial
		qconn, err = quic.DialAddrEarly(ctx, addr, c.TLSConfig, c.QUICConfig)
		if err != nil {
			c.Logger.Error("failed to dial with quic", "error", err.Error(), "address", addr, "attempt", i+1)
			if i+1 >= len(ips) {
				err = errors.New("no more IPs to try")
				c.Logger.Error("failed to dial to the host",
					"error", err.Error(),
					"host", req.uri.Hostname(),
				)
				return nil, SetupResponse{}, err

			}
			continue
		}
		c.Logger.Info("QUIC connection established", "address", addr) // Changed to Info
		break
	}

	var param message.Parameters
	if c.SetupExtensions != nil {
		param = c.SetupExtensions.paramMap
		c.Logger.Debug("using setup extensions", "extensions", param)
	} else {
		c.Logger.Debug("no setup extensions provided")
	}
	isess, stream, err := internal.OpenSession(transport.NewMORQConnection(qconn), param)
	if err != nil {
		c.Logger.Error("failed to open session stream", "error", err.Error())
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("session established over QUIC", "selectedVersion", stream.SessionServerMessage.SelectedVersion)

	rsp := SetupResponse{
		selectedVersion: stream.SessionServerMessage.SelectedVersion,
		Parameters:      Parameters{stream.SessionServerMessage.Parameters},
	}

	c.Logger.Info("setup response received", "version", rsp.selectedVersion, "parameters", rsp.Parameters) // Changed to Info

	return &session{internalSession: isess}, rsp, nil
}
