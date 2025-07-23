package moqt

import (
	"context"
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ context.Context = (*subscribeContext)(nil)

type subscribeContext struct {
	context.Context
}

func (c *subscribeContext) Cause() error {
	reason := context.Cause(c.Context)

	var strErr *quic.StreamError
	if errors.As(reason, &strErr) {
		return &SubscribeError{
			StreamError: strErr,
		}
	}

	return reason
}

type announceContext struct {
	context.Context
}

func (c *announceContext) Cause() error {
	reason := context.Cause(c.Context)

	var strErr *quic.StreamError
	if errors.As(reason, &strErr) {
		return &AnnounceError{
			StreamError: strErr,
		}
	}

	return reason
}

type sessionContext struct {
	context.Context
}

func (c *sessionContext) Cause() error {
	reason := context.Cause(c.Context)

	var appErr *quic.ApplicationError
	if errors.As(reason, &appErr) {
		return &SessionError{
			ApplicationError: appErr,
		}
	}

	return reason
}
