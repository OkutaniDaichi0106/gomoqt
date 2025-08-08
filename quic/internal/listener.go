package internal

import (
	"context"
	"net"
)

type EarlyListener interface {
	Accept(ctx context.Context) (Connection, error)
	Addr() net.Addr
	Close() error
}
