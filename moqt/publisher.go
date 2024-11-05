package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Publisher struct {
	// Handlers
	Handler PublisherHandler
}

type PublisherHandler interface {
	InterestHandler
	SubscribeHandler
	FetchHandler
	InfoRequestHandler

	DataWriter
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
