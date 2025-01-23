package moqt

import (
	"fmt"
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type FetchUpdate struct {
	TrackPriority TrackPriority
}

func (fu FetchUpdate) String() string {
	return fmt.Sprintf("FetchUpdate: { TrackPriority: %d }", fu.TrackPriority)
}

func readFetchUpdate(r io.Reader) (FetchUpdate, error) {
	var fum message.FetchUpdateMessage
	err := fum.Decode(r)
	if err != nil {
		return FetchUpdate{}, err
	}

	return FetchUpdate{}, nil
}

func writeFetchUpdate(w io.Writer, update FetchUpdate) error {
	// Send a fetch update message
	fum := message.FetchUpdateMessage{
		TrackPriority: message.TrackPriority(update.TrackPriority),
	}
	err := fum.Encode(w)
	if err != nil {
		return err
	}

	return nil
}

func updateFetch(fetch FetchRequest, update FetchUpdate) (FetchRequest, error) {
	fetch.TrackPriority = update.TrackPriority

	return fetch, nil
}
