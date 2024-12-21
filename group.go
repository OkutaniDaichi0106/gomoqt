package moqt

import (
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type Group interface {
	SubscribeID() SubscribeID
	GroupSequence() GroupSequence
	GroupPriority() GroupPriority
}

type ReceivedGroup struct {
	subscribeID SubscribeID

	groupSequence GroupSequence

	groupPriority GroupPriority

	/*
	 * Fields not in wire
	 */
	// Time when the Group was received
	receivedAt time.Time // TODO:
}

func (g ReceivedGroup) SubscribeID() SubscribeID {
	return g.subscribeID
}

func (g ReceivedGroup) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g ReceivedGroup) GroupPriority() GroupPriority {
	return g.groupPriority
}

type SentGroup struct {
	subscribeID SubscribeID

	groupSequence GroupSequence

	groupPriority GroupPriority

	/*
	 * Fields not in wire
	 */
	// Time when the Group was sent
	sentAt time.Time // TODO:
}

func (g SentGroup) SubscribeID() SubscribeID {
	return g.subscribeID
}

func (g SentGroup) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g SentGroup) GroupPriority() GroupPriority {
	return g.groupPriority
}

func readGroup(r io.Reader) (ReceivedGroup, error) {
	// Read a GROUP message
	var gm message.GroupMessage
	err := gm.Decode(r)
	if err != nil {
		slog.Error("failed to read a GROUP message", slog.String("error", err.Error()))
		return ReceivedGroup{}, err
	}

	//
	return ReceivedGroup{
		subscribeID:   SubscribeID(gm.SubscribeID),
		groupSequence: GroupSequence(gm.GroupSequence),
		groupPriority: GroupPriority(gm.GroupPriority),
		receivedAt:    time.Now(),
	}, nil
}
