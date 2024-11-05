package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Publisher struct {
	// Handlers
	Handler PublisherHandler

	LocalTrack []string

	interestCh chan struct {
		Interest
		AnnounceWriter
	}

	subscriptionCh chan struct {
		Subscription
		SubscribeResponceWriter
	}

	infoReqCh chan struct {
		InfoRequest
		InfoWriter
	}

	fetchReqCh chan struct {
		FetchRequest
		FetchResponceWriter
	}
}

type PublisherHandler interface {
	InterestHandler
	SubscribeHandler
	FetchHandler
	InfoRequestHandler
}

func (p Publisher) init() {
	p.interestCh = make(chan struct {
		Interest
		AnnounceWriter
	}, 1<<2) // TODO: tune the size

	p.subscriptionCh = make(chan struct {
		Subscription
		SubscribeResponceWriter
	}, 1<<2) // TODO: tune the size

	p.infoReqCh = make(chan struct {
		InfoRequest
		InfoWriter
	}, 1<<2) // TODO: tune the size

	p.fetchReqCh = make(chan struct {
		FetchRequest
		FetchResponceWriter
	}, 1<<2) // TODO: tune the size
}

func (p Publisher) listen() {
	for {
		select {
		case v := <-p.interestCh:
			p.Handler.HandleInterest(v.Interest, v.AnnounceWriter)
		case v := <-p.subscriptionCh:
			p.Handler.HandleSubscribe(v.Subscription, v.SubscribeResponceWriter)
		case v := <-p.infoReqCh:
			p.Handler.HandleInfoRequest(v.InfoRequest, v.InfoWriter)
		case v := <-p.fetchReqCh:
			p.Handler.HandleFetch(v.FetchRequest, v.FetchResponceWriter)
		}
	}
}

func getInterest(r quicvarint.Reader) (Interest, error) {
	var aim message.AnnounceInterestMessage
	err := aim.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return Interest{}, err
	}
	return Interest(aim), nil
}

func getFetchRequest(r quicvarint.Reader) (FetchRequest, error) {
	var frm message.FetchMessage
	err := frm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a FETCH message", slog.String("error", err.Error()))
		return FetchRequest{}, err
	}

	return FetchRequest(frm), nil
}

func getInfoRequest(r quicvarint.Reader) (InfoRequest, error) {
	var irm message.InfoRequestMessage
	err := irm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read an INFO_REQUEST message", slog.String("error", err.Error()))
		return InfoRequest{}, err
	}

	return InfoRequest(irm), nil
}
