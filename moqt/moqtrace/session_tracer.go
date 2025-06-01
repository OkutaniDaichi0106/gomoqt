package moqtrace

import (
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func InitSessionTracer(tracer *SessionTracer) {
	if tracer == nil {
		panic("SessionTracer must not be nil")
	}

	if tracer.SessionEstablished == nil {
		tracer.SessionEstablished = DefaultSessionEstablished
	}

	if tracer.SessionTerminated == nil {
		tracer.SessionTerminated = DefaultSessionTerminated
	}

	if tracer.QUICStreamOpened == nil {
		tracer.QUICStreamOpened = DefaultQUICStreamOpened
	}

	if tracer.QUICStreamAccepted == nil {
		tracer.QUICStreamAccepted = DefaultQUICStreamAccepted
	}

	if tracer.QUICUniStreamOpened == nil {
		tracer.QUICUniStreamOpened = DefaultQUICUniStreamOpened
	}

	if tracer.QUICUniStreamAccepted == nil {
		tracer.QUICUniStreamAccepted = DefaultQUICUniStreamAccepted
	}
}

type SessionTracer struct {
	SessionEstablished func(local, remote net.Addr, alpn string, version protocol.Version, extension map[uint64][]byte)
	SessionTerminated  func(reason error)

	// QUIC
	QUICStreamOpened   func(quic.StreamID) *StreamTracer
	QUICStreamAccepted func(quic.StreamID) *StreamTracer

	QUICUniStreamOpened   func(quic.StreamID) *SendStreamTracer
	QUICUniStreamAccepted func(quic.StreamID) *ReceiveStreamTracer
}

// Default functions for SessionTracer function fields

// DefaultSessionEstablished is the default implementation for SessionEstablished
var DefaultSessionEstablished = func(local, remote net.Addr, alpn string, version protocol.Version, extension map[uint64][]byte) {
	// Default implementation: no-op
}

// DefaultSessionTerminated is the default implementation for SessionTerminated
var DefaultSessionTerminated = func(reason error) {
	// Default implementation: no-op
}

// DefaultQUICStreamOpened is the default implementation for QUICStreamOpened
var DefaultQUICStreamOpened = func(streamID quic.StreamID) *StreamTracer {
	// Return a StreamTracer with default functions
	return &StreamTracer{
		SendStreamTracer: SendStreamTracer{
			SendStreamFinished:         DefaultStreamFinished,
			SendStreamStopped:          DefaultStreamStopped,
			SendStreamReset:            DefaultStreamReset,
			StreamTypeMessageSent:      DefaultStreamTypeMessageSent,
			SessionClientMessageSent:   DefaultSessionClientMessageSent,
			SessionServerMessageSent:   DefaultSessionServerMessageSent,
			SessionUpdateMessageSent:   DefaultSessionUpdateMessageSent,
			AnnouncePleaseMessageSent:  DefaultAnnouncePleaseMessageSent,
			AnnounceMessageSent:        DefaultAnnounceMessageSent,
			SubscribeMessageSent:       DefaultSubscribeMessageSent,
			SubscribeOkMessageSent:     DefaultSubscribeOkMessageSent,
			SubscribeUpdateMessageSent: DefaultSubscribeUpdateMessageSent,
			GroupMessageSent:           DefaultGroupMessageSent,
			FrameMessageSent:           DefaultFrameMessageSent,
		},
		ReceiveStreamTracer: ReceiveStreamTracer{
			ReceiveStreamStopped:          DefaultStreamStopped,
			StreamTypeMessageReceived:     DefaultStreamTypeMessageReceived,
			SessionClientMessageReceived:  DefaultSessionClientMessageReceived,
			SessionServerMessageReceived:  DefaultSessionServerMessageReceived,
			SessionUpdateMessageReceived:  DefaultSessionUpdateMessageReceived,
			AnnouncePleaseMessageReceived: DefaultAnnouncePleaseMessageReceived,
			AnnounceMessageReceived:       DefaultAnnounceMessageReceived,
			SubscribeMessageReceived:      DefaultSubscribeMessageReceived,
			SubscribeOkMessageReceived:    DefaultSubscribeOkMessageReceived,
			GroupMessageReceived:          DefaultGroupMessageReceived,
			FrameMessageReceived:          DefaultFrameMessageReceived,
		},
	}
}

// DefaultQUICStreamAccepted is the default implementation for QUICStreamAccepted
var DefaultQUICStreamAccepted = func(streamID quic.StreamID) *StreamTracer {
	// Return a StreamTracer with default functions
	return &StreamTracer{
		SendStreamTracer: SendStreamTracer{
			SendStreamFinished:         DefaultStreamFinished,
			SendStreamStopped:          DefaultStreamStopped,
			SendStreamReset:            DefaultStreamReset,
			StreamTypeMessageSent:      DefaultStreamTypeMessageSent,
			SessionClientMessageSent:   DefaultSessionClientMessageSent,
			SessionServerMessageSent:   DefaultSessionServerMessageSent,
			SessionUpdateMessageSent:   DefaultSessionUpdateMessageSent,
			AnnouncePleaseMessageSent:  DefaultAnnouncePleaseMessageSent,
			AnnounceMessageSent:        DefaultAnnounceMessageSent,
			SubscribeMessageSent:       DefaultSubscribeMessageSent,
			SubscribeOkMessageSent:     DefaultSubscribeOkMessageSent,
			SubscribeUpdateMessageSent: DefaultSubscribeUpdateMessageSent,
			GroupMessageSent:           DefaultGroupMessageSent,
			FrameMessageSent:           DefaultFrameMessageSent,
		},
		ReceiveStreamTracer: ReceiveStreamTracer{
			ReceiveStreamFinished:         DefaultStreamFinished,
			ReceiveStreamStopped:          DefaultStreamStopped,
			ReceiveStreamReset:            DefaultStreamReset,
			StreamTypeMessageReceived:     DefaultStreamTypeMessageReceived,
			SessionClientMessageReceived:  DefaultSessionClientMessageReceived,
			SessionServerMessageReceived:  DefaultSessionServerMessageReceived,
			SessionUpdateMessageReceived:  DefaultSessionUpdateMessageReceived,
			AnnouncePleaseMessageReceived: DefaultAnnouncePleaseMessageReceived,
			AnnounceMessageReceived:       DefaultAnnounceMessageReceived,
			SubscribeMessageReceived:      DefaultSubscribeMessageReceived,
			SubscribeOkMessageReceived:    DefaultSubscribeOkMessageReceived,
			GroupMessageReceived:          DefaultGroupMessageReceived,
			FrameMessageReceived:          DefaultFrameMessageReceived,
		},
	}
}

var DefaultQUICUniStreamOpened = func(streamID quic.StreamID) *SendStreamTracer {
	// Return a StreamTracer with default functions
	return &SendStreamTracer{
		SendStreamFinished:         DefaultStreamFinished,
		SendStreamReset:            DefaultStreamReset,
		StreamTypeMessageSent:      DefaultStreamTypeMessageSent,
		SessionClientMessageSent:   DefaultSessionClientMessageSent,
		SessionServerMessageSent:   DefaultSessionServerMessageSent,
		SessionUpdateMessageSent:   DefaultSessionUpdateMessageSent,
		AnnouncePleaseMessageSent:  DefaultAnnouncePleaseMessageSent,
		AnnounceMessageSent:        DefaultAnnounceMessageSent,
		SubscribeMessageSent:       DefaultSubscribeMessageSent,
		SubscribeOkMessageSent:     DefaultSubscribeOkMessageSent,
		SubscribeUpdateMessageSent: DefaultSubscribeUpdateMessageSent,
		GroupMessageSent:           DefaultGroupMessageSent,
		FrameMessageSent:           DefaultFrameMessageSent,
	}
}

var DefaultQUICUniStreamAccepted = func(streamID quic.StreamID) *ReceiveStreamTracer {
	// Return a StreamTracer with default functions
	return &ReceiveStreamTracer{
		ReceiveStreamStopped:          DefaultStreamStopped,
		StreamTypeMessageReceived:     DefaultStreamTypeMessageReceived,
		SessionClientMessageReceived:  DefaultSessionClientMessageReceived,
		SessionServerMessageReceived:  DefaultSessionServerMessageReceived,
		SessionUpdateMessageReceived:  DefaultSessionUpdateMessageReceived,
		AnnouncePleaseMessageReceived: DefaultAnnouncePleaseMessageReceived,
		AnnounceMessageReceived:       DefaultAnnounceMessageReceived,
		SubscribeMessageReceived:      DefaultSubscribeMessageReceived,
		SubscribeOkMessageReceived:    DefaultSubscribeOkMessageReceived,
		GroupMessageReceived:          DefaultGroupMessageReceived,
		FrameMessageReceived:          DefaultFrameMessageReceived,
	}
}
