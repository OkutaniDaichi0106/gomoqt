package moqt

import "time"

type JitterManager interface {
	NextFrame() ([]byte, error)
	AddFrame()
	MaxLatency() time.Duration
} //TODO
