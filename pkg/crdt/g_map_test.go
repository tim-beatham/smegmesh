// crdt_test unit tests the crdt implementations
package crdt

import (
	"hash/fnv"
	"slices"
	"testing"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/lib"
)

func NewGmap() *GMap[string, bool] {
	vectorClock := NewVectorClock("a", func(key string) uint64 {
		hash := fnv.New64a()
		hash.Write([]byte(key))
		return hash.Sum64()
	}, 1) // 1 second stale time

	gMap := NewGMap[string, bool](vectorClock)
	return gMap
}

func TestGMapPutInsertsItems(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("bruh1234", true)

	if !gMap.Contains("bruh1234") {
		t.Fatalf(`value not added to map`)
	}
}

func TestGMapPutReplacesItems(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("bruh1234", true)
	gMap.Put("bruh1234", false)

	value := gMap.Get("bruh1234")

	if value {
		t.Fatalf(`value should ahve been replaced to false`)
	}
}

func TestContainsValueNotPresent(t *testing.T) {
	gMap := NewGmap()

	if gMap.Contains("sdhjsdhsdj") {
		t.Fatalf(`value should not be present in the map`)
	}
}

func TestContainsValuePresent(t *testing.T) {
	gMap := NewGmap()
	key := "hehehehe"
	gMap.Put(key, false)

	if !gMap.Contains(key) {
		t.Fatalf(`%s should not be present in the map`, key)
	}
}

func TestGMapGetNotPresentReturnsError(t *testing.T) {
	gMap := NewGmap()
	value := gMap.Get("bruh123")

	if value != false {
		t.Fatalf(`value should be default type false`)
	}
}

func TestGMapGetReturnsValue(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("bobdylan", true)

	value := gMap.Get("bobdylan")

	if !value {
		t.Fatalf("value should be true but was false")
	}
}

func TestMarkMarksTheValue(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("hello123", true)

	gMap.Mark("hello123")

	if !gMap.IsMarked("hello123") {
		t.Fatal(`hello123 should be marked`)
	}
}

func TestMarkValueNotPresent(t *testing.T) {
	gMap := NewGmap()
	gMap.Mark("ok123456")
}

func TestKeysMapEmpty(t *testing.T) {
	gMap := NewGmap()

	keys := gMap.Keys()

	if len(keys) != 0 {
		t.Fatal(`list of keys was not empty but should be empty`)
	}
}

func TestKeysMapReturnsKeysInMap(t *testing.T) {
	gMap := NewGmap()

	gMap.Put("a", false)
	gMap.Put("b", false)
	gMap.Put("c", false)

	keys := gMap.Keys()

	if len(keys) != 3 {
		t.Fatal(`key length should be 3`)
	}
}

func TestSaveMapEmptyReturnsEmptyMap(t *testing.T) {
	gMap := NewGmap()

	saveMap := gMap.Save()

	if len(saveMap) != 0 {
		t.Fatal(`saves should be empty`)
	}
}

func TestSaveMapReturnsMapOfBuckets(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("a", false)
	gMap.Put("b", false)
	gMap.Put("c", false)

	saveMap := gMap.Save()

	if len(saveMap) != 3 {
		t.Fatalf(`save length should be 3`)
	}
}

func TestSaveWithKeysNoKeysReturnsEmptyBucket(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("a", false)
	gMap.Put("b", false)
	gMap.Put("c", false)

	saveMap := gMap.SaveWithKeys([]uint64{})

	if len(saveMap) != 0 {
		t.Fatalf(`save map should be empty`)
	}
}

func TestSaveWithKeysReturnsIntersection(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("a", false)
	gMap.Put("b", false)
	gMap.Put("c", false)

	clock := lib.MapKeys(gMap.GetClock())
	clock = clock[:len(clock)-1]

	values := gMap.SaveWithKeys(clock)
	if len(values) != len(clock) {
		t.Fatalf(`intersection not returned`)
	}
}

func TestGetClockMapEmptyReturnsEmptyClock(t *testing.T) {
	gMap := NewGmap()

	clocks := gMap.GetClock()

	if len(clocks) != 0 {
		t.Fatalf(`vector clock is not empty`)
	}
}

func TestGetClockReturnsAllCLocks(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("a", false)
	gMap.Put("b", false)
	gMap.Put("c", false)

	clocks := lib.MapValues(gMap.GetClock())
	slices.Sort(clocks)

	if !slices.Equal([]uint64{0, 1, 2}, clocks) {
		t.Fatalf(`clocks are invalid`)
	}
}

func TestGetHashChangesHashOnValueAdded(t *testing.T) {
	gMap := NewGmap()
	gMap.Put("a", false)
	prevHash := gMap.GetHash()

	gMap.Put("b", true)

	if prevHash == gMap.GetHash() {
		t.Fatalf(`hash should be different`)
	}
}

func TestPruneGarbageCollectsValuesThatHaveNotBeenUpdated(t *testing.T) {
	gMap := NewGmap()
	gMap.clock.Put("c", 12)
	gMap.Put("c", false)
	gMap.Put("a", false)

	time.Sleep(4 * time.Second)
	gMap.Put("a", true)

	gMap.Prune()

	if gMap.Contains("c") {
		t.Fatalf(`a should have been pruned`)
	}
}
