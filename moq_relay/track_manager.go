package moqrelay

import (
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransfork"
)

type TrackManager interface {
	//
	ServeAnnouncements([]moqtransfork.Announcement) error

	ServeTrack(moqtransfork.Subscription, *TrackBuffer) error
}

var _ TrackManager = (*trackManager)(nil)

type trackManager struct {
	//
	trackTree *trackTree
}

func (manager *trackManager) ServeAnnouncements(ann []moqtransfork.Announcement) error {
	// Serve announcements
	for _, a := range ann {
		annBufs := make([]announcementBuffer, 0)

		// Serve announcements to track prefix nodes
		err := manager.trackTree.handleDescendants(strings.Split(a.TrackPath, "/"), func(node *trackPrefixNode) error {
			// Add announcement to the buffer
			node.announcementBuffer.Add(a)

			/*
			 * Handle the track node
			 */
			// Initialize track node if the track prefix matches the announcement track path
			if node.trackPrefix == a.TrackPath {
				switch a.AnnounceStatus {
				case moqtransfork.ACTIVE:
					if node.track == nil {
						node.initTrack()
					}

					// Set announcement
					node.track.mu.Lock()

					node.track.announcement = a

					node.track.mu.Unlock()
				case moqtransfork.ENDED:
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
		if a.AnnounceStatus == moqtransfork.ENDED {
			defer func() {
				slog.Debug("removing track prefix", slog.String("track path", a.TrackPath))

				err := manager.trackTree.removeTrackPrefix(strings.Split(a.TrackPath, "/"))
				if err != nil {
					slog.Error("failed to remove track prefix", slog.String("error", err.Error()))
				}

				slog.Debug("removed track prefix", slog.String("track path", a.TrackPath))
			}()
		}

		slog.Debug("served an announcement", slog.String("track path", a.TrackPath))
	}

	slog.Debug("Successfully served announcements")

	return nil
}

func (manager *trackManager) ServeTrack(sub moqtransfork.Subscription, trackBuf *TrackBuffer) error {
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
