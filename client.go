package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
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

	SetupHijackerFunc func(SetupResponce) error

	announcements map[string]Announcement

	SubscribeHijackerFunc func(Subscription) error

	SessionHandler ClientSessionHandler

	CacheManager CacheManager
}

func (c Client) AddAnnouncement(announcement Announcement) {
	if c.announcements == nil {
		c.announcements = make(map[string]Announcement)
	}

	c.announcements[announcement.TrackPath] = announcement
}

func (c Client) Run(ctx context.Context) error {
	/*
	 * Initialize the Client
	 */
	if c.SupportedVersions == nil {
		c.SupportedVersions = []Version{Default}
	}
	if c.CacheManager == nil {
		// TODO: Handle the nil Cache Manager
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
	// Open a bidirectional Stream for the Session Stream
	var sess ClientSession
	err = sess.init(conn)
	if err != nil {
		slog.Error("failed to open a Session Stream", slog.String("error", err.Error()))
		return err
	}

	/*
	 * Set up
	 */
	// Send a set-up request
	err = sendSetupRequest(sess.stream, req)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return err
	}

	// Receive a set-up responce
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
	if c.SetupHijackerFunc != nil {
		err := c.SetupHijackerFunc(rsp)
		if err != nil {
			slog.Error("setup hijacker returns an error", slog.String("error", err.Error()))
			return err
		}
	}

	/*
	 * Handle the Client Session
	 */
	go c.SessionHandler.HandleClientSession(&sess)

	/*
	 * Listen a bidirectional stream
	 */
	go c.listenBiStreams(&sess)

	// Catch a shutting down signal
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

func (c Client) listenBiStreams(sess *ClientSession) {
	for {
		stream, err := sess.conn.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		slog.Debug("some control stream was opened")

		go func(stream moq.Stream) {
			/*
			 * Read a Stream Type
			 */
			var stm message.StreamTypeMessage
			err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			switch stm.StreamType {
			case stream_type_announce:
				slog.Debug("announce stream was opened")

				interest, err := readInterest(stream)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				aw := AnnounceWriter{
					stream: stream,
				}

				// Announce
				for _, announcement := range c.announcements {
					// Verify if the Announcement's Track Namespace has the Track Prefix
					if strings.HasPrefix(announcement.TrackPath, interest.TrackPrefix) {
						// Announce the Track Namespace
						aw.Announce(announcement)
					}
				}
			case stream_type_subscribe:
				slog.Debug("subscribe stream was opened")

				//
				sw := SubscribeResponceWriter{
					stream: stream,
				}

				subscription, err := readSubscription(stream)
				if err != nil {
					slog.Error("failed to get a subscription", slog.String("error", err.Error()))

					sw.Reject(err)
					return
				}

				// Hijack
				if c.SubscribeHijackerFunc != nil {
					err := c.SubscribeHijackerFunc(subscription)
					if err != nil {
						sw.Reject(err)
						return
					}
				}

				// Get any Infomation of the track
				info, ok := sess.getInfo(subscription.TrackPath)
				if ok {
					sw.Accept(info)
				} else {

					sw.Reject(ErrTrackDoesNotExist)
					return

				}

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
					}

					// Get any Infomation of the track
					info, ok := sess.getInfo(subscription.TrackPath)
					if ok {
						sw.Accept(info)
					} else {
						sw.Reject(ErrTrackDoesNotExist)
						return
					}

					slog.Info("updated a subscription", slog.Any("from", subscription), slog.Any("to", update))

					/*
					 * Update the subscription
					 */
					sess.updateSubscription(update)
					subscription = update
				}

				sess.removeSubscription(subscription)

				// Close the Stream gracefully
				sw.Reject(nil)

			case stream_type_fetch:
				slog.Debug("fetch stream was opened")

				frw := FetchResponceWriter{
					stream: stream,
				}

				req, err := readFetchRequest(stream)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				// Get a data reader
				r, err := c.CacheManager.GetFrame(req.TrackPath, req.GroupSequence, req.FrameSequence)
				if err != nil {
					slog.Error("failed to get a frame", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				// Verify if subscriptions corresponding to the ftch request exists
				for _, subscription := range sess.receivedSubscriptions {
					if subscription.TrackPath != req.TrackPath {
						continue
					}

					// Send the group data
					w, err := frw.SendGroup(Group{
						subscribeID:       subscription.subscribeID,
						groupSequence:     req.GroupSequence,
						PublisherPriority: PublisherPriority(req.SubscriberPriority), // TODO: Handle Publisher Priority
					})
					if err != nil {
						slog.Error("failed to send a group", slog.String("error", err.Error()))
						frw.Reject(err)
						return
					}

					// Send the data by copying it from the reader
					io.Copy(w, r)
				}

				// Close the Fetch Stream gracefully
				frw.Reject(nil)
				return
			case stream_type_info:
				slog.Debug("info stream was opened")

				req, err := readInfoRequest(stream)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				// Initialize an Info Writer
				iw := InfoWriter{
					stream: stream,
				}

				info, ok := sess.getInfo(req.TrackPath)
				if ok {
					iw.Answer(info)
				} else {
					iw.Reject(ErrTrackDoesNotExist)
					return
				}

				// Close the Info Stream gracefully
				iw.Reject(nil)
				return
			default:
				slog.Debug("unknown stream was opend")

				// Terminate the session
				sess.Terminate(ErrInvalidStreamType)

				return
			}
		}(stream)
	}
}
