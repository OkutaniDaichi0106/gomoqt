package moqt

import (
	"log/slog"
	"strings"
)

type TrackManager interface {
	//
	ServeAnnouncements([]Announcement) error

	ServeTrack(Subscription, *TrackBuffer) error
}

var _ TrackManager = (*trackManager)(nil)

type trackManager struct {
	//
	trackTree *trackTree
}

func (manager *trackManager) ServeAnnouncements(ann []Announcement) error {
	annBufs := make([]announcementBuffer, 0)

	// Serve announcements
	for _, a := range ann {
		err := manager.trackTree.handleDescendants(strings.Split(a.TrackPath, "/"), func(node *trackPrefixNode) error {
			// Add announcement to the buffer
			node.announcementBuffer.Add(a)

			// Initialize track node if the track prefix matches the announcement track path
			if node.trackPrefix == a.TrackPath {
				if node.track == nil {
					node.initTrack()
				}

				// Set announcement
				node.track.mu.Lock()

				node.track.announcement = a

				node.track.mu.Unlock()
			}

			// Register the announcement buffer to broadcast
			annBufs = append(annBufs, node.announcementBuffer)

			return nil
		})

		slog.Error("failed to serve announcements", slog.String("error", err.Error()))
	}

	// Broadcast
	for _, annBuf := range annBufs {
		annBuf.Broadcast()
	}

	return nil
}

func (manager *trackManager) ServeTrack(sub Subscription, trackBuf *TrackBuffer) error {
	node, ok := manager.trackTree.traceTrackPrefix(strings.Split(sub.TrackPath, "/"))
	if !ok {
		// Insert the track prefix to the track tree
		node = manager.trackTree.insertTrackPrefix(strings.Split(sub.TrackPath, "/"))
	}

	// Initialize track node
	if node.track == nil {
		node.initTrack()
	}

	// Serve subscription and track buffer
	node.track.mu.Lock()

	node.track.subscription = sub
	node.track.trackBuf = trackBuf

	node.track.mu.Unlock()

	return nil
}
