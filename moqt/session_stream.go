package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SessionStream interface {
	UpdateSession(bitrate uint64) error
}

var _ SessionStream = (*sessionStream)(nil)

type sessionStream struct {
	stream  transport.Stream
	bitrate uint64
}

func (ss *sessionStream) UpdateSession(bitrate uint64) error {
	sum := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}

	err := sum.Encode(ss.stream)
	if err != nil {
		return err
	}

	ss.bitrate = bitrate
	return nil
}
