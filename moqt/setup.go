package moqt

import (
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
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

type SetupResponceWriter interface {
	Accept(Version)
	Reject(TerminateError)
	WithExtension(Parameters) SetupResponceWriter
}

type SetupHandler interface {
	HandleSetup(SetupRequest, SetupResponceWriter)
}

var _ SetupResponceWriter = (*defaultSetupResponceWriter)(nil)

type defaultSetupResponceWriter struct {
	errCh  chan error
	once   *sync.Once
	stream SessionStream
	params Parameters
}

func (w defaultSetupResponceWriter) Accept(version Version) {
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Version(version),
		Parameters:      message.Parameters(w.params),
	}

	_, err := w.stream.Write(ssm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SESSION_SERVER message", slog.String("error", err.Error()))
		w.Reject(ErrInternalError)
	}

	w.errCh <- nil

	close(w.errCh)
}

func (w defaultSetupResponceWriter) Reject(err TerminateError) {
	slog.Error(err.Error(), slog.Any("Code", err.TerminateErrorCode()))

	/*
	 * Send the Error
	 */
	w.stream.CancelRead(StreamErrorCode(err.TerminateErrorCode()))
	w.stream.CancelWrite(StreamErrorCode(err.TerminateErrorCode()))

	w.errCh <- err

	close(w.errCh)
}

func (w defaultSetupResponceWriter) WithExtension(params Parameters) SetupResponceWriter {
	return defaultSetupResponceWriter{
		once:   w.once,
		stream: w.stream,
		params: params,
	}
}
