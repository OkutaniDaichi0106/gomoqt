package bitrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEWMAShiftDetector(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 5)

	assert.NotNil(t, detector)
	assert.Equal(t, 0.2, detector.alpha)
	assert.Equal(t, 0.3, detector.threshold)
	assert.Equal(t, 5, detector.minSamples)
}

func TestEWMAShiftDetector_Detect_InitialSamples(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 3)

	// First 3 samples should return false (minSamples = 3)
	assert.False(t, detector.Detect(1000), "first sample should return false")
	assert.False(t, detector.Detect(2000), "second sample should return false")
	assert.False(t, detector.Detect(3000), "third sample should return false")

	// After minSamples, detection should work
	// The average is set to the last sample (3000) during initialization
	// A value significantly different should be detected
	result := detector.Detect(10000) // Much higher than 3000
	// This depends on the threshold, but with 0.3 threshold, 10000 > 3000*1.3 should be true
	assert.True(t, result, "large deviation after initial samples should be detected")
}

func TestEWMAShiftDetector_Detect_NoShiftWithinThreshold(t *testing.T) {
	detector := NewEWMAShiftDetector(0.5, 0.3, 0) // alpha=0.5, threshold=30%, no initial samples

	// Set initial average
	detector.Detect(1000)

	// Values within 30% threshold should not trigger detection
	// Average after first detect: 0.5*1000 + 0.5*0 = 500 (but initial average is 0)
	// Let's reset and test properly
	detector = NewEWMAShiftDetector(0.5, 0.3, 0)
	detector.average = 1000 // Set average directly for testing

	// 1100 is within 30% of 1000 (700-1300 range)
	assert.False(t, detector.Detect(1100), "value within threshold should not trigger")

	// 900 is within 30% of average
	assert.False(t, detector.Detect(900), "value within threshold should not trigger")
}

func TestEWMAShiftDetector_Detect_ShiftAboveThreshold(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 0)
	detector.average = 1000 // Set average directly

	// Value more than 30% above average (1300+) should trigger
	assert.True(t, detector.Detect(1500), "value above threshold should trigger")
}

func TestEWMAShiftDetector_Detect_ShiftBelowThreshold(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 0)
	detector.average = 1000 // Set average directly

	// Value more than 30% below average (below 700) should trigger
	assert.True(t, detector.Detect(500), "value below threshold should trigger")
}

func TestEWMAShiftDetector_Detect_EWMACalculation(t *testing.T) {
	// alpha = 0.5 for easy calculation
	detector := NewEWMAShiftDetector(0.5, 0.3, 0)
	detector.average = 1000

	// After Detect(1000), average should still be ~1000
	detector.Detect(1000)
	// new average = 0.5*1000 + 0.5*1000 = 1000
	assert.InDelta(t, 1000, detector.average, 0.01)

	// After Detect(2000), average should be 1500
	detector.Detect(2000)
	// new average = 0.5*2000 + 0.5*1000 = 1500
	assert.InDelta(t, 1500, detector.average, 0.01)
}

func TestEWMAShiftDetector_Detect_MinSamplesDecrement(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 3)

	assert.Equal(t, 3, detector.minSamples)

	detector.Detect(1000)
	assert.Equal(t, 2, detector.minSamples)

	detector.Detect(1000)
	assert.Equal(t, 1, detector.minSamples)

	detector.Detect(1000)
	assert.Equal(t, 0, detector.minSamples)

	// After this, minSamples should stay at 0
	detector.Detect(1000)
	assert.Equal(t, 0, detector.minSamples)
}

func TestEWMAShiftDetector_Detect_AverageSetDuringInitialSamples(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 2)

	detector.Detect(1000)
	assert.Equal(t, float64(1000), detector.average, "average should be set to bps during initial samples")

	detector.Detect(2000)
	assert.Equal(t, float64(2000), detector.average, "average should be set to latest bps during initial samples")
}

func TestEWMAShiftDetector_ImplementsInterface(t *testing.T) {
	var _ ShiftDetector = (*EWMAShiftDetector)(nil)
}

func TestEWMAShiftDetector_Detect_ZeroBPS(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 0)
	detector.average = 1000

	// Zero BPS should be detected as a significant drop
	assert.True(t, detector.Detect(0), "zero BPS should trigger detection when average is non-zero")
}

func TestEWMAShiftDetector_Detect_ZeroAverage(t *testing.T) {
	detector := NewEWMAShiftDetector(0.2, 0.3, 0)
	detector.average = 0

	// When average is 0, any positive BPS is infinitely above threshold
	// The check is: bps > average*(1+threshold) = 0*1.3 = 0
	// So 100 > 0 should be true
	result := detector.Detect(100)
	assert.True(t, result, "positive BPS should trigger when average is zero")
}
