package bps

var _ ShiftDetector = (*EWMAShiftDetector)(nil)

func NewEWMAShiftDetector(alpha, threshold float64, minSamples int) *EWMAShiftDetector {
	return &EWMAShiftDetector{
		alpha:      alpha,
		threshold:  threshold,
		minSamples: minSamples,
	}
}

type EWMAShiftDetector struct {
	alpha      float64
	average    float64
	threshold  float64
	minSamples int
}

func (d *EWMAShiftDetector) Detect(bps float64) bool {
	// Handle initial samples: if minSamples > 0, decrement and set average to bps, return false
	if d.minSamples > 0 {
		d.minSamples--
		d.average = bps
		return false
	}
	// Update EWMA: calculate new average
	d.average = d.alpha*bps + (1-d.alpha)*d.average
	// Detect shift: if bps is outside the threshold range of average, return true
	if bps > d.average*(1+d.threshold) || bps < d.average*(1-d.threshold) {
		return true
	}
	return false
}
