package lib

import (
	"slices"
	"testing"
)

func stringToInt(input string) int {
	return len(input)
}

func intDiv(input int) int {
	return input / 2
}

func TestMapValuesMapsValues(t *testing.T) {
	values := []int{1, 4, 11, 92}
	var theMap map[string]int = map[string]int{
		"mynameisjeff": values[0],
		"tim":          values[1],
		"bob":          values[2],
		"derek":        values[3],
	}

	mapValues := MapValues(theMap)

	for _, elem := range mapValues {
		if !slices.Contains(values, elem) {
			t.Fatalf(`%d is not an expected value`, elem)
		}
	}

	if len(mapValues) != len(theMap) {
		t.Fatalf(`Expected length %d got %d`, len(theMap), len(mapValues))
	}
}

func TestMapValuesWithExcludeExcludesValues(t *testing.T) {
	values := []int{1, 9, 22}
	var theMap map[string]int = map[string]int{
		"mynameisbob": values[0],
		"tim":         values[1],
		"bob":         values[2],
	}

	exclude := map[string]struct{}{
		"tim": {},
	}

	mapValues := MapValuesWithExclude(theMap, exclude)

	if slices.Contains(mapValues, values[1]) {
		t.Fatalf(`Failed to exclude expected value`)
	}

	if len(mapValues) != 2 {
		t.Fatalf(`Incorrect expected length`)
	}

	for _, value := range theMap {
		if !slices.Contains(values, value) {
			t.Fatalf(`Element does not exist in the list of
				expected values`)
		}
	}
}

func TestMapKeys(t *testing.T) {
	keys := []string{"1", "2", "3"}

	theMap := map[string]int{
		keys[0]: 1,
		keys[1]: 2,
		keys[2]: 3,
	}

	mapKeys := MapKeys(theMap)

	for _, elem := range mapKeys {
		if !slices.Contains(keys, elem) {
			t.Fatalf(`%s elem is not an expected key`, elem)
		}
	}

	if len(mapKeys) != len(theMap) {
		t.Fatalf(`Missing expected values`)
	}
}

func TestMapValues(t *testing.T) {
	array := []string{"mynameisjeff", "tim", "bob", "derek"}

	intArray := Map(array, stringToInt)

	for index, elem := range intArray {
		if len(array[index]) != elem {
			t.Fatalf(`Have %d want %d`, elem, len(array[index]))
		}
	}
}

func TestFilterFilterAll(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}

	filterFunc := func(n int) bool {
		return false
	}

	newValues := Filter(values, filterFunc)

	if len(newValues) != 0 {
		t.Fatalf(`Expected value was 0`)
	}
}

func TestFilterFilterNone(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}

	filterFunc := func(n int) bool {
		return true
	}

	newValues := Filter(values, filterFunc)

	if !slices.Equal(values, newValues) {
		t.Fatalf(`Expected lists to be the same`)
	}
}

func TestFilterFilterSome(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}

	filterFunc := func(n int) bool {
		return n < 3
	}

	expected := []int{1, 2}

	actual := Filter(values, filterFunc)

	if !slices.Equal(expected, actual) {
		t.Fatalf(`Expected expected and actual to be the same`)
	}
}
