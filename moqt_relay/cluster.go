package moqtrelay

import "github.com/OkutaniDaichi0106/gomoqt/moqt"

type Cluster struct{}

type Origins struct {
	src    moqt.AnnouncementReader
	routes map[moqt.TrackPath][]moqt.Session
}

type Route struct {
}
