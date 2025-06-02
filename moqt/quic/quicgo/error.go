package quicgo

import (
	quic "github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

var ConnectionRefused = quic.ConnectionRefused

func WrapStreamError(qerr quicgo.StreamError) *quic.StreamError {
	return &quic.StreamError{
		StreamID:  quic.StreamID(qerr.StreamID),
		ErrorCode: quic.StreamErrorCode(qerr.ErrorCode),
		Remote:    qerr.Remote,
	}
}
