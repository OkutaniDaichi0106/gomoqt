package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type Client struct {
	TLSConfig *tls.Config

	QUICConfig *quic.Config

	Config *Config
}

func (c Client) Dial(urlStr string, ctx context.Context) (Session, SetupResponce, error) {
	slog.Debug("dialing to the server")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		slog.Error("failed to parse the url", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	/*
	 * Dial
	 */
	switch parsedURL.Scheme {
	case "https":
		return c.DialWebTransport(parsedURL.Hostname(), parsedURL.Port(), parsedURL.Path, ctx)
	case "moqt":
		return c.DialQUIC(parsedURL.Hostname(), parsedURL.Port(), parsedURL.Path, ctx)
	default:

		err = errors.New("invalid scheme")
		slog.Error("unsupported url scheme", slog.String("scheme", parsedURL.Scheme))
		return nil, SetupResponce{}, err
	}
}

func (c Client) DialWebTransport(host string, port string, path string, ctx context.Context) (Session, SetupResponce, error) {
	// Check if Config is initialized
	if c.Config == nil {
		c.Config = &Config{
			SetupExtensions: NewParameters(),
		}
	}

	// Dial on webtransport
	var d webtransport.Dialer
	_, wtsess, err := d.Dial(ctx, "https://"+host+":"+port+path, http.Header{}) // TODO: configure the header
	if err != nil {
		slog.Error("failed to dial with webtransport", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	// Ensure wtsess is not nil before proceeding
	if wtsess == nil {
		return nil, SetupResponce{}, errors.New("webtransport session is nil after dial")
	}

	sess, ssm, err := internal.SetupWebTransport(ctx, wtsess, c.Config.SetupExtensions.paramMap)
	if err != nil {
		slog.Error("failed to setup webtransport session", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	rsp := SetupResponce{selectedVersion: ssm.SelectedVersion, Parameters: Parameters{ssm.Parameters}}

	return &session{internalSession: sess}, rsp, nil
}

func (c Client) DialQUIC(host string, port string, path string, ctx context.Context) (Session, SetupResponce, error) {
	// Check if Config is initialized
	if c.Config == nil {
		c.Config = &Config{
			SetupExtensions: NewParameters(),
		}
	}

	// Add path parameter
	c.Config.SetupExtensions.SetString(param_type_path, path)

	// Look up the IP address
	var ips []net.IP
	ips, err := net.LookupIP(host)
	if err != nil {
		slog.Error("failed to look up IP address", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	var qconn quic.Connection

	// Try all IPs

	for i, ip := range ips {
		// Get Address
		addr := ip.String()
		if strings.Contains(addr, ":") && !strings.HasPrefix(addr, "[") {
			addr = "[" + addr + "]"
		}
		addr += ":" + port

		// Dial
		qconn, err = quic.DialAddrEarly(ctx, addr, c.TLSConfig, c.QUICConfig)
		if err != nil {
			slog.Error("failed to dial with quic", slog.String("error", err.Error()))
			if i+1 >= len(ips) {
				err = errors.New("no more IPs")
				slog.Error("failed to dial to the host",
					slog.String("error", err.Error()),
					slog.String("host", host),
				)
				return nil, SetupResponce{}, err

			}
			continue
		}

		break
	}

	isess, ssm, err := internal.SetupQUIC(ctx, qconn, c.Config.SetupExtensions.paramMap)
	if err != nil {
		slog.Error("failed to dial with quic", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	rsp := SetupResponce{selectedVersion: ssm.SelectedVersion, Parameters: Parameters{ssm.Parameters}}

	return &session{internalSession: isess}, rsp, nil
}

func openSession(conn transport.Connection, params Parameters) (Session, SetupResponce, error) {
	scm := message.SessionClientMessage{
		SupportedVersions: internal.DefaultClientVersions,
		Parameters:        message.Parameters(params.paramMap),
	}

	sess, ssm, err := internal.OpenSessionStream(conn, scm)
	if err != nil {
		slog.Error("failed to setup a session", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	rsp := SetupResponce{
		selectedVersion: ssm.SelectedVersion,
		Parameters:      Parameters{ssm.Parameters},
	}

	return &session{internalSession: internal.NewSession(conn, sess)}, rsp, nil
}
