package main

import "github.com/OkutaniDaichi0106/gomoqt/moqt"

var _ moqt.TrackHandler = (*Router)(nil)

type Router struct {
	announcement *moqt.Announcement
}

func (r *Router) ServeAnnouncement(w moqt.AnnouncementWriter, config moqt.AnnounceConfig) {
	if r.announcement.TrackPath.HasPrefix(config.TrackPattern) {
		return
	}

	w.WriteAnnouncement([]*moqt.Announcement{r.announcement})
}

func (r *Router) ServeTrack(w moqt.TrackWriter, config moqt.SubscribeConfig) {
	// TODO: Handle track subscription
}

func (r *Router) ServeInfo(ch chan<- moqt.Info, req moqt.InfoRequest) {
	// TODO: Implement
}
