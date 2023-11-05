package lib

import (
	"slices"
	"testing"
)

// Test that a random subset of length 0 produces a zero length
// list
func TestRandomSubsetOfLength0(t *testing.T) {
	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	randomValues := RandomSubsetOfLength(values, 0)

	if len(randomValues) != 0 {
		t.Fatalf(`Expected length to be 0`)
	}
}

func TestRandomSubsetOfLength1(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}

	randomValues := RandomSubsetOfLength(values, 1)

	if len(randomValues) != 1 {
		t.Fatalf(`Expected length to be 1`)
	}

	if !slices.Contains(values, randomValues[0]) {
		t.Fatalf(`Expected length to be 1`)
	}
}

func TestRandomSubsetEntireList(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}
	randomValues := RandomSubsetOfLength(values, len(values))

	if len(randomValues) != len(values) {
		t.Fatalf(`Expected length to be %d was %d`, len(values), len(randomValues))
	}

	slices.Sort(randomValues)

	if !slices.Equal(values, randomValues) {
		t.Fatalf(`Expected slices to be equal`)
	}
}
