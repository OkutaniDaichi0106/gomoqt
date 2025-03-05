package internal

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
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

func openStream(conn transport.Connection, st message.StreamType) (transport.Stream, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		return nil, err
	}

	stream.Write([]byte{byte(st)})

	return stream, nil
}
