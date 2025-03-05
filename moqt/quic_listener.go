package moqt

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

type QUICEarlyListener interface {
	Accept(ctx context.Context) (quic.EarlyConnection, error)
	Addr() net.Addr
	Close() error
}
