package moqtrace

var DefaultSessionTracer = func() *SessionTracer {
	return &SessionTracer{
		SessionEstablished: DefaultSessionEstablished,
		SessionTerminated:  DefaultSessionTerminated,
		QUICStreamOpened:   DefaultQUICStreamOpened,
		QUICStreamAccepted: DefaultQUICStreamAccepted,
	}
}
