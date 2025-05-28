package moqtrace

import (
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type SessionTracer struct {
	SessionEstablished func(local, remote net.Addr, alpn string, version protocol.Version, extension map[uint64][]byte)
	SessionTerminated  func(reason error)

	// QUIC
	QUICStreamOpened   func(quic.StreamID) StreamTracer
	QUICStreamAccepted func(quic.StreamID) StreamTracer
}

type StreamTracer struct {
	// QUIC
	StreamClosed           func()                 // FIN
	SendStreamCancelled    func(quic.StreamError) // RESET_STREAM
	ReceiveStreamCancelled func(quic.StreamError) // STOP_SENDING

	// MOQ
	// Stream Type
	StreamTypeMessageSent     func(message.StreamTypeMessage)
	StreamTypeMessageReceived func(message.StreamTypeMessage)

	// Session
	SessionClientMessageSent     func(message.SessionClientMessage)
	SessionClientMessageReceived func(message.SessionClientMessage)
	SessionServerMessageSent     func(message.SessionServerMessage)
	SessionServerMessageReceived func(message.SessionServerMessage)
	SessionUpdateMessageSent     func(message.SessionUpdateMessage)
	SessionUpdateMessageReceived func(message.SessionUpdateMessage)

	// Announce
	AnnouncePleaseMessageSent     func(message.AnnouncePleaseMessage)
	AnnouncePleaseMessageReceived func(message.AnnouncePleaseMessage)
	AnnounceMessageSent           func(message.AnnounceMessage)
	AnnounceMessageReceived       func(message.AnnounceMessage)

	// Subscribe
	SubscribeMessageSent       func(message.SubscribeMessage)
	SubscribeMessageReceived   func(message.SubscribeMessage)
	SubscribeOkMessageSent     func(message.SubscribeOkMessage)
	SubscribeOkMessageReceived func(message.SubscribeOkMessage)

	// Group
	GroupMessageSent     func(message.GroupMessage)
	GroupMessageReceived func(message.GroupMessage)

	// Frame
	FrameMessageSent     func(message.FrameMessage)
	FrameMessageReceived func(message.FrameMessage)
}
