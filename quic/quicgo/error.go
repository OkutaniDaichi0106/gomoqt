package quicgo

import (
	"github.com/OkutaniDaichi0106/gomoqt/quic/internal"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func wrapError(err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *quicgo_quicgo.StreamError:
		return &internal.StreamError{
			StreamID:  internal.StreamID(e.StreamID),
			ErrorCode: internal.StreamErrorCode(e.ErrorCode),
			Remote:    e.Remote,
			Err:       e,
		}
	case *quicgo_quicgo.TransportError:
		return &internal.TransportError{
			Remote:       e.Remote,
			FrameType:    e.FrameType,
			ErrorCode:    internal.TransportErrorCode(e.ErrorCode),
			ErrorMessage: e.ErrorMessage,
			Err:          e,
		}
	case *quicgo_quicgo.ApplicationError:
		return &internal.ApplicationError{
			Remote:       e.Remote,
			ErrorCode:    internal.ApplicationErrorCode(e.ErrorCode),
			ErrorMessage: e.ErrorMessage,
			Err:          e,
		}
	case *quicgo_quicgo.VersionNegotiationError:
		ours := make([]internal.Version, len(e.Ours))
		for i, v := range e.Ours {
			ours[i] = internal.Version(v)
		}
		theirs := make([]internal.Version, len(e.Theirs))
		for i, v := range e.Theirs {
			theirs[i] = internal.Version(v)
		}
		return &internal.VersionNegotiationError{
			Ours:   ours,
			Theirs: theirs,
			Err:    e,
		}
	case *quicgo_quicgo.StatelessResetError:
		return &internal.StatelessResetError{
			Err: e,
		}

	case *quicgo_quicgo.IdleTimeoutError:
		return &internal.IdleTimeoutError{
			Err: e,
		}
	case *quicgo_quicgo.HandshakeTimeoutError:
		return &internal.HandshakeTimeoutError{
			Err: e,
		}
	default:
		// If the error is not recognized, return it as is
		return err
	}
}

func WrapTransportError(qerr quicgo_quicgo.TransportError) *internal.TransportError {
	return &internal.TransportError{
		Remote:       qerr.Remote,
		FrameType:    qerr.FrameType,
		ErrorCode:    internal.TransportErrorCode(qerr.ErrorCode),
		ErrorMessage: qerr.ErrorMessage,
	}
}
