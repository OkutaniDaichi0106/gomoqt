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
}

type SessionTracer struct {
	SessionEstablished func(local, remote net.Addr, alpn string, version protocol.Version, extension map[uint64][]byte)
	SessionTerminated  func(reason error)

	// QUIC
	QUICStreamOpened   func(quic.StreamID) *StreamTracer
	QUICStreamAccepted func(quic.StreamID) *StreamTracer
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
		StreamClosed:                  DefaultStreamClosed,
		SendStreamCancelled:           DefaultSendStreamCancelled,
		ReceiveStreamCancelled:        DefaultReceiveStreamCancelled,
		StreamTypeMessageSent:         DefaultStreamTypeMessageSent,
		StreamTypeMessageReceived:     DefaultStreamTypeMessageReceived,
		SessionClientMessageSent:      DefaultSessionClientMessageSent,
		SessionClientMessageReceived:  DefaultSessionClientMessageReceived,
		SessionServerMessageSent:      DefaultSessionServerMessageSent,
		SessionServerMessageReceived:  DefaultSessionServerMessageReceived,
		SessionUpdateMessageSent:      DefaultSessionUpdateMessageSent,
		SessionUpdateMessageReceived:  DefaultSessionUpdateMessageReceived,
		AnnouncePleaseMessageSent:     DefaultAnnouncePleaseMessageSent,
		AnnouncePleaseMessageReceived: DefaultAnnouncePleaseMessageReceived,
		AnnounceMessageSent:           DefaultAnnounceMessageSent,
		AnnounceMessageReceived:       DefaultAnnounceMessageReceived,
		SubscribeMessageSent:          DefaultSubscribeMessageSent,
		SubscribeMessageReceived:      DefaultSubscribeMessageReceived,
		SubscribeOkMessageSent:        DefaultSubscribeOkMessageSent,
		SubscribeOkMessageReceived:    DefaultSubscribeOkMessageReceived,
		GroupMessageSent:              DefaultGroupMessageSent,
		GroupMessageReceived:          DefaultGroupMessageReceived,
		FrameMessageSent:              DefaultFrameMessageSent,
		FrameMessageReceived:          DefaultFrameMessageReceived,
	}
}

// DefaultQUICStreamAccepted is the default implementation for QUICStreamAccepted
var DefaultQUICStreamAccepted = func(streamID quic.StreamID) *StreamTracer {
	// Return a StreamTracer with default functions
	return &StreamTracer{
		StreamClosed:                  DefaultStreamClosed,
		SendStreamCancelled:           DefaultSendStreamCancelled,
		ReceiveStreamCancelled:        DefaultReceiveStreamCancelled,
		StreamTypeMessageSent:         DefaultStreamTypeMessageSent,
		StreamTypeMessageReceived:     DefaultStreamTypeMessageReceived,
		SessionClientMessageSent:      DefaultSessionClientMessageSent,
		SessionClientMessageReceived:  DefaultSessionClientMessageReceived,
		SessionServerMessageSent:      DefaultSessionServerMessageSent,
		SessionServerMessageReceived:  DefaultSessionServerMessageReceived,
		SessionUpdateMessageSent:      DefaultSessionUpdateMessageSent,
		SessionUpdateMessageReceived:  DefaultSessionUpdateMessageReceived,
		AnnouncePleaseMessageSent:     DefaultAnnouncePleaseMessageSent,
		AnnouncePleaseMessageReceived: DefaultAnnouncePleaseMessageReceived,
		AnnounceMessageSent:           DefaultAnnounceMessageSent,
		AnnounceMessageReceived:       DefaultAnnounceMessageReceived,
		SubscribeMessageSent:          DefaultSubscribeMessageSent,
		SubscribeMessageReceived:      DefaultSubscribeMessageReceived,
		SubscribeOkMessageSent:        DefaultSubscribeOkMessageSent,
		SubscribeOkMessageReceived:    DefaultSubscribeOkMessageReceived,
		GroupMessageSent:              DefaultGroupMessageSent,
		GroupMessageReceived:          DefaultGroupMessageReceived,
		FrameMessageSent:              DefaultFrameMessageSent,
		FrameMessageReceived:          DefaultFrameMessageReceived,
	}
}
