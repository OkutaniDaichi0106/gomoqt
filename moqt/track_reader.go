package moqt

import "context"

type TrackReader interface {
	// Accept a group
	AcceptGroup(context.Context) (GroupReader, error)

	ReceiveGap(context.Context) (Gap, error)

	Close() error

	CloseWithError(error) error
}
