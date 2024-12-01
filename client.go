package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type Client struct {
	URL string

	SupportedVersions []Version

	TLSConfig *tls.Config

	QUICConfig *quic.Config

	SetupParameters Parameters

	HijackSetupRarams func(Parameters) error

	ClientSessionHandler ClientSessionHandler

	RequestHandler RequestHandler

	RelayManager *RelayManager
}

func (c Client) Run(ctx context.Context) error {
	/*
	 * Initialize the Client
	 */
	if c.RelayManager == nil {
		c.RelayManager = defaultRelayManager
	}

	if c.SupportedVersions == nil {
		c.SupportedVersions = []Version{Default}
	}

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
	var conn moq.Connection
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

		conn = moq.NewMOWTConnection(sess)

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
			// Get Address
			addr := ip.String()
			if strings.Contains(addr, ":") && !strings.HasPrefix(addr, "[") {
				addr = "[" + addr + "]"
			}
			addr += ":" + parsedURL.Port()

			// Dial
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

			conn = moq.NewMORQConnection(qconn)

			break
		}

		req = SetupRequest{
			SupportedVersions: c.SupportedVersions,
			Parameters:        c.SetupParameters,
		}

		// Add the path to the parameters
		req.Parameters.Add(PATH, parsedURL.Path)

	default:
		err := errors.New("invalid scheme")
		slog.Error("url scheme must be https or moqt", slog.String("scheme", parsedURL.Scheme))
		return err
	}

	/*
	 * Get a new Session
	 */
	// Initialize a Session
	sess := ClientSession{
		session: &session{
			conn:                  conn,
			subscribeWriters:      make(map[SubscribeID]*SubscribeWriter),
			receivedSubscriptions: make(map[SubscribeID]Subscription),
			doneCh:                make(chan struct{}, 1),
		},
	}

	// Open a bidirectional Stream for the Session Stream
	stream, err := sess.openControlStream(stream_type_session)
	if err != nil {
		slog.Error("failed to open a Session Stream", slog.String("error", err.Error()))
		return err
	}
	sess.stream = stream

	/*
	 * Set up
	 */
	/*
	 * Send a SESSION_CLIENT message
	 */
	err = sendSetupRequest(sess.stream, req)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return err
	}

	rsp, err := readSetupResponce(sess.stream)
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return err
	}

	// Verify the selceted version is contained in the supported versions
	if !ContainVersion(rsp.SelectedVersion, req.SupportedVersions) {
		err = errors.New("unexpected version was seleted")
		slog.Error("failed to negotiate versions", slog.String("error", err.Error()), slog.Any("selected version", rsp.SelectedVersion))
		return err
	}

	// Handle the responce
	if c.HijackSetupRarams != nil {
		err := c.HijackSetupRarams(rsp.Parameters)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	}

	/*
	 * Handle the Client Session
	 */
	go c.ClientSessionHandler.HandleClientSession(&sess)

	/*
	 * Listen a bidirectional stream
	 */
	go c.listenBiStreams(&sess)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	slog.Info("Client is running. Press ctrl+C to exit.")
	<-signalCh

	slog.Info("Shunnting down client")

	return nil
}

func sendSetupRequest(w io.Writer, req SetupRequest) error {
	scm := message.SessionClientMessage{
		SupportedVersions: make([]protocol.Version, 0),
		Parameters:        message.Parameters(req.Parameters),
	}

	for _, v := range req.SupportedVersions {
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(v))
	}

	err := scm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func readSetupResponce(r io.Reader) (SetupResponce, error) {
	/***/
	var ssm message.SessionServerMessage
	err := ssm.Decode(r)
	if err != nil {
		slog.Error("failed to read a SESSION_SERVER message", slog.String("error", err.Error()))
		return SetupResponce{}, err
	}

	return SetupResponce{
		SelectedVersion: Version(ssm.SelectedVersion),
		Parameters:      Parameters(ssm.Parameters),
	}, nil
}

func (c Client) listenBiStreams(sess *ClientSession) {
	for {
		stream, err := sess.conn.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		go func(stream moq.Stream) {
			/*
			 * Read a Stream Type
			 */
			buf := make([]byte, 1)
			_, err := stream.Read(buf)
			if err != nil {
				slog.Error("failed to read a Stream Type ID", slog.String("error", err.Error()))
			}

			switch StreamType(buf[0]) {
			case stream_type_announce:
				slog.Info("Announce Stream was opened")

				interest, err := readInterest(stream)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				log.Print("INTEREST", interest)

				w := AnnounceWriter{
					doneCh: make(chan struct{}, 1),
					stream: stream,
				}

				// Announce
				announcements, ok := c.RelayManager.FindAnnouncements(interest.TrackPrefix)
				if !ok || announcements == nil {
					announcements = make([]Announcement, 0)
				}

				c.RequestHandler.HandleInterest(interest, announcements, w)
				<-w.doneCh
			case stream_type_subscribe:
				slog.Info("Subscribe Stream was opened")

				subscription, err := readSubscription(stream)
				if err != nil {
					slog.Error("failed to get a subscription", slog.String("error", err.Error()))

					// Close the Stream gracefully
					slog.Info("closing a Subscribe Stream", slog.String("error", err.Error()))
					err = stream.Close()
					if err != nil {
						slog.Error("failed to close the stream", slog.String("error", err.Error()))
						return
					}

					return
				}

				//
				sw := SubscribeResponceWriter{
					stream: stream,
					doneCh: make(chan struct{}, 1),
				}

				// Get any Infomation of the track
				info, ok := c.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
				if ok {
					// Handle with out
					c.RequestHandler.HandleSubscribe(subscription, &info, sw)
				} else {
					c.RequestHandler.HandleSubscribe(subscription, nil, sw)
				}

				<-sw.doneCh

				/*
				 * Accept the new subscription
				 */
				sess.acceptSubscription(subscription)

				/*
				 * Catch any Subscribe Update or any error from the subscriber
				 */
				for {
					update, err := readSubscribeUpdate(subscription, stream)
					if err != nil {
						slog.Info("catched an error from the subscriber", slog.String("error", err.Error()))
						break
					}

					slog.Info("received a subscribe update request", slog.Any("subscription", update))

					sw := SubscribeResponceWriter{
						stream: stream,
						doneCh: make(chan struct{}, 1),
					}

					// Get any Infomation of the track
					info, ok := c.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
					if ok {
						c.RequestHandler.HandleSubscribe(update, &info, sw)
					} else {
						c.RequestHandler.HandleSubscribe(update, nil, sw)
					}

					<-sw.doneCh

					slog.Info("updated a subscription", slog.Any("from", subscription), slog.Any("to", update))

					/*
					 * Update the subscription
					 */
					sess.updateSubscription(update)
					subscription = update
				}

				sess.stopSubscription(subscription.subscribeID)

				// Close the Stream gracefully
				err = stream.Close()
				if err != nil {
					slog.Error("failed to close the stream", slog.String("error", err.Error()))
					return
				}
			case stream_type_fetch:
				slog.Info("Fetch Stream was opened")

				fetchRequest, err := readFetchRequest(stream)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					return
				}

				w := FetchResponceWriter{
					doneCh: make(chan struct{}, 1),
					stream: stream,
				}

				c.RequestHandler.HandleFetch(fetchRequest, w)

				<-w.doneCh
			case stream_type_info:
				slog.Info("Info Stream was opened")

				infoRequest, err := readInfoRequest(stream)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				w := InfoWriter{
					doneCh: make(chan struct{}, 1),
					stream: stream,
				}

				info, ok := c.RelayManager.GetInfo(infoRequest.TrackNamespace, infoRequest.TrackName)
				if ok {
					c.RequestHandler.HandleInfoRequest(infoRequest, &info, w)
				} else {
					c.RequestHandler.HandleInfoRequest(infoRequest, nil, w)
				}

				<-w.doneCh
			default:
				err := ErrInvalidStreamType

				// Cancel reading and writing
				stream.CancelRead(err.StreamErrorCode())
				stream.CancelWrite(err.StreamErrorCode())

				return
			}
		}(stream)
	}
}
