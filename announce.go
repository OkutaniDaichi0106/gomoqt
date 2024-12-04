package moqt

import (
	"errors"
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type Interest struct {
	TrackPrefix string
	Parameters  Parameters
}

type InterestHandler interface {
	HandleInterest(Interest, AnnounceWriter)
}

type Announcement struct {
	TrackPath         string
	AuthorizationInfo string
	Parameters        Parameters
}

type AnnounceStream struct {
	stream moq.Stream
}

func (a AnnounceStream) ReadAnnouncement() (Announcement, error) {
	return readAnnouncement(a.stream)
}

func (a AnnounceStream) Reject(err error) {
	if err == nil {
		a.Close()
	}

	annerr, ok := err.(AnnounceError)
	if !ok {
		annerr = ErrInternalError
	}

	slog.Info("trying to close an Announce Stream", slog.String("reason", annerr.Error()))

	a.stream.CancelWrite(moq.StreamErrorCode(annerr.AnnounceErrorCode()))
	a.stream.CancelRead(moq.StreamErrorCode(annerr.AnnounceErrorCode()))
}

func (a AnnounceStream) Close() {
	err := a.stream.Close()
	if err != nil {
		slog.Error("failed to close the stream", slog.String("error", err.Error()))
		return
	}
}

func readAnnouncement(r io.Reader) (Announcement, error) {
	// Read an ANNOUNCE message
	var am message.AnnounceMessage
	err := am.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Initialize an Announcement
	announcement := Announcement{
		TrackPath:  am.TrackPath,
		Parameters: Parameters(am.Parameters),
	}

	//
	authInfo, ok := getAuthorizationInfo(announcement.Parameters)
	if ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}

type AnnounceWriter struct {
	stream moq.Stream
}

func (w AnnounceWriter) Announce(announcement Announcement) {
	// Add AUTHORIZATION_INFO parameter
	if announcement.AuthorizationInfo != "" {
		announcement.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	// Initialize an ANNOUNCE message
	am := message.AnnounceMessage{
		TrackPath:  announcement.TrackPath,
		Parameters: message.Parameters(announcement.Parameters),
	}

	// Encode the ANNOUNCE message
	err := am.Encode(w.stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
		return
	}

	slog.Info("Successfully announced", slog.Any("announcement", announcement))
}

func (aw AnnounceWriter) Reject(err error) {
	if err == nil {
		aw.Close()
	}

	var code moq.StreamErrorCode

	var strerr moq.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		annerr, ok := err.(AnnounceError)
		if ok {
			code = moq.StreamErrorCode(annerr.AnnounceErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	aw.stream.CancelRead(code)
	aw.stream.CancelWrite(code)

	slog.Info("closed an Announce Stream")
}

func (aw AnnounceWriter) Close() {
	err := aw.stream.Close()
	if err != nil {
		slog.Error("catch an erro when closing an Announce Stream", slog.String("error", err.Error()))
	}
}
