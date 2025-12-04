package bitrate

type ShiftDetector interface {
	Detect(rate float64) bool
}
