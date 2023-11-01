package lib

// MapToSlice converts a map to a slice in go
func MapValues[K comparable, V any](m map[K]V) []V {
	return MapValuesWithExclude(m, map[K]struct{}{})
}

func MapValuesWithExclude[K comparable, V any](m map[K]V, exclude map[K]struct{}) []V {
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

func MapKeys[K comparable, V any](m map[K]V) []K {
	values := make([]K, len(m))

	i := 0
	for k, _ := range m {
		values[i] = k
		i++
	}

	return values
}

type convert[V1 any, V2 any] func(V1) V2

func Map[V1 any, V2 any](list []V1, f convert[V1, V2]) []V2 {
	newList := make([]V2, len(list))

	for i, elem := range list {
		newList[i] = f(elem)
	}

	return newList
}
