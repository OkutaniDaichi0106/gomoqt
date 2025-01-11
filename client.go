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

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type Client struct {
	TLSConfig *tls.Config

	QUICConfig *quic.Config

	//supportedVersions []Version

	// JitterManager JitterManager
}

func (c Client) Dial(req SetupRequest, ctx context.Context) (clientSession, SetupResponce, error) {
	// Initialize the request
	err := req.init()
	if err != nil {
		slog.Error("failed to initialize the request", slog.String("error", err.Error()))
		return clientSession{}, SetupResponce{}, err
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
		return clientSession{}, SetupResponce{}, err
	}
}

func (c Client) DialWebTransport(req SetupRequest, ctx context.Context) (clientSession, SetupResponce, error) {
	// Initialize the request
	err := req.init()
	if err != nil {
		slog.Error("failed to initialize the request", slog.String("error", err.Error()))
		return clientSession{}, SetupResponce{}, err
	}

	// Check the scheme
	if req.parsedURL.Scheme != "https" {
		slog.Error("unsupported url scheme", slog.String("scheme", req.parsedURL.Scheme))
		return clientSession{}, SetupResponce{}, errors.New("invalid scheme")
	}

	// Dial on webtransport
	var wtsess *webtransport.Session
	var d webtransport.Dialer
	_, wtsess, err = d.Dial(ctx, req.URL, http.Header{}) // TODO: configure the header
	if err != nil {
		slog.Error("failed to dial with webtransport", slog.String("error", err.Error()))
		return clientSession{}, SetupResponce{}, err
	}

	// Get a moq.Connection
	conn := transport.NewMOWTConnection(wtsess)

	return setupConnection(req, conn)
}

func (c Client) DialQUIC(req SetupRequest, ctx context.Context) (clientSession, SetupResponce, error) {
	// Initialize the request
	err := req.init()
	if err != nil {
		slog.Error("failed to initialize the request", slog.String("error", err.Error()))
		return clientSession{}, SetupResponce{}, err
	}

	// Check the scheme
	if req.parsedURL.Scheme != "moqt" {
		err = errors.New("invalid scheme")
		slog.Error("unsupported url scheme", slog.String("scheme", req.parsedURL.Scheme))
		return clientSession{}, SetupResponce{}, err
	}

	// Look up the IP address
	var ips []net.IP
	ips, err = net.LookupIP(req.parsedURL.Hostname())
	if err != nil {
		slog.Error("failed to look up IP address", slog.String("error", err.Error()))
		return clientSession{}, SetupResponce{}, err
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
				return clientSession{}, SetupResponce{}, err
			}
			continue
		}

		// Get a moq.Connection
		conn = transport.NewMORQConnection(qconn)

		break
	}

	return setupConnection(req, conn)
}

func setupConnection(req SetupRequest, conn transport.Connection) (clientSession, SetupResponce, error) {
	// Open a Session Stream
	stream, err := openSessionStream(conn)
	if err != nil {
		slog.Error("failed to open a Session Stream")
		return clientSession{}, SetupResponce{}, err
	}

	// Send a set-up request
	err = sendSetupRequest(stream, req)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return clientSession{}, SetupResponce{}, err
	}

	// Receive a set-up responce
	rsp, err := readSetupResponce(stream)
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return clientSession{}, SetupResponce{}, err
	}

	sess := clientSession{
		session: newSession(conn, stream),
	}

	go listenSession(sess.session, context.Background()) // TODO:

	return sess, rsp, nil
}

func listenSession(sess *session, ctx context.Context) {
	// Listen the bidirectional streams
	go listenBiStreams(sess, ctx)

	// Listen the unidirectional streams
	go listenUniStreams(sess, ctx)

	// Listen the datagrams
	go listenDatagrams(sess, ctx)
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

	return stream, nil
}

func sendSetupRequest(w io.Writer, req SetupRequest) error {
	scm := message.SessionClientMessage{
		SupportedVersions: make([]protocol.Version, 0),
		Parameters:        message.Parameters(req.SetupParameters),
	}

	for _, v := range req.supportedVersions {
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(v))
	}

	err := scm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func listenBiStreams(sess *session, ctx context.Context) {
	for {
		/*
		 * Accept a bidirectional stream
		 */
		stream, err := sess.conn.AcceptStream(ctx)
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		slog.Debug("some control stream was opened")

		// Handle the stream
		go func(stream transport.Stream) {
			/*
			 * Get a Stream Type ID
			 */
			var stm message.StreamTypeMessage
			err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			// Handle the stream by the Stream Type ID
			switch stm.StreamType {
			case stream_type_announce:
				// Handle the announce stream
				slog.Debug("announce stream was opened")
				// Get an Interest
				interest, err := readInterest(stream)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					closeStreamWithInternalError(stream, err)
					return
				}

				sas := &sendAnnounceStream{
					interest:     interest,
					stream:       stream,
					activeTracks: make(map[string]Announcement),
				}

				// Enqueue the interest
				sess.sendAnnounceStreamQueue.Enqueue(sas)
			case stream_type_subscribe:
				slog.Debug("subscribe stream was opened")

				id, subscription, err := readSubscription(stream)
				if err != nil {
					slog.Error("failed to get a received subscription", slog.String("error", err.Error()))
					closeStreamWithInternalError(stream, err)
					return
				}

				rss := &receiveSubscribeStream{
					subscribeID:  id,
					subscription: subscription,
					stream:       stream,
				}

				// Enqueue the subscription
				sess.receiveSubscribeStreamQueue.Enqueue(rss)

				// Listen updates
				for {
					update, err := readSubscribeUpdate(stream)
					if err != nil {
						slog.Error("failed to read a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
						closeStreamWithInternalError(stream, err)
						break
					}

					subscription, err := updateSubscription(rss.subscription, update)
					if err != nil {
						slog.Error("failed to update a subscription", slog.String("error", err.Error()))
						closeStreamWithInternalError(stream, err)
						return
					}

					rss.subscription = subscription
				}
			case stream_type_fetch:
				slog.Debug("fetch stream was opened")
				// Get a fetch-request
				fetch, err := readFetch(stream)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					closeStreamWithInternalError(stream, err)
					return
				}

				rfs := &receiveFetchStream{
					fetch:  fetch,
					stream: stream,
				}

				// Enqueue the fetch
				sess.receiveFetchStreamQueue.Enqueue(rfs)

				// Listen updates
				for {
					update, err := readFetchUpdate(stream)
					if err != nil {
						slog.Error("failed to read a FETCH_UPDATE message", slog.String("error", err.Error()))
						closeStreamWithInternalError(stream, err)
						break
					}

					fetch, err := updateFetch(rfs.fetch, update)
					if err != nil {
						slog.Error("failed to update a fetch", slog.String("error", err.Error()))
						closeStreamWithInternalError(stream, err)
						return
					}

					rfs.fetch = fetch

					slog.Info("updated a fetch", slog.Any("fetch", rfs.fetch))
				}
			case stream_type_info:
				slog.Debug("info stream was opened")

				// Get a received info-request
				req, err := newReceivedInfoRequest(stream)
				if err != nil {
					slog.Error("failed to get a received info-request", slog.String("error", err.Error()))
					closeStreamWithInternalError(stream, err)
					return
				}

				// Enqueue the info-request
				sess.receivedInfoRequestQueue.Enqueue(req)
			default:
				slog.Debug("An unknown type of stream was opend")

				// Terminate the session
				sess.Terminate(ErrProtocolViolation)

				return
			}
		}(stream)
	}
}

func listenUniStreams(sess *session, ctx context.Context) {
	for {
		/*
		 * Accept a unidirectional stream
		 */
		stream, err := sess.conn.AcceptUniStream(ctx)
		if err != nil {
			slog.Error("failed to accept a unidirectional stream", slog.String("error", err.Error()))
			return
		}

		slog.Debug("some data stream was opened")

		// Handle the stream
		go func(stream transport.ReceiveStream) {
			/*
			 * Get a Stream Type ID
			 */
			var stm message.StreamTypeMessage
			err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			// Handle the stream by the Stream Type ID
			switch stm.StreamType {
			case stream_type_group:
				slog.Debug("group stream was opened")

				id, group, err := readGroup(stream)
				if err != nil {
					slog.Error("failed to get a group", slog.String("error", err.Error()))
					return
				}

				data := &receiveDataStream{
					subscribeID:   id,
					ReceiveStream: stream,
					ReceivedGroup: group,
				}

				queue, ok := sess.dataReceiveStreamQueues[data.SubscribeID()]
				if !ok {
					slog.Error("failed to get a data receive stream queue", slog.String("error", "queue not found"))
					closeReceiveStreamWithInternalError(stream, ErrProtocolViolation) // TODO:
					return
				}

				// Enqueue the receiver
				queue.Enqueue(data)
			default:
				slog.Debug("An unknown type of stream was opend")

				// Terminate the session
				sess.Terminate(ErrProtocolViolation)

				return
			}
		}(stream)
	}
}

func listenDatagrams(sess *session, ctx context.Context) {
	for {
		/*
		 * Receive a datagram
		 */
		buf, err := sess.conn.ReceiveDatagram(ctx)
		if err != nil {
			slog.Error("failed to receive a datagram", slog.String("error", err.Error()))
			return
		}

		// Handle the datagram
		go func(buf []byte) {
			data, err := newReceivedDatagram(buf)
			if err != nil {
				slog.Error("failed to get a received datagram", slog.String("error", err.Error()))
				return
			}

			//
			queue, ok := sess.receivedDatagramQueues[data.SubscribeID()]
			if !ok {
				slog.Error("failed to get a data receive stream queue", slog.String("error", "queue not found"))
				return
			}

			// Enqueue the datagram
			queue.Enqueue(data)
		}(buf)
	}
}

func closeStreamWithInternalError(stream transport.Stream, err error) {
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
}

func closeReceiveStreamWithInternalError(stream transport.ReceiveStream, err error) {
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
}
