package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type Group interface {
	GroupSequence() GroupSequence
	GroupPriority() GroupPriority
}

var _ Group = (*group)(nil)

type group struct {
	groupSequence GroupSequence
	groupPriority GroupPriority
}

func (g group) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g group) GroupPriority() GroupPriority {
	return g.groupPriority
}

func readGroup(r io.Reader) (SubscribeID, Group, error) {
	// Read a GROUP message
	var gm message.GroupMessage
	err := gm.Decode(r)
	if err != nil {
		slog.Error("failed to read a GROUP message", slog.String("error", err.Error()))
		return 0, nil, err
	}

	//
	return SubscribeID(gm.SubscribeID), group{
		groupSequence: GroupSequence(gm.GroupSequence),
		groupPriority: GroupPriority(gm.GroupPriority),
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
