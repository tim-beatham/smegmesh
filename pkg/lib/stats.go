// lib contains helper functions for the implementation
package lib

import (
	"math"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// Modelling the distribution using a normal distribution get the count
// of the outliers
func GetOutliers[K comparable](counts map[K]uint64, alpha float64) []K {
	n := float64(len(counts))

	keys := MapKeys(counts)
	values := make([]float64, len(keys))

	for index, key := range keys {
		values[index] = float64(counts[key])
	}

	mean := stat.Mean(values, nil)
	stdDev := stat.StdDev(values, nil)

	moe := distuv.Normal{Mu: 0, Sigma: 1}.Quantile(1-alpha/2) * (stdDev / math.Sqrt(n))

	lowerBound := mean - moe

	var outliers []K

	for i, count := range values {
		if count < lowerBound {
			outliers = append(outliers, keys[i])
		}
	}

	return outliers
}
