package drift

import "math"

// Detectors holds ready-to-use statistical drift detectors.
// Each operates over a feature window produced by Monitor.

// KolmogorovSmirnov computes the KS statistic between two samples.
func KolmogorovSmirnov(baseline, current []float64) float64 {
	if len(baseline) == 0 || len(current) == 0 {
		return 0
	}
	b := append([]float64{}, baseline...)
	c := append([]float64{}, current...)
	sortFloats(b)
	sortFloats(c)

	var max float64
	bi, ci := 0, 0
	for bi < len(b) && ci < len(c) {
		bv, cv := b[bi], c[ci]
		var d float64
		switch {
		case bv < cv:
			d = cdf(b, bi, len(b)) - cdf(c, ci-1, len(c))
			bi++
		case bv > cv:
			d = cdf(c, ci, len(c)) - cdf(b, bi-1, len(b))
			ci++
		default:
			d = cdf(b, bi, len(b)) - cdf(c, ci, len(c))
			bi++
			ci++
		}
		if d < 0 {
			d = -d
		}
		if d > max {
			max = d
		}
	}
	return max
}

func cdf(sorted []float64, i, n int) float64 {
	if i < 0 {
		return 0
	}
	return float64(i+1) / float64(n)
}

func sortFloats(a []float64) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}

// ZScoreOutliers returns values whose z-score exceeds the threshold.
func ZScoreOutliers(values []float64, threshold float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	var sum, sumSq float64
	for _, v := range values {
		sum += v
		sumSq += v * v
	}
	mean := sum / float64(len(values))
	std := math.Sqrt(sumSq/float64(len(values)) - mean*mean)
	if std == 0 {
		return nil
	}
	var outliers []float64
	for _, v := range values {
		if math.Abs((v-mean)/std) > threshold {
			outliers = append(outliers, v)
		}
	}
	return outliers
}
