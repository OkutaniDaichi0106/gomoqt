package moqt

import (
	"strings"
)

func NewAnnouncement(trackPath TrackPath) *Announcement {
	ann := Announcement{
		active:    true,
		TrackPath: trackPath,
	}

	return &ann
}

func newEndedAnnouncement(trackPath TrackPath) *Announcement {
	ann := Announcement{
		active:    false,
		TrackPath: trackPath,
	}

	return &ann
}

type Announcement struct {
	active bool

	/*
	 *
	 */
	TrackPath TrackPath
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
	sb.WriteString(a.TrackPath.String())
	sb.WriteString(" }")
	return sb.String()
}

func (a *Announcement) activate() {
	a.active = true
}

func (a *Announcement) end() {
	a.active = false
}

func (a Announcement) IsActive() bool {
	return a.active
}

func (a Announcement) IsEnded() bool {
	return !a.active
}
