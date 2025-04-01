package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

const (
	/*
	 * Bidirectional Stream Type
	 */
	stream_type_session   message.StreamType = 0x0
	stream_type_announce  message.StreamType = 0x1
	stream_type_subscribe message.StreamType = 0x2
	stream_type_info      message.StreamType = 0x4

	/*
	 * Unidirectional Stream Type
	 */
	stream_type_group message.StreamType = 0x0
)

func openStream(conn quic.Connection, st message.StreamType) (quic.Stream, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		return nil, err
	}

	stream.Write([]byte{byte(st)})

	return stream, nil
}
