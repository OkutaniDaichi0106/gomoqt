package moqt

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

/*
 * Sequence number of a group in a track
 * When this is integer more than 1, the number means the sequence number.
 * When this is 0, it indicates the sequence number is currently unknown .
 * 0 is used to specify "the latest sequence number" or "the final sequence number of an open-ended track", "the first sequence number of the default order".
 */
type GroupSequence uint64

const (
	FirstSequence  GroupSequence = 1
	LatestSequence GroupSequence = NotSpecified
	FinalSequence  GroupSequence = NotSpecified
	NotSpecified   GroupSequence = 0
	MaxSequence    GroupSequence = 0xFFFFFFFF
)

func (gs GroupSequence) String() string {
	return fmt.Sprintf("GroupSequence: %d", gs)
}

func (gs GroupSequence) Next() GroupSequence {
	if gs == FinalSequence {
		return FinalSequence
	}

	if gs == LatestSequence {
		return LatestSequence
	}

	if gs == MaxSequence {
		return 1
	}

	return gs + 1
}

/***/
type FetchRequest struct {
	SubscribeID   SubscribeID
	TrackPath     []string
	TrackPriority TrackPriority
	GroupSequence GroupSequence
	FrameSequence FrameSequence
}

func (fr FetchRequest) String() string {
	var sb strings.Builder
	sb.WriteString("FetchRequest: {")
	sb.WriteString(" SubscribeID: ")
	sb.WriteString(fmt.Sprintf("%d", fr.SubscribeID))
	sb.WriteString(", TrackPath: [")
	for i, path := range fr.TrackPath {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(path)
	}
	sb.WriteString("], TrackPriority: ")
	sb.WriteString(fmt.Sprintf("%d", fr.TrackPriority))
	sb.WriteString(", GroupSequence: ")
	sb.WriteString(fmt.Sprintf("%d", fr.GroupSequence))
	sb.WriteString(", FrameSequence: ")
	sb.WriteString(fmt.Sprintf("%d", fr.FrameSequence))
	sb.WriteString(" }")
	return sb.String()
}

func readFetch(r io.Reader) (FetchRequest, error) {
	var fm message.FetchMessage
	_, err := fm.Decode(r)
	if err != nil {
		slog.Error("failed to read a FETCH message", slog.String("error", err.Error()))
		return FetchRequest{}, err
	}

	req := FetchRequest{
		SubscribeID:   SubscribeID(fm.SubscribeID),
		TrackPath:     fm.TrackPath,
		TrackPriority: TrackPriority(fm.TrackPriority),
		GroupSequence: GroupSequence(fm.GroupSequence),
		FrameSequence: FrameSequence(fm.FrameSequence),
	}

	return req, nil
}

func writeFetch(w io.Writer, fetch FetchRequest) error {
	fm := message.FetchMessage{
		SubscribeID:   message.SubscribeID(fetch.SubscribeID),
		TrackPath:     fetch.TrackPath,
		TrackPriority: message.TrackPriority(fetch.TrackPriority),
		GroupSequence: message.GroupSequence(fetch.GroupSequence),
		FrameSequence: message.FrameSequence(fetch.FrameSequence),
	}
	_, err := fm.Encode(w)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return err
	}

	return nil
}
