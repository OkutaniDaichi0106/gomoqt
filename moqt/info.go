package moqt

import (
	"fmt"
)

type Info struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
}

func (i Info) String() string {
	return fmt.Sprintf("Info: { TrackPriority: %d, LatestGroupSequence: %d, GroupOrder: %d }", i.TrackPriority, i.LatestGroupSequence, i.GroupOrder)
}

// func readInfo(r io.Reader) (Info, error) {
// 	slog.Debug("reading an info")

// 	// Read an INFO message
// 	var im message.InfoMessage
// 	_, err := im.Decode(r)
// 	if err != nil {
// 		slog.Error("failed to read a INFO message", slog.String("error", err.Error()))
// 		return Info{}, err
// 	}

// 	info := Info{
// 		TrackPriority:       TrackPriority(im.TrackPriority),
// 		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
// 		GroupOrder:          GroupOrder(im.GroupOrder),
// 	}

// 	slog.Debug("read an info")

// 	return info, nil
// }

// func writeInfo(w io.Writer, info Info) error {
// 	slog.Debug("writing an info")

// 	im := message.InfoMessage{
// 		TrackPriority:       message.TrackPriority(info.TrackPriority),
// 		LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
// 		GroupOrder:          message.GroupOrder(info.GroupOrder),
// 	}

// 	_, err := im.Encode(w)
// 	if err != nil {
// 		slog.Error("failed to send a INFO message", slog.String("error", err.Error()))
// 		return err
// 	}

// 	slog.Debug("wrote an info")

// 	return nil
// }
