package moqtransport

import (
	"sync"

	"github.com/quic-go/webtransport-go"
)

type PubSubAgent struct {
	clientAgent

	origin *PublisherAgent

	/*
	 * Destination the Agent send to
	 */
	destinations struct {
		sessions []*webtransport.Session
		mu       sync.Mutex
	}
}

func (*PubSubAgent) Role() Role {
	return PUB_SUB
}
