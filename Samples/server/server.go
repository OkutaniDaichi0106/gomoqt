package main

import (
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
)

func main() {
	moqs := moqt.Server{
		Addr: "0.0.0.0:8443",
		QUICConfig: &quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
		SupportedVersions: []moqt.Version{moqt.Devlop},
	}

	moqs.SetCertFiles("localhost.pem", "localhost-key.pem")

	relayer := moqt.Relayer{
		Path: "/webtransport",
		Publisher: moqt.Publisher{
			LocalTrack: []string{"foalk", "japan", "kyoto"},
		},
		Subscriber: moqt.Subscriber{
			RemoteTrack: make([][]string, 1),
		},
	}

	moqs.RunOnQUIC(relayer)
}

/***/
type SetupHandler struct{}

func (SetupHandler) HandleSetup(r moqt.SetupRequest, w moqt.SetupResponceWriter) moqt.TerminateError {
	slog.Info("receive a set-up request",
		slog.Group("request",
			slog.String("path", r.Path),
			slog.Any("versions", r.SupportedVersions),
			slog.Any("parameters", r.Parameters)),
	)

	if !moqt.ContainVersion(moqt.Devlop, r.SupportedVersions) {
		return moqt.ErrInternalError
	}

	w.Accept(moqt.Devlop)

	return nil
}

var defaultRelayHandler = RelayHandler{
	trackManager: trackManager{},
}

var _ moqt.PublisherHandler = (*RelayHandler)(nil)

type RelayHandler struct {
	trackManager trackManager
}

func (rh RelayHandler) HandleInterest(i moqt.Interest, w moqt.AnnounceWriter) {
	if i.TrackPrefix[0] != "foalk" {
		w.Reject(moqt.ErrTrackNotFound)
		slog.Error("rejected the interest", slog.Any("track prefix", i.TrackPrefix))
	}

	go func() {
		for {
			slog.Info("find tracks related to the interest", slog.Any("track namespace", i.TrackPrefix))
			node, ok := rh.trackManager.findTrackNamespace(i.TrackPrefix)

			if !ok {
				slog.Info("track namespace not found")
				time.Sleep(3 * time.Minute)
				continue
			}

			for _, node := range node.tracks {
				w.Announce(node.announcement)
			}

			time.Sleep(5 * time.Minute)
		}
	}()

}

func (RelayHandler) HandleSubscribe(s moqt.Subscription, w moqt.SubscribeResponceWriter) {
	slog.Info("receive a subscribe request",
		slog.Group("subscription",
			slog.Any("subscribe ID", s.SubscribeID),
			slog.Any("track namespace", s.Announcement.TrackNamespace),
			slog.Any("track name", s.TrackName),
			slog.Any("subscriber priority", s.SubscriberPriority),
			slog.Any("group order", s.GroupOrder),
			slog.Any("min", s.MinGroupSequence),
			slog.Any("max", s.MaxGroupSequence),
		),
	)
}

func (RelayHandler) HandleFetch(r moqt.FetchRequest, w moqt.FetchResponceWriter) {
}

func (RelayHandler) HandleInfoRequest(r moqt.InfoRequest, w moqt.InfoWriter) {}
func (RelayHandler) HandleAnnounce(r moqt.Announcement, w moqt.AnnounceResponceWriter) {
}
func (RelayHandler) HandleGroup(moqt.Group) {}
