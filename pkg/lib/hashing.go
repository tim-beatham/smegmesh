package lib

import (
	"hash/fnv"
	"sort"
)

type consistentHashRecord[V any] struct {
	record V
	value  int
}

func HashString(value string) int {
	f := fnv.New32a()
	f.Write([]byte(value))
	return int(f.Sum32())
}

// ConsistentHash implementation. Traverse the values until we find a key
// less than ours.
func ConsistentHash[V any](values []V, client V, keyFunc func(V) int) V {
	if len(values) == 0 {
		panic("values is empty")
	}

	vs := Map(values, func(v V) consistentHashRecord[V] {
		return consistentHashRecord[V]{
			v,
			keyFunc(v),
		}
	})

	sort.SliceStable(vs, func(i, j int) bool {
		return vs[i].value < vs[j].value
	})

	ourKey := keyFunc(client)

	for _, record := range vs {
		if ourKey < record.value {
			return record.record
		}
	}

	return vs[0].record
}
