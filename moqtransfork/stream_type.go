package moqtransfork

import "github.com/OkutaniDaichi0106/gomoqt/internal/message"

const (
	/*
	 * Bidirectional Stream Type
	 */
	stream_type_session   message.StreamType = 0x0
	stream_type_announce  message.StreamType = 0x1
	stream_type_subscribe message.StreamType = 0x2
	stream_type_fetch     message.StreamType = 0x3
	stream_type_info      message.StreamType = 0x4

	/*
	 * Unidirectional Stream Type
	 */
	stream_type_group message.StreamType = 0x0
)
