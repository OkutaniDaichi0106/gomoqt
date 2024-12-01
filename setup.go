package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type SessionStream moq.Stream //TODO:

/*
 *
 */
type SetupRequest struct {
	SupportedVersions []Version
	Path              string // TODO:
	MaxSubscribeID    uint64 // TODO:
	Parameters        Parameters
}

/*
 * Server
 */
type SetupResponce struct {
	SelectedVersion Version
	Parameters      Parameters
}

// type SetupHandler interface {
// 	HandleSetup(SetupRequest, SetupResponceWriter)
// }

// type SetupHandlerFunc func(SetupRequest, SetupResponceWriter)

// func (f SetupHandlerFunc) HandleSetup(r SetupRequest, w SetupResponceWriter) {
// 	f(r, w)
// }

// type SetupResponceWriter struct {
// 	doneCh chan struct{}
// 	once   *sync.Once
// 	stream SessionStream
// 	params Parameters
// }

// func (w SetupResponceWriter) Accept(version Version) {
// 	ssm := message.SessionServerMessage{
// 		SelectedVersion: protocol.Version(version),
// 		Parameters:      message.Parameters(w.params),
// 	}

// 	err := ssm.Encode(w.stream)
// 	if err != nil {
// 		slog.Error("failed to send a SESSION_SERVER message", slog.String("error", err.Error()))
// 		w.Reject(ErrInternalError)
// 	}

// 	w.doneCh <- struct{}{}

// 	close(w.doneCh)
// }

// func (w SetupResponceWriter) Reject(err TerminateError) {
// 	slog.Error(err.Error(), slog.Any("Code", err.TerminateErrorCode()))

// 	/*
// 	 * Send the Error
// 	 */
// 	w.stream.CancelRead(moq.StreamErrorCode(err.TerminateErrorCode()))
// 	w.stream.CancelWrite(moq.StreamErrorCode(err.TerminateErrorCode()))

// 	w.doneCh <- struct{}{}

// 	close(w.doneCh)
// }

// func (w SetupResponceWriter) WithExtension(params Parameters) SetupResponceWriter {
// 	return SetupResponceWriter{
// 		once:   w.once,
// 		stream: w.stream,
// 		params: params,
// 	}
// }
