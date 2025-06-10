package moqt

import "context"

type TrackReader interface {
	// Accept a group
	AcceptGroup(context.Context) (GroupReader, error)

	Close() error

	CloseWithError(SubscribeErrorCode) error
}
