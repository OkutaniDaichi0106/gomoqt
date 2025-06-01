package moqtrace

var DefaultTracer = func() *SessionTracer {
	return &SessionTracer{
		SessionEstablished: DefaultSessionEstablished,
		SessionTerminated:  DefaultSessionTerminated,
		QUICStreamOpened:   DefaultQUICStreamOpened,
		QUICStreamAccepted: DefaultQUICStreamAccepted,
	}
}
