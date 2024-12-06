package moqt

import (
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type Track struct {
	TrackPath string
}

type Publisher interface {
	NewTrack(Announcement) Track
	OpenDataStream(Track, Group) (moq.SendStream, error)
	Terminate(error)
}

type publisherManager struct {
	/*
	 * Received Subscriptions
	 */
	subscribeReceivers map[SubscribeID]*SubscribeReceiver
	srMu               sync.RWMutex
}
