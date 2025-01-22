package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type Client struct {
	TLSConfig *tls.Config

	QUICConfig *quic.Config

	//supportedVersions []Version

	// JitterManager JitterManager
}

func (c Client) Dial(req SetupRequest, ctx context.Context) (Session, SetupResponce, error) {
	slog.Debug("dialing to the server")

	// Initialize the request
	err := req.init()
	if err != nil {
		slog.Error("failed to initialize the request", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	/*
	 * Dial
	 */
	switch req.parsedURL.Scheme {
	case "https":
		return c.DialWebTransport(req, ctx)
	case "moqt":
		return c.DialQUIC(req, ctx)
	default:
		err = errors.New("invalid scheme")
		slog.Error("unsupported url scheme", slog.String("scheme", req.parsedURL.Scheme))
		return nil, SetupResponce{}, err
	}
}

func (c Client) DialWebTransport(req SetupRequest, ctx context.Context) (Session, SetupResponce, error) {
	slog.Debug("dialing to the server with webtransport")
	// Initialize the request
	err := req.init()
	if err != nil {
		slog.Error("failed to initialize the request", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	// Check the scheme
	if req.parsedURL.Scheme != "https" {
		slog.Error("unsupported url scheme", slog.String("scheme", req.parsedURL.Scheme))
		return nil, SetupResponce{}, errors.New("invalid scheme")
	}

	// Dial on webtransport
	var wtsess *webtransport.Session
	var d webtransport.Dialer
	_, wtsess, err = d.Dial(ctx, req.URL, http.Header{}) // TODO: configure the header
	if err != nil {
		slog.Error("failed to dial with webtransport", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	// Get a moq.Connection
	conn := transport.NewMOWTConnection(wtsess)

	return setupConnection(req, conn)
}

// TODO: test
func (c Client) DialQUIC(req SetupRequest, ctx context.Context) (Session, SetupResponce, error) {
	slog.Debug("dialing to the server with webtransport")

	// Initialize the request
	err := req.init()
	if err != nil {
		slog.Error("failed to initialize the request", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	// Check the scheme
	if req.parsedURL.Scheme != "moqt" {
		err = errors.New("invalid scheme")
		slog.Error("unsupported url scheme", slog.String("scheme", req.parsedURL.Scheme))
		return nil, SetupResponce{}, err
	}

	// Add path parameter
	req.SetupParameters.SetString(path, req.parsedURL.Path)

	// Look up the IP address
	var ips []net.IP
	ips, err = net.LookupIP(req.parsedURL.Hostname())
	if err != nil {
		slog.Error("failed to look up IP address", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	var conn transport.Connection

	// Try all IPs
	for i, ip := range ips {
		// Get Address
		addr := ip.String()
		if strings.Contains(addr, ":") && !strings.HasPrefix(addr, "[") {
			addr = "[" + addr + "]"
		}
		addr += ":" + req.parsedURL.Port()

		// Dial
		var qconn quic.Connection
		qconn, err = quic.DialAddrEarly(ctx, addr, c.TLSConfig, c.QUICConfig)
		if err != nil {
			slog.Error("failed to dial with quic", slog.String("error", err.Error()))
			if i+1 >= len(ips) {
				err = errors.New("no more IPs")
				slog.Error("failed to dial to the host",
					slog.String("error", err.Error()),
					slog.String("host", req.parsedURL.Hostname()),
				)
				return nil, SetupResponce{}, err
			}
			continue
		}

		// Get a moq.Connection
		conn = transport.NewMORQConnection(qconn)

		break
	}

	return setupConnection(req, conn)
}

func setupConnection(req SetupRequest, conn transport.Connection) (Session, SetupResponce, error) {
	// Open a Session Stream
	stream, err := openSessionStream(conn)
	if err != nil {
		slog.Error("failed to open a Session Stream")
		return nil, SetupResponce{}, err
	}

	// Send a set-up request
	err = sendSetupRequest(stream, req)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	// Receive a set-up responce
	rsp, err := readSetupResponce(stream)
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return nil, SetupResponce{}, err
	}

	return newSession(conn, stream), rsp, nil
}

func openSessionStream(conn transport.Connection) (transport.Stream, error) {
	slog.Debug("opening a session stream")

	/***/
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_session,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Debug("opened a session stream")

	return stream, nil
}

func sendSetupRequest(w io.Writer, req SetupRequest) error {
	slog.Debug("sending a set-up request")

	scm := message.SessionClientMessage{
		SupportedVersions: make([]protocol.Version, 0),
		Parameters:        message.Parameters(req.SetupParameters.paramMap),
	}

	for _, v := range req.supportedVersions {
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(v))
	}

	err := scm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("sent a set-up request")

	return nil
}

func closeStreamWithInternalError(stream transport.Stream, err error) {

	slog.Debug("closing the stream with an internal error", slog.String("error", err.Error()))

	if err == nil {
		stream.Close()
	}

	// TODO:

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	stream.CancelRead(code)
	stream.CancelWrite(code)

	slog.Debug("closed the stream with an internal error")
}

func closeReceiveStreamWithInternalError(stream transport.ReceiveStream, err error) {

	slog.Debug("closing the receive stream with an internal error", slog.String("error", err.Error()))

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	stream.CancelRead(code)

	slog.Debug("closed the receive stream with an internal error")
}
