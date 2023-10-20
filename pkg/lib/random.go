package lib

import "math/rand"

// RandomSubsetOfLength: Given an array of nodes generate of random
// subset of 'num' length.
func RandomSubsetOfLength[V any](vs []V, num int) []V {
	randomSubset := make([]V, 0)
	selectedIndices := make(map[int]struct{})

	for i := 0; i < num; {
		if len(selectedIndices) == len(vs) {
			return randomSubset
		}

		randomIndex := rand.Intn(len(vs))

		if _, ok := selectedIndices[randomIndex]; !ok {
			randomSubset = append(randomSubset, vs[randomIndex])
			i++
		}
	}

	return randomSubset
}
