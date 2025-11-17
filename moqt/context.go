package moqt

import (
	"context"
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

// Cause translates a Go context cancellation reason into a package-specific error type.
// When the provided context was canceled because of a QUIC stream error or application error,
// Cause converts that into the corresponding moqt error (e.g., SessionError, AnnounceError,
// SubscribeError, GroupError).
// If no specific translation is available, the original context cause is returned unchanged.
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
