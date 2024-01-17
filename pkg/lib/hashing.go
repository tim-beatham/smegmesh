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
// greater than ours.
func ConsistentHash[V any, K any](values []V, client K, bucketFunc func(V) int, keyFunc func(K) int) V {
	if len(values) == 0 {
		panic("values is empty")
	}

	vs := Map(values, func(v V) consistentHashRecord[V] {
		return consistentHashRecord[V]{
			v,
			bucketFunc(v),
		}
	})

	sort.SliceStable(vs, func(i, j int) bool {
		return vs[i].value < vs[j].value
	})

	ourKey := keyFunc(client)

	idx := sort.Search(len(vs), func(i int) bool {
		return vs[i].value >= ourKey
	})

	if idx == len(vs) {
		return vs[0].record
	}

	return vs[idx].record
}
