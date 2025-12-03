package bps

type ShiftDetector interface {
	Detect(bps float64) bool
}
