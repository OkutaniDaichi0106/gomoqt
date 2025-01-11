package moqt

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type FetchUpdate struct {
	GroupPriority GroupPriority
}

func readFetchUpdate(r io.Reader) (FetchUpdate, error) {
	var fum message.FetchUpdateMessage
	err := fum.Decode(r)
	if err != nil {
		return FetchUpdate{}, err
	}

	return FetchUpdate{
		GroupPriority: GroupPriority(fum.GroupPriority),
	}, nil
}

func writeFetchUpdate(w io.Writer, update FetchUpdate) error {
	// Send a fetch update message
	fum := message.FetchUpdateMessage{
		GroupPriority: message.GroupPriority(update.GroupPriority),
	}
	err := fum.Encode(w)
	if err != nil {
		return err
	}

	return nil
}

func updateFetch(fetch FetchRequest, update FetchUpdate) (FetchRequest, error) {
	fetch.GroupPriority = update.GroupPriority

	return fetch, nil
}
