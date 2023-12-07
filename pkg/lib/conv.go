package lib

import "cmp"

// MapToSlice converts a map to a slice in go
func MapValues[K cmp.Ordered, V any](m map[K]V) []V {
	return MapValuesWithExclude(m, map[K]struct{}{})
}

func MapValuesWithExclude[K cmp.Ordered, V any](m map[K]V, exclude map[K]struct{}) []V {
	values := make([]V, len(m)-len(exclude))

	i := 0

	if len(m)-len(exclude) <= 0 {
		return values
	}

	for k, v := range m {
		if _, excluded := exclude[k]; excluded {
			continue
		}

		values[i] = v
		i++
	}

	return values
}

func MapKeys[K cmp.Ordered, V any](m map[K]V) []K {
	values := make([]K, len(m))

	i := 0
	for k := range m {
		values[i] = k
		i++
	}

	return values
}

type convert[V1 any, V2 any] func(V1) V2

// Map turns a list of type V1 into type V2
func Map[V1 any, V2 any](list []V1, f convert[V1, V2]) []V2 {
	newList := make([]V2, len(list))

	for i, elem := range list {
		newList[i] = f(elem)
	}

	return newList
}

type filterFunc[V any] func(V) bool

// Filter filters out elements given a filter function.
// If filter function is true keep it in otherwise leave it out
func Filter[V any](list []V, f filterFunc[V]) []V {
	newList := make([]V, 0)

	for _, elem := range list {
		if f(elem) {
			newList = append(newList, elem)
		}
	}

	return newList
}

func Contains[V any](list []V, proposition func(V) bool) bool {
	for _, elem := range list {
		if proposition(elem) {
			return true
		}
	}

	return false
}

func Reduce[A any, V any](start A, values []V, reduce func(A, V) A) A {
	accum := start

	for _, elem := range values {
		accum = reduce(accum, elem)
	}

	return accum
}
