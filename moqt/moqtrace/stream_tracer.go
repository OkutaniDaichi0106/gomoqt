package moqtrace

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func InitStreamTracer(tracer *StreamTracer) {
	if tracer == nil {
		panic("StreamTracer must not be nil")
	}

	if tracer.SendStreamFinished == nil {
		tracer.SendStreamFinished = DefaultStreamFinished
	}
	if tracer.SendStreamReset == nil {
		tracer.SendStreamReset = DefaultStreamReset
	}
	if tracer.ReceiveStreamStopped == nil {
		tracer.ReceiveStreamStopped = DefaultStreamStopped
	}

	if tracer.StreamTypeMessageSent == nil {
		tracer.StreamTypeMessageSent = DefaultStreamTypeMessageSent
	}
	if tracer.StreamTypeMessageReceived == nil {
		tracer.StreamTypeMessageReceived = DefaultStreamTypeMessageReceived
	}

	if tracer.SessionClientMessageSent == nil {
		tracer.SessionClientMessageSent = DefaultSessionClientMessageSent
	}
	if tracer.SessionClientMessageReceived == nil {
		tracer.SessionClientMessageReceived = DefaultSessionClientMessageReceived
	}
	if tracer.SessionServerMessageSent == nil {
		tracer.SessionServerMessageSent = DefaultSessionServerMessageSent
	}
	if tracer.SessionServerMessageReceived == nil {
		tracer.SessionServerMessageReceived = DefaultSessionServerMessageReceived
	}
	if tracer.SessionUpdateMessageSent == nil {
		tracer.SessionUpdateMessageSent = DefaultSessionUpdateMessageSent
	}
	if tracer.SessionUpdateMessageReceived == nil {
		tracer.SessionUpdateMessageReceived = DefaultSessionUpdateMessageReceived
	}

	if tracer.AnnouncePleaseMessageSent == nil {
		tracer.AnnouncePleaseMessageSent = DefaultAnnouncePleaseMessageSent
	}
	if tracer.AnnouncePleaseMessageReceived == nil {
		tracer.AnnouncePleaseMessageReceived = DefaultAnnouncePleaseMessageReceived
	}
	if tracer.AnnounceMessageSent == nil {
		tracer.AnnounceMessageSent = DefaultAnnounceMessageSent
	}
	if tracer.AnnounceMessageReceived == nil {
		tracer.AnnounceMessageReceived = DefaultAnnounceMessageReceived
	}

	if tracer.SubscribeMessageSent == nil {
		tracer.SubscribeMessageSent = DefaultSubscribeMessageSent
	}
	if tracer.SubscribeMessageReceived == nil {
		tracer.SubscribeMessageReceived = DefaultSubscribeMessageReceived
	}
	if tracer.SubscribeOkMessageSent == nil {
		tracer.SubscribeOkMessageSent = DefaultSubscribeOkMessageSent
	}
	if tracer.SubscribeOkMessageReceived == nil {
		tracer.SubscribeOkMessageReceived = DefaultSubscribeOkMessageReceived
	}

	if tracer.GroupMessageSent == nil {
		tracer.GroupMessageSent = DefaultGroupMessageSent
	}
	if tracer.GroupMessageReceived == nil {
		tracer.GroupMessageReceived = DefaultGroupMessageReceived
	}

	if tracer.FrameMessageSent == nil {
		tracer.FrameMessageSent = DefaultFrameMessageSent
	}
	if tracer.FrameMessageReceived == nil {
		tracer.FrameMessageReceived = DefaultFrameMessageReceived
	}
}

type StreamTracer struct {
	SendStreamTracer
	ReceiveStreamTracer
}

type SendStreamTracer struct {
	// QUIC
	SendStreamFinished func()                             // FIN
	SendStreamReset    func(quic.StreamErrorCode, string) // RESET_STREAM
	SendStreamStopped  func(quic.StreamErrorCode, string) // STOP_SENDING

	// MOQ
	// Stream Type
	StreamTypeMessageSent func(message.StreamTypeMessage)

	// Session
	SessionClientMessageSent func(message.SessionClientMessage)
	SessionServerMessageSent func(message.SessionServerMessage)
	SessionUpdateMessageSent func(message.SessionUpdateMessage)

	// Announce
	AnnouncePleaseMessageSent func(message.AnnouncePleaseMessage)
	AnnounceMessageSent       func(message.AnnounceMessage)

	// Subscribe
	SubscribeMessageSent       func(message.SubscribeMessage)
	SubscribeOkMessageSent     func(message.SubscribeOkMessage)
	SubscribeUpdateMessageSent func(message.SubscribeUpdateMessage)

	// Group
	GroupMessageSent func(message.GroupMessage)

	// Frame
	FrameMessageSent func(frameCount, byteCount uint64)
}

type ReceiveStreamTracer struct {
	// QUIC
	ReceiveStreamFinished func()                             // FIN
	ReceiveStreamStopped  func(quic.StreamErrorCode, string) // STOP_SENDING
	ReceiveStreamReset    func(quic.StreamErrorCode, string) // RESET_STREAM

	// MOQ
	// Stream Type
	StreamTypeMessageReceived func(message.StreamTypeMessage)

	// Session
	SessionClientMessageReceived func(message.SessionClientMessage)
	SessionServerMessageReceived func(message.SessionServerMessage)
	SessionUpdateMessageReceived func(message.SessionUpdateMessage)

	// Announce
	AnnouncePleaseMessageReceived func(message.AnnouncePleaseMessage)
	AnnounceMessageReceived       func(message.AnnounceMessage)

	// Subscribe
	SubscribeMessageReceived       func(message.SubscribeMessage)
	SubscribeOkMessageReceived     func(message.SubscribeOkMessage)
	SubscribeUpdateMessageReceived func(message.SubscribeUpdateMessage)

	// Group
	GroupMessageReceived func(message.GroupMessage)

	// Frame
	FrameMessageReceived func(frameCount, byteCount uint64)
}

// Default functions for StreamTracer function fields

// DefaultStreamFinished is the default implementation for StreamClosed
var DefaultStreamFinished = func() {
	// Default implementation: no-op
}

// DefaultStreamReset is the default implementation for SendStreamCancelled
var DefaultStreamReset = func(code quic.StreamErrorCode, reason string) {
	// Default implementation: no-op
}

// DefaultStreamStopped is the default implementation for ReceiveStreamCancelled
var DefaultStreamStopped = func(code quic.StreamErrorCode, reason string) {
	// Default implementation: no-op
}

// DefaultStreamTypeMessageSent is the default implementation for StreamTypeMessageSent
var DefaultStreamTypeMessageSent = func(msg message.StreamTypeMessage) {
	// Default implementation: no-op
}

// DefaultStreamTypeMessageReceived is the default implementation for StreamTypeMessageReceived
var DefaultStreamTypeMessageReceived = func(msg message.StreamTypeMessage) {
	// Default implementation: no-op
}

// DefaultSessionClientMessageSent is the default implementation for SessionClientMessageSent
var DefaultSessionClientMessageSent = func(msg message.SessionClientMessage) {
	// Default implementation: no-op
}

// DefaultSessionClientMessageReceived is the default implementation for SessionClientMessageReceived
var DefaultSessionClientMessageReceived = func(msg message.SessionClientMessage) {
	// Default implementation: no-op
}

// DefaultSessionServerMessageSent is the default implementation for SessionServerMessageSent
var DefaultSessionServerMessageSent = func(msg message.SessionServerMessage) {
	// Default implementation: no-op
}

// DefaultSessionServerMessageReceived is the default implementation for SessionServerMessageReceived
var DefaultSessionServerMessageReceived = func(msg message.SessionServerMessage) {
	// Default implementation: no-op
}

// DefaultSessionUpdateMessageSent is the default implementation for SessionUpdateMessageSent
var DefaultSessionUpdateMessageSent = func(msg message.SessionUpdateMessage) {
	// Default implementation: no-op
}

// DefaultSessionUpdateMessageReceived is the default implementation for SessionUpdateMessageReceived
var DefaultSessionUpdateMessageReceived = func(msg message.SessionUpdateMessage) {
	// Default implementation: no-op
}

// DefaultAnnouncePleaseMessageSent is the default implementation for AnnouncePleaseMessageSent
var DefaultAnnouncePleaseMessageSent = func(msg message.AnnouncePleaseMessage) {
	// Default implementation: no-op
}

// DefaultAnnouncePleaseMessageReceived is the default implementation for AnnouncePleaseMessageReceived
var DefaultAnnouncePleaseMessageReceived = func(msg message.AnnouncePleaseMessage) {
	// Default implementation: no-op
}

// DefaultAnnounceMessageSent is the default implementation for AnnounceMessageSent
var DefaultAnnounceMessageSent = func(msg message.AnnounceMessage) {
	// Default implementation: no-op
}

// DefaultAnnounceMessageReceived is the default implementation for AnnounceMessageReceived
var DefaultAnnounceMessageReceived = func(msg message.AnnounceMessage) {
	// Default implementation: no-op
}

// DefaultSubscribeMessageSent is the default implementation for SubscribeMessageSent
var DefaultSubscribeMessageSent = func(msg message.SubscribeMessage) {
	// Default implementation: no-op
}

// DefaultSubscribeMessageReceived is the default implementation for SubscribeMessageReceived
var DefaultSubscribeMessageReceived = func(msg message.SubscribeMessage) {
	// Default implementation: no-op
}

// DefaultSubscribeOkMessageSent is the default implementation for SubscribeOkMessageSent
var DefaultSubscribeOkMessageSent = func(msg message.SubscribeOkMessage) {
	// Default implementation: no-op
}

// DefaultSubscribeOkMessageReceived is the default implementation for SubscribeOkMessageReceived
var DefaultSubscribeOkMessageReceived = func(msg message.SubscribeOkMessage) {
	// Default implementation: no-op
}

// DefaultSubscribeUpdateMessageSent is the default implementation for SubscribeUpdateMessageSent
var DefaultSubscribeUpdateMessageSent = func(msg message.SubscribeUpdateMessage) {
	// Default implementation: no-op
}

// DefaultSubscribeUpdateMessageReceived is the default implementation for SubscribeUpdateMessageReceived
var DefaultSubscribeUpdateMessageReceived = func(msg message.SubscribeUpdateMessage) {
	// Default implementation: no-op
}

// DefaultGroupMessageSent is the default implementation for GroupMessageSent
var DefaultGroupMessageSent = func(msg message.GroupMessage) {
	// Default implementation: no-op
}

// DefaultGroupMessageReceived is the default implementation for GroupMessageReceived
var DefaultGroupMessageReceived = func(msg message.GroupMessage) {
	// Default implementation: no-op
}

// DefaultFrameMessageSent is the default implementation for FrameMessageSent
var DefaultFrameMessageSent = func(frameCount, byteCount uint64) {
	// Default implementation: no-op
}

// DefaultFrameMessageReceived is the default implementation for FrameMessageReceived
var DefaultFrameMessageReceived = func(frameCount, byteCount uint64) {
	// Default implementation: no-op
}
