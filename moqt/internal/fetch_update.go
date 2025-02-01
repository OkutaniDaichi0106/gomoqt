package internal

import "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"

func updateFetch(fm *message.FetchMessage, fum *message.FetchUpdateMessage) {
	if fm == nil || fum == nil {
		fm = &message.FetchMessage{}
		return
	}

	fm.TrackPriority = fum.TrackPriority
}
