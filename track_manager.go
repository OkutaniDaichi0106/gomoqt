package moqt

import (
	"sync"
	"time"
)

type TrackManager struct {
	mu sync.RWMutex

	/*
	 * The keys are Track Namespace
	 */
	announcements map[string]*announcementNode
}

func (tm *TrackManager) NewAnnouncement(a Announcement) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.announcements == nil {
		tm.announcements = make(map[string]*announcementNode)
	}

	_, ok := tm.announcements[a.TrackPath]
	if !ok {
		tm.announcements[a.TrackPath] = &announcementNode{
			announcement: a,
			tracks:       make(map[string]trackConfig),
		}
	}
}

type announcementNode struct {
	mu sync.RWMutex

	/*
	 *
	 */
	announcement Announcement

	/*
	 * The keys are Track Names
	 */
	tracks map[string]trackConfig
}

type trackConfig struct {
	TrackPath         string
	MinGroupSequence  GroupSequence
	MaxGroupSequence  GroupSequence
	PublisherPriority PublisherPriority
	GroupOrder        GroupOrder
	GroupExpires      time.Duration
}
