package moqt

import (
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

type SessionStream Stream

type SetupRequest struct {
	Path              string
	SupportedVersions []Version
	Parameters        Parameters
}

/*
 *
 */

/*
 * Server
 */
type SetupResponce struct {
	SelectedVersion Version
	Parameters      Parameters
}

type SetupHandler interface {
	HandleSetup(SetupRequest, SetupResponceWriter)
}

type SetupHandlerFunc func(SetupRequest, SetupResponceWriter)

func (f SetupHandlerFunc) HandleSetup(r SetupRequest, w SetupResponceWriter) {
	f(r, w)
}

type SetupResponceWriter struct {
	doneCh chan struct{}
	once   *sync.Once
	stream SessionStream
	params Parameters
}

func (w SetupResponceWriter) Accept(version Version) {
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Version(version),
		Parameters:      message.Parameters(w.params),
	}

	_, err := w.stream.Write(ssm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SESSION_SERVER message", slog.String("error", err.Error()))
		w.Reject(ErrInternalError)
	}

	w.doneCh <- struct{}{}

	close(w.doneCh)
}

func (w SetupResponceWriter) Reject(err TerminateError) {
	slog.Error(err.Error(), slog.Any("Code", err.TerminateErrorCode()))

	/*
	 * Send the Error
	 */
	w.stream.CancelRead(StreamErrorCode(err.TerminateErrorCode()))
	w.stream.CancelWrite(StreamErrorCode(err.TerminateErrorCode()))

	w.doneCh <- struct{}{}

	close(w.doneCh)
}

func (w SetupResponceWriter) WithExtension(params Parameters) SetupResponceWriter {
	return SetupResponceWriter{
		once:   w.once,
		stream: w.stream,
		params: params,
	}
}
