package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Interest message.AnnounceInterestMessage

type InterestWriter interface {
	Interest(Interest) error
}

type AnnounceWriter interface {
	Announce(Announcement)
	Reject(InterestError)
	New(Stream) AnnounceWriter
}

type InterestHandler interface {
	HandleInterest(Interest, AnnounceWriter)
}

type Announcement struct {
	TrackNamespace    []string
	AuthorizationInfo string
	Parameters        message.Parameters
}

type AnnounceResponceWriter interface {
	Accept(Announcement)
	Reject(AnnounceError)
}

type AnnounceHandler interface {
	HandleAnnounce(Announcement, AnnounceResponceWriter)
}

var _ AnnounceResponceWriter = (*defaultAnnounceResponceWriter)(nil)

type defaultAnnounceResponceWriter struct {
	stream Stream
}

func (defaultAnnounceResponceWriter) Accept(Announcement) {

}

func (w defaultAnnounceResponceWriter) Reject(err AnnounceError) {
	w.stream.CancelRead(StreamErrorCode(err.AnnounceErrorCode()))
	w.stream.CancelWrite(StreamErrorCode(err.AnnounceErrorCode()))

	slog.Info("reject the interest", slog.String("error", err.Error()))
}

type AnnounceReader interface {
	Read(r quicvarint.Reader) (Announcement, error)
}

var _ InterestWriter = (*defaultInterestWriter)(nil)

type defaultInterestWriter struct {
	stream Stream
}

func (w defaultInterestWriter) Interest(interest Interest) error {
	aim := message.AnnounceInterestMessage{
		TrackPrefix: interest.TrackPrefix,
		Parameters:  interest.Parameters,
	}

	_, err := w.stream.Write(aim.SerializePayload())
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Interested", slog.Any("track prefix", interest.TrackPrefix))
	return nil
}

/*
 * MOQTransfork implementation
 */
var _ AnnounceWriter = (*defaultAnnounceWriter)(nil)

type defaultAnnounceWriter struct {
	stream Stream
}

func (irw defaultAnnounceWriter) New(stream Stream) AnnounceWriter {
	return defaultAnnounceWriter{
		stream: stream,
	}
}

func (irw defaultAnnounceWriter) Announce(announcement Announcement) {
	am := message.AnnounceMessage{
		TrackNamespace: announcement.TrackNamespace,
		Parameters:     announcement.Parameters,
	}

	if announcement.AuthorizationInfo != "" {
		am.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	_, err := irw.stream.Write(am.SerializePayload())
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
	}

	slog.Info("announced", slog.Any("track namespace", announcement.TrackNamespace))
}

func (w defaultAnnounceWriter) Reject(err InterestError) {
	w.stream.CancelRead(StreamErrorCode(err.InterestErrorCode()))
	w.stream.CancelWrite(StreamErrorCode(err.InterestErrorCode()))

	slog.Info("reject the interest", slog.String("error", err.Error()))
}
