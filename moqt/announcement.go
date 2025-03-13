package moqt

import (
	"errors"
	"strings"
	"sync"
)

func NewAnnouncement(trackPath TrackPath) *Announcement {
	ann := Announcement{
		active: true,
		cond:   sync.NewCond(&sync.Mutex{}),
		path:   trackPath,
	}

	return &ann
}

type Announcement struct {
	active bool
	cond   *sync.Cond

	/*
	 *
	 */
	path TrackPath
}

func (a Announcement) String() string {
	var sb strings.Builder
	sb.WriteString("Announcement: {")
	sb.WriteString(" ")
	sb.WriteString("AnnounceStatus: ")
	if a.active {
		sb.WriteString("ACTIVE")
	} else {
		sb.WriteString("ENDED")
	}
	sb.WriteString(", ")
	sb.WriteString("TrackPath: ")
	sb.WriteString(a.path.String())
	sb.WriteString(" }")
	return sb.String()
}

func (a *Announcement) TrackPath() TrackPath {
	return a.path
}

func (a *Announcement) End() error {
	a.cond.L.Lock()
	defer a.cond.L.Unlock()

	if !a.active {
		return errors.New("announcement has already ended")
	}

	a.active = false

	// notify any waiting goroutines that the announcement has ended
	a.cond.Broadcast()

	return nil
}

func (a *Announcement) AwaitEnd() {
	a.cond.L.Lock()
	defer a.cond.L.Unlock()

	for a.active {
		a.cond.Wait()
	}
}

func (a Announcement) IsActive() bool {
	return a.active
}
