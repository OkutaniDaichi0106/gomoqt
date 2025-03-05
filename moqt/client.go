package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
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
	Logger slog.Logger
}

func (c Client) Dial(urlStr string, ctx context.Context) (Session, SetupResponse, error) {
	c.Logger.Debug("dialing to the server", "url", urlStr)
	req, err := NewSetupRequest(urlStr)
	if err != nil {
		c.Logger.Error("failed to parse the url", "error", err.Error())
		return nil, SetupResponse{}, err
	}

	// Dial based on the scheme
	switch req.uri.Scheme {
	case "https":
		return c.DialWebTransport(*req, ctx)
	case "moqt":
		return c.DialQUIC(*req, ctx)
	default:
		err = errors.New("invalid scheme")
		c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		return nil, SetupResponse{}, err
	}
}

func (c Client) DialWebTransport(req SetupRequest, ctx context.Context) (Session, SetupResponse, error) {
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
		c.Logger.Error("failed to dial with webtransport", "error", err.Error())
		return nil, SetupResponse{}, err
	}

	// Ensure wtsess is not nil before proceeding
	if wtsess == nil {
		return nil, SetupResponse{}, errors.New("webtransport session is nil after dial")
	}

	// Open a session stream
	sess, stream, err := internal.OpenSession(transport.NewMOWTConnection(wtsess), c.SetupExtensions.paramMap)
	if err != nil {
		c.Logger.Error("failed to open a session stream", slog.String("error", err.Error()))
		sess.Terminate(err)
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("A session established over WebTransport", "selectedVersion", stream.SessionServerMessage.SelectedVersion)

	rsp := SetupResponse{
		selectedVersion: stream.SessionServerMessage.SelectedVersion,
		Parameters:      Parameters{stream.SessionServerMessage.Parameters},
	}

	return &session{internalSession: sess}, rsp, nil
}

func (c Client) DialQUIC(req SetupRequest, ctx context.Context) (Session, SetupResponse, error) {
	if req.uri.Scheme != "moqt" {
		err := errors.New("invalid scheme")
		c.Logger.Error("unsupported url scheme", "scheme", req.uri.Scheme)
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("dialing QUIC", "host", req.uri.Hostname(), "port", req.uri.Port(), "path", req.uri.Path)

	// Add path parameter
	c.SetupExtensions.SetString(param_type_path, req.uri.Path)

	// Look up the IP address
	var ips []net.IP
	ips, err := net.LookupIP(req.uri.Hostname())
	if err != nil {
		c.Logger.Error("failed to look up IP address", "error", err.Error())
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
			c.Logger.Error("failed to dial with quic", "error", err.Error(), "attempt", i+1)
			if i+1 >= len(ips) {
				err = errors.New("no more IPs")
				c.Logger.Error("failed to dial to the host",
					"error", err.Error(),
					"host", req.uri.Hostname(),
				)
				return nil, SetupResponse{}, err

			}
			continue
		}
		c.Logger.Debug("successful QUIC dial", "address", addr)
		break
	}

	isess, stream, err := internal.OpenSession(transport.NewMORQConnection(qconn), c.SetupExtensions.paramMap)
	if err != nil {
		c.Logger.Error("failed to dial with quic", "error", err.Error())
		return nil, SetupResponse{}, err
	}

	c.Logger.Debug("A session established over QUIC", "selected version", stream.SessionServerMessage.SelectedVersion)

	rsp := SetupResponse{
		selectedVersion: stream.SessionServerMessage.SelectedVersion,
		Parameters:      Parameters{stream.SessionServerMessage.Parameters},
	}

	return &session{internalSession: isess}, rsp, nil
}
