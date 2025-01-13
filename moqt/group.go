package moqt

import (
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type Group interface {
	GroupSequence() GroupSequence
	GroupPriority() GroupPriority
}

type ReceivedGroup interface {
	Group
	ReceivedAt() time.Time
}

var _ ReceivedGroup = (*receivedGroup)(nil)

type receivedGroup struct {
	// subscribeID SubscribeID

	groupSequence GroupSequence

	groupPriority GroupPriority

	/*
	 * Fields not in wire
	 */
	// Time when the Group was received
	receivedAt time.Time // TODO:
}

func (g receivedGroup) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g receivedGroup) GroupPriority() GroupPriority {
	return g.groupPriority
}

func (g receivedGroup) ReceivedAt() time.Time {
	return g.receivedAt
}

type SentGroup interface {
	Group
	SentAt() time.Time
}

var _ SentGroup = (*sentGroup)(nil)

type sentGroup struct {
	// subscribeID SubscribeID

	groupSequence GroupSequence

	groupPriority GroupPriority

	/*
	 * Fields not in wire
	 */
	// Time when the Group was sent
	sentAt time.Time // TODO:
}

func (g sentGroup) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g sentGroup) GroupPriority() GroupPriority {
	return g.groupPriority
}

func (g sentGroup) SentAt() time.Time {
	return g.sentAt
}

func readGroup(r io.Reader) (SubscribeID, ReceivedGroup, error) {
	// Read a GROUP message
	var gm message.GroupMessage
	err := gm.Decode(r)
	if err != nil {
		slog.Error("failed to read a GROUP message", slog.String("error", err.Error()))
		return 0, receivedGroup{}, err
	}

	//
	return SubscribeID(gm.SubscribeID), receivedGroup{
		groupSequence: GroupSequence(gm.GroupSequence),
		groupPriority: GroupPriority(gm.GroupPriority),
		receivedAt:    time.Now(),
	}, nil
}

func writeGroup(w io.Writer, id SubscribeID, g Group) error {
	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(id),
		GroupSequence: message.GroupSequence(g.GroupSequence()),
		GroupPriority: message.GroupPriority(g.GroupPriority()),
	}
	err := gm.Encode(w)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return err
	}

	return nil
}
