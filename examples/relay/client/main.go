package main

import (
	"context"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	client := moqt.Client{}

	sess, err := client.Dial(context.Background(), "https://localhost:8080", nil)
	if err != nil {
		return
	}
	annstr, err := sess.OpenAnnounceStream("/")
	if err != nil {
		return
	}
	for {
		ann, err := annstr.ReceiveAnnouncement(context.Background())
		if err != nil {
			return
		}

		go func(ann *moqt.Announcement) {
			sess.OpenTrackStream(ann.BroadcastPath(), "", nil)
		}(ann)
	}
}
