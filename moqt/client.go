package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Client struct {
	URL string

	SupportedVersions []Version

	TLSConfig *tls.Config

	QUICConfig *quic.Config

	SetupParameters Parameters

	HandleSetupResponce func(SetupResponce) error

	PublisherHandler
	SubscriberHandler
}

func (c Client) init() {
	if c.SupportedVersions == nil {
		c.SupportedVersions = []Version{Default}
	}
}

func (c Client) Run(ctx context.Context) error {
	c.init()

	/*
	 * Dial
	 */
	// Verify the URI
	parsedURL, err := url.ParseRequestURI(c.URL)
	if err != nil {
		slog.Error("failed to parse the url", slog.String("error", err.Error()))
		return err
	}

	// Handle the scheme
	var conn Connection
	var req SetupRequest
	switch parsedURL.Scheme {
	case "https":
		// Dial with webtransport
		var d webtransport.Dialer
		_, sess, err := d.Dial(ctx, c.URL, http.Header{}) // TODO: configure the header

		if err != nil {
			slog.Error("failed to dial with webtransport", slog.String("error", err.Error()))
			return err
		}

		conn = NewMOWTConnection(sess)

		req = SetupRequest{
			SupportedVersions: c.SupportedVersions,
			Parameters:        c.SetupParameters,
		}
	case "moqt":
		// Dial with raw quic
		ips, err := net.LookupIP(parsedURL.Hostname())
		if err != nil {
			slog.Error("failed to look up IP address", slog.String("error", err.Error()))
			return err
		}

		// Try all IPs
		for i, ip := range ips {
			addr := ip.String() + parsedURL.Port()
			qconn, err := quic.DialAddrEarly(ctx, addr, c.TLSConfig, c.QUICConfig)
			if err != nil {
				slog.Error("failed to dial with quic", slog.String("error", err.Error()))
				if i+1 >= len(ips) {
					err = errors.New("no more IPs")
					slog.Error("failed to dial to the host", slog.String("error", err.Error()), slog.String("host", parsedURL.Hostname()))
					return err
				}
				continue
			}

			conn = NewMORQConnection(qconn)

			break
		}

		req = SetupRequest{
			Path:              parsedURL.Path,
			SupportedVersions: c.SupportedVersions,
			Parameters:        c.SetupParameters,
		}
	default:
		err := errors.New("invalid scheme")
		slog.Error("url scheme must be https or moqt", slog.String("scheme", parsedURL.Scheme))
		return err
	}

	/*
	 * Get a Session
	 */
	// Open a bidirectional Stream for the Session Stream
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return err
	}
	// Initialize a Session
	sess := Session{
		Connection:    conn,
		SessionStream: stream,
	}

	/*
	 * Set up
	 */
	// Send the Session Stream Type
	_, err = sess.SessionStream.Write([]byte{byte(SESSION)})
	if err != nil {
		slog.Error("failed to send a Session Stream Type", slog.String("error", err.Error()))
		return err
	}

	err = sendSetupRequest(sess.SessionStream, req)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return err
	}

	rsp, err := getSetupResponce(quicvarint.NewReader(sess.SessionStream))
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return err
	}

	// Verify the selceted version is contained in the
	if !ContainVersion(rsp.SelectedVersion, req.SupportedVersions) {
		err = errors.New("unexpected version was seleted")
		slog.Error("failed to negotiate versions", slog.String("error", err.Error()), slog.Any("selected version", rsp.SelectedVersion))
		return err
	}

	// Handle the responce
	if c.HandleSetupResponce != nil {
		err := c.HandleSetupResponce(rsp)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	}

	//

	return nil
}

func sendSetupRequest(stream Stream, req SetupRequest) error {
	scm := message.SessionClientMessage{
		SupportedVersions: make([]protocol.Version, 0),
	}

	for _, v := range req.SupportedVersions {
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(v))
	}

	scm.Parameters.Add(PATH, req.Path)

	_, err := stream.Write(scm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", err.Error())
		return err
	}

	return nil
}

func getSetupResponce(r quicvarint.Reader) (SetupResponce, error) {
	/***/
	var ssm message.SessionServerMessage
	err := ssm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a SESSION_SERVER message", slog.String("error", err.Error()))
		return SetupResponce{}, err
	}

	return SetupResponce{
		SelectedVersion: Version(ssm.SelectedVersion),
		Parameters:      Parameters(ssm.Parameters),
	}, nil
}
