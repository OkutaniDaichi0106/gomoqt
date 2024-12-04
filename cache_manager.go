package moqt

import (
	"io"
	"time"
)

type CacheManager interface {
	GetFrame(string, string, GroupSequence, FrameSequence) (io.Reader, error)
	GetGroup(string, string, GroupSequence) (io.Reader, error)
	GroupExpires(string, string, GroupSequence) time.Time
}

//TODO:
