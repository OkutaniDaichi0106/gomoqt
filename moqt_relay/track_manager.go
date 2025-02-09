package moqtrelay

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

type TrackManager interface {
	//
	ServeAnnouncements([]moqt.Announcement) error

	ServeTrack(moqt.SubscribeConfig, *TrackBuffer) error
}

var _ TrackManager = (*trackManager)(nil)

type trackManager struct {
	//
	trackTree *trackTree
}

func (manager *trackManager) ServeAnnouncements(ann []moqt.Announcement) error {
	// Serve announcements
	for _, a := range ann {
		annBufs := make([]announcementBuffer, 0)

		// Serve announcements to track prefix nodes
		err := manager.trackTree.handleDescendants(a.TrackPath, func(node *trackPrefixNode) error {
			// Add announcement to the buffer
			node.announcementBuffer.Add(a)

			/*
			 * Handle the track node
			 */
			// Initialize track node if the track prefix matches the announcement track path
			if moqt.IsSamePath(node.trackPrefix, a.TrackPath) {
				switch a.AnnounceStatus {
				case moqt.ACTIVE:
					if node.track == nil {
						node.initTrack()
					}

					// Set announcement
					node.track.mu.Lock()

					node.track.announcement = a

					node.track.mu.Unlock()
				case moqt.ENDED:
					// Remove the track node
					node.track = nil
				}
			}

			// Register the announcement buffer to broadcast
			annBufs = append(annBufs, node.announcementBuffer)

			return nil
		})

		if err != nil {
			slog.Error("failed to serve announcements", slog.String("error", err.Error()))
		}

		// Broadcast
		for _, annBuf := range annBufs {
			annBuf.Broadcast()
		}

		// Remove the track prefix if the announcement status is ENDED
		if a.AnnounceStatus == moqt.ENDED {
			defer func() {
				slog.Debug("removing track prefix", slog.String("track path", moqt.TrackPartsString(a.TrackPath)))

				err := manager.trackTree.removeTrackPrefix(a.TrackPath)
				if err != nil {
					slog.Error("failed to remove track prefix", slog.String("error", err.Error()))
				}

				slog.Debug("removed track prefix", slog.String("track path", moqt.TrackPartsString(a.TrackPath)))
			}()
		}

		slog.Debug("served an announcement", slog.String("track path", moqt.TrackPartsString(a.TrackPath)))
	}

	slog.Debug("Successfully served announcements")

	return nil
}

func (manager *trackManager) ServeTrack(sub moqt.SubscribeConfig, trackBuf *TrackBuffer) error {
	node, ok := manager.trackTree.traceTrackPrefix(sub.TrackPath)
	if !ok {
		// Insert the track prefix to the track tree
		node = manager.trackTree.insertTrackPrefix(sub.TrackPath)
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
