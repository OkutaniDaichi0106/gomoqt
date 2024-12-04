package moqt

import (
	"io"
	"time"
)

type CacheManager interface {
	GetFrame(string, GroupSequence, FrameSequence) (io.Reader, error)
	GetGroup(string, GroupSequence) (io.Reader, error)
	GroupExpires(string, GroupSequence) time.Time
}

//TODO:
