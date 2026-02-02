package processor

import (
	"math"
)

// AddDB Increase volume
func AddDB(samples []float32, dB float64) []float32 {
	// Formula : gain = 10^(dB/20)
	gain := math.Pow(10.0, dB/20.0)

	result := make([]float32, len(samples))

	for i, sample := range samples {
		result[i] = sample * float32(gain)

		// Clipping
		if result[i] > 1.0 {
			result[i] = 1.0
		} else if result[i] < -1.0 {
			result[i] = -1.0
		}
	}

	return result
}

// SubDB Decrease volume
func SubDB(samples []float32, dB float64) []float32 {
	// Formula : gain = 10^(dB/20)
	gain := math.Pow(10.0, dB/20.0)

	result := make([]float32, len(samples))

	for i, sample := range samples {
		result[i] = sample / float32(gain)

		// Clipping
		if result[i] > 1.0 {
			result[i] = 1.0
		} else if result[i] < -1.0 {
			result[i] = -1.0
		}
	}

	return result
}
