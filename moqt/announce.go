package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type InterestStream Stream

type Interest struct {
	TrackPrefix []string
	Parameters  Parameters
}

type InterestWriter interface {
	Interest(Interest) error
}

var _ InterestWriter = (*defaultInterestWriter)(nil)

type defaultInterestWriter struct {
	stream InterestStream
}

func (w defaultInterestWriter) Interest(interest Interest) error {
	aim := message.AnnounceInterestMessage{
		TrackPrefix: interest.TrackPrefix,
		Parameters:  message.Parameters(interest.Parameters),
	}

	_, err := w.stream.Write(aim.SerializePayload())
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Interested", slog.Any("track prefix", interest.TrackPrefix))
	return nil
}

type AnnounceWriter interface {
	Announce(Announcement)
	Reject(InterestError)
	New(Stream) AnnounceWriter
}

type InterestHandler interface {
	HandleInterest(Interest, AnnounceWriter)
}

/*
 * MOQTransfork implementation
 */
var _ AnnounceWriter = (*defaultAnnounceWriter)(nil)

type defaultAnnounceWriter struct {
	stream InterestStream
}

func (irw defaultAnnounceWriter) New(stream Stream) AnnounceWriter {
	return defaultAnnounceWriter{
		stream: stream,
	}
}

func (irw defaultAnnounceWriter) Announce(announcement Announcement) {
	am := message.AnnounceMessage{
		TrackNamespace: announcement.TrackNamespace,
		Parameters:     message.Parameters(announcement.Parameters),
	}

	if announcement.AuthorizationInfo != "" {
		am.Parameters.AddParameter(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
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

type InterestReader interface {
	Read(r quicvarint.Reader) (Interest, error)
}

var _ InterestReader = (*defaultInterestReader)(nil)

type defaultInterestReader struct{}

func (defaultInterestReader) Read(r quicvarint.Reader) (Interest, error) {
	var aim message.AnnounceInterestMessage
	err := aim.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return Interest{}, err
	}
	return Interest{
		TrackPrefix: aim.TrackPrefix,
		Parameters:  Parameters(aim.Parameters),
	}, nil
}

type AnnounceStream Stream

type Announcement struct {
	TrackNamespace    []string
	AuthorizationInfo string
	Parameters        Parameters
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
	stream AnnounceStream
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

var _ AnnounceReader = (*defaultAnnouncementReader)(nil)

type defaultAnnouncementReader struct{}

func (defaultAnnouncementReader) Read(r quicvarint.Reader) (Announcement, error) {
	// Read an ANNOUNCE message
	var am message.AnnounceMessage
	err := am.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Return an Announcement
	announcement := Announcement{
		TrackNamespace: am.TrackNamespace,
		Parameters:     Parameters(am.Parameters),
	}

	authInfo, ok := getAuthorizationInfo(am.Parameters)
	if ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}
