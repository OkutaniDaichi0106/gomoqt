package main

import (
	"context"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func NewCluster(urlstr string, ctx context.Context) (*Cluster, error) {

	client := moqt.Client{}

	sess, _, err := client.Dial(urlstr, ctx)
	if err != nil {
		return nil, err
	}

	cluster := &Cluster{
		urlstr:  urlstr,
		session: sess,
	}

	annstr, err := sess.OpenAnnounceStream(&moqt.AnnounceConfig{
		TrackPattern: "/**",
	})
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			anns, err := annstr.ReceiveAnnouncements(ctx)
			if err != nil {
				return
			}

			for _, ann := range anns {
				cluster.mux.Handle(string(ann.TrackPath))

			}
		}
	}()

	return cluster, nil
}

var _ moqt.TrackHandler = (*Cluster)(nil)

type Cluster struct {
	urlstr string

	session moqt.Session

	mux *moqt.TrackMux
}
