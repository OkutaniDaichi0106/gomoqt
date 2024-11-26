package main

import (
	"context"
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	handler := requestHandler{}

	c := moqt.Client{
		URL: "https://localhost:8443/path",
		DataHandler: moqt.DataHandlerFunc(func(g moqt.Group, rs moqt.ReceiveStream) {
			buf := make([]byte, 1<<5)

			for {
				_, err := rs.Read(buf)
				if err != nil {
					log.Print(err)
					return
				}

				log.Println(string(buf))
			}
		}),
		RequestHandler: handler,
	}

	err := c.Run(context.Background())
	if err != nil {
		log.Print(err)
		return
	}

}

var _ moqt.RequestHandler = (*requestHandler)(nil)

type requestHandler struct {
	localTrack moqt.Announcement
}

func (requestHandler) HandleFetch(r moqt.FetchRequest, w moqt.FetchResponceWriter) {
	w.Reject(nil)
}

func (requestHandler) HandleInfoRequest(r moqt.InfoRequest, i *moqt.Info, w moqt.InfoWriter) {
	if i == nil {
		w.Reject(moqt.ErrNoGroup)
	}
}

func (h requestHandler) HandleInterest(i moqt.Interest, a []moqt.Announcement, w moqt.AnnounceWriter) {
	h.localTrack = moqt.Announcement{
		TrackNamespace: i.TrackPrefix + "room-0x000001/user-0x000001",
	}

	w.Announce(h.localTrack)

	w.Close(nil)
}

func (h requestHandler) HandleSubscribe(s moqt.Subscription, i *moqt.Info, w moqt.SubscribeResponceWriter) {
	if h.localTrack.TrackNamespace != s.TrackNamespace {
		w.Reject(nil)
	}

	if i == nil {
		w.Reject(moqt.ErrTrackDoesNotExist)
	}

	w.Accept(*i)
}
