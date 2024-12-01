package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type PublisherPriority message.PublisherPriority

type Group struct {
	subscribeID SubscribeID

	groupSequence GroupSequence

	PublisherPriority PublisherPriority
}

func (g Group) SubscribeID() SubscribeID {
	return g.subscribeID
}

func (g Group) GroupSequence() GroupSequence {
	return g.groupSequence
}

// type GroupDrop message.GroupDrop
func readGroup(r io.Reader) (Group, error) {
	// Read a GROUP message
	var gm message.GroupMessage
	err := gm.Decode(r)
	if err != nil {
		slog.Error("failed to read a GROUP message", slog.String("error", err.Error()))
		return Group{}, err
	}

	//
	return Group{
		subscribeID:       SubscribeID(gm.SubscribeID),
		groupSequence:     GroupSequence(gm.GroupSequence),
		PublisherPriority: PublisherPriority(gm.PublisherPriority),
	}, nil
}
