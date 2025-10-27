package moqt

import (
	"context"
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func Cause(ctx context.Context) error {
	reason := context.Cause(ctx)

	var strErr *quic.StreamError
	if errors.As(reason, &strErr) {
		st, ok := ctx.Value(&biStreamTypeCtxKey).(message.StreamType)
		if ok {
			switch st {
			case message.StreamTypeSession:
				return &SessionError{
					ApplicationError: &quic.ApplicationError{
						Remote:       strErr.Remote,
						ErrorCode:    quic.ApplicationErrorCode(strErr.ErrorCode),
						ErrorMessage: "moqt: closed session stream",
					},
				}
			case message.StreamTypeAnnounce:
				return &AnnounceError{
					StreamError: strErr,
				}
			case message.StreamTypeSubscribe:
				return &SubscribeError{
					StreamError: strErr,
				}
			}

			return reason
		}

		st, ok = ctx.Value(&uniStreamTypeCtxKey).(message.StreamType)
		if ok {
			switch st {
			case message.StreamTypeGroup:
				return &GroupError{
					StreamError: strErr,
				}
			}
		}

		return reason
	}

	var appErr *quic.ApplicationError
	if errors.As(reason, &appErr) {
		return &SessionError{
			ApplicationError: appErr,
		}
	}

	return reason
}

var biStreamTypeCtxKey int
var uniStreamTypeCtxKey int
