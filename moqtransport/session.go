package moqtransport

import (
	"bufio"
	"context"
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type Session struct {
	Connection Connection

	setupStream Stream

	selectedVersion moqtmessage.Version

	trackAliasMap *trackAliasMap
	//
	subscribeCounter uint64
	//
	maxSubscribeID *moqtmessage.SubscribeID
}

const (
	SETUP_STREAM     StreamType = 0x0
	ANNOUNCE_STREAM  StreamType = 0x1
	SUBSCRIBE_STREAM StreamType = 0x2
)

func (sess Session) OpenAnnounceStream(stream Stream) (*ReceiveAnnounceStream, error) {
	// Send the Stream Type ID and notify the Stream Type is the Announce
	_, err := stream.Write([]byte{byte(ANNOUNCE_STREAM)})
	if err != nil {
		return nil, err
	}

	// Set the Stream Type to the Announce
	stream.SetType(ANNOUNCE_STREAM)

	return &ReceiveAnnounceStream{
		stream:   stream,
		qvReader: quicvarint.NewReader(stream),
	}, nil
}

func (sess Session) AcceptAnnounceStream(stream Stream, ctx context.Context) (*SendAnnounceStream, error) {
	/*
	 * Verify the Stream Type is the Announce
	 */
	// Peek and read the Stream Type
	peeker := bufio.NewReader(stream)
	b, err := peeker.Peek(1)
	if err != nil {
		return nil, err
	}
	// Verify the Stream Type ID
	if StreamType(b[0]) != ANNOUNCE_STREAM {
		return nil, ErrUnexpectedStreamType
	}
	// Read and advance by 1 byte
	streamTypeBuf := make([]byte, 1)
	_, err = stream.Read(streamTypeBuf)
	if err != nil {
		return nil, err
	}

	// Set the Stream Type to the Announce
	stream.SetType(ANNOUNCE_STREAM)

	return &SendAnnounceStream{
		stream:   stream,
		qvReader: quicvarint.NewReader(stream),
	}, nil
}

func (sess Session) OpenSubscribeStream(stream Stream) (*SendSubscribeStream, error) {
	// Send the Stream Type ID and notify the Stream Type is the Subscribe
	_, err := stream.Write([]byte{byte(SUBSCRIBE_STREAM)})
	if err != nil {
		return nil, err
	}

	// Set the Stream Type to the Subscribe
	stream.SetType(ANNOUNCE_STREAM)

	return &SendSubscribeStream{
		stream:           stream,
		qvReader:         quicvarint.NewReader(stream),
		subscribeCounter: &sess.subscribeCounter,
		trackAliasMap:    sess.trackAliasMap,
	}, nil
}

func (sess Session) AcceptSubscribeStream(stream Stream, ctx context.Context) (*ReceiveSubscribeStream, error) {
	/*
	 * Verify the Stream Type is the Subscribe
	 */
	// Read and advance by 1 byte
	streamTypeBuf := make([]byte, 1)
	_, err := stream.Read(streamTypeBuf)
	if err != nil {
		return nil, err
	}

	// Set the Stream Type to the Subscribe
	stream.SetType(SUBSCRIBE_STREAM)

	return &ReceiveSubscribeStream{
		stream:   stream,
		qvReader: quicvarint.NewReader(stream),
	}, nil
}

func (sess Session) PeekStreamType(stream Stream) (StreamType, error) {
	// Peek and read the Stream Type
	peeker := bufio.NewReader(stream)
	b, err := peeker.Peek(1)
	if err != nil {
		return 0, err
	}

	return StreamType(b[0]), nil
}

type AnnounceConfig struct {
	AuthorizationInfo []string

	MaxCacheDuration time.Duration
}

type SubscribeConfig struct {
	// Required
	SubscriberPriority moqtmessage.SubscriberPriority
	GroupOrder         moqtmessage.GroupOrder
	MinGroupSequence   uint64
	MaxGroupSequence   uint64

	// Optional
	AuthorizationInfo *string
	DeliveryTimeout   *time.Duration
	Parameters        moqtmessage.Parameters
}

var ErrUnexpectedStreamType = errors.New("unexpected stream type")
