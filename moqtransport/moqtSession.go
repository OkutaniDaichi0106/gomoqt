package moqtransport

import (
	"context"
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type moqtSession struct {
	Connection

	sessionStream Stream

	selectedVersion moqtmessage.Version

	trackAliasMap *trackAliasMap
	//
	subscribeCounter uint64
}

const (
	SETUP_STREAM     byte = 0x0
	ANNOUNCE_STREAM  byte = 0x1
	SUBSCRIBE_STREAM byte = 0x2
)

func (sess moqtSession) OpenAnnounceStream() (*ReceiveAnnounceStream, error) {
	// Open an Stream
	stream, err := sess.Connection.OpenStream()
	if err != nil {
		return nil, err
	}
	// Send the Announce Stream ID and notify the stream type is the Announce
	stream.Write([]byte{ANNOUNCE_STREAM})

	return &ReceiveAnnounceStream{
		stream:   stream,
		qvReader: quicvarint.NewReader(stream),
	}, nil
}

func (sess moqtSession) AcceptAnnounceStream(ctx context.Context) (*SendAnnounceStream, error) {
	stream, err := sess.Connection.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}

	// Receive a Stream ID
	idBuf := make([]byte, 1)
	stream.Read(idBuf)
	// Verify the Stream Type is an Announce
	if idBuf[0] != ANNOUNCE_STREAM {
		return nil, ErrUnexpectedStream
	}

	return &SendAnnounceStream{
		stream:   stream,
		qvReader: quicvarint.NewReader(stream),
	}, nil
}

func (sess moqtSession) AcceptSubscribeStream(ctx context.Context) (*ReceiveSubscribeStream, error) {
	stream, err := sess.Connection.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}

	// Receive a Stream ID
	idBuf := make([]byte, 1)
	stream.Read(idBuf)
	// Verify the Stream Type is an Announce
	if idBuf[0] != ANNOUNCE_STREAM {
		return nil, ErrUnexpectedStream
	}

	return &ReceiveSubscribeStream{
		stream:   stream,
		qvReader: quicvarint.NewReader(stream),
	}, nil
}

func (sess moqtSession) OpenSubscribeStream() (*SendSubscribeStream, error) {
	// Open an Stream
	stream, err := sess.Connection.OpenStream()
	if err != nil {
		return nil, err
	}
	// Send the Announce Stream ID and notify the stream type is the Announce
	stream.Write([]byte{SUBSCRIBE_STREAM})

	return &SendSubscribeStream{
		stream:           stream,
		qvReader:         quicvarint.NewReader(stream),
		subscribeCounter: &sess.subscribeCounter,
		trackAliasMap:    sess.trackAliasMap,
	}, nil
}

func (sess moqtSession) Terminate(err TerminateError) {
	sess.CloseWithError(SessionErrorCode(err.Code()), err.Error())
}

type AnnounceConfig struct {
	AuthorizationInfo []string

	MaxCacheDuration time.Duration
}

type SubscribeConfig struct {
	// Required
	SubscriberPriority moqtmessage.SubscriberPriority
	GroupOrder         moqtmessage.GroupOrder

	// Optional
	AuthorizationInfo *string
	DeliveryTimeout   *time.Duration
	Parameters        *moqtmessage.Parameters
}

var ErrUnexpectedStream = errors.New("unexpected stream")
