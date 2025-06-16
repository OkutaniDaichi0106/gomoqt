package main

import (
	"context"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func NewCluster(urlstr string, ctx context.Context) (*Cluster, error) {

	client := moqt.Client{}

	sess, err := client.Dial(ctx, urlstr, nil)
	if err != nil {
		return nil, err
	}

	cluster := &Cluster{
		urlstr:   urlstr,
		upstream: sess,
		mux:      moqt.NewTrackMux(),
	}

	annstr, err := sess.OpenAnnounceStream("/")
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			ann, err := annstr.ReceiveAnnouncement(ctx)
			if err != nil {
				return
			}

			cluster.mux.Announce(ann, nil, nil)
		}
	}()

	return cluster, nil
}

var _ moqt.TrackHandler = (*Cluster)(nil)

type Cluster struct {
	urlstr string

	upstream *moqt.Session

	mux *moqt.TrackMux
}

func (c *Cluster) ServeTrack(pub *moqt.Publisher) {

}
