package moqt

import "context"

type TrackReader interface {
	ReadInfo() Info

	// Accept a group
	AcceptGroup(context.Context) (GroupReader, error)

	Close() error

	CloseWithError(SubscribeErrorCode) error
}
