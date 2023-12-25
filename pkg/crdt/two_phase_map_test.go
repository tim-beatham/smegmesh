package crdt

import (
	"hash/fnv"
	"slices"
	"testing"
)

func NewMap(processId string) *TwoPhaseMap[string, string] {
	theMap := NewTwoPhaseMap[string, string](processId, func(key string) uint64 {
		hash := fnv.New64a()
		hash.Write([]byte(key))
		return hash.Sum64()
	}, 1)
	return theMap
}

func TestTwoPhaseMapEmpty(t *testing.T) {
	theMap := NewMap("a")

	if theMap.Contains("a") {
		t.Fatalf(`a should not be present in the map`)
	}
}

func TestTwoPhaseMapValuePresent(t *testing.T) {
	theMap := NewMap("a")
	theMap.Put("a", "")

	if !theMap.Contains("a") {
		t.Fatalf(`should be present within the map`)
	}
}

func TestTwoPhaseMapValueNotPresent(t *testing.T) {
	theMap := NewMap("a")
	theMap.Put("b", "")

	if theMap.Contains("a") {
		t.Fatalf(`a should not be present in the map`)
	}
}

func TestTwoPhaseMapPutThenRemove(t *testing.T) {
	theMap := NewMap("a")

	theMap.Put("a", "")
	theMap.Remove("a")

	if theMap.Contains("a") {
		t.Fatalf(`a should not be present within the map`)
	}
}

func TestTwoPhaseMapPutThenRemoveThenPut(t *testing.T) {
	theMap := NewMap("a")

	theMap.Put("a", "")
	theMap.Remove("a")
	theMap.Put("a", "")

	if !theMap.Contains("a") {
		t.Fatalf(`a should be present within the map`)
	}
}

func TestMarkMarksTheValueIn2PMap(t *testing.T) {
	theMap := NewMap("a")

	theMap.Put("a", "")
	theMap.Mark("a")

	if !theMap.IsMarked("a") {
		t.Fatalf(`a should be marked`)
	}
}

func TestAsListReturnsItemsInList(t *testing.T) {
	theMap := NewMap("a")

	theMap.Put("a", "bob")
	theMap.Put("b", "dylan")

	keys := theMap.AsList()
	slices.Sort(keys)

	if !slices.Equal([]string{"bob", "dylan"}, keys) {
		t.Fatalf(`values should be bob, dylan`)
	}
}

func TestSnapShotRemoveMapEmpty(t *testing.T) {
	theMap := NewMap("a")
	theMap.Put("a", "bob")
	theMap.Put("b", "dylan")

	snapshot := theMap.Snapshot()

	if len(snapshot.Add) != 2 {
		t.Fatalf(`add values length should be 2`)
	}

	if len(snapshot.Remove) != 0 {
		t.Fatalf(`remove map length should be 0`)
	}
}

func TestSnapshotMapEmpty(t *testing.T) {
	theMap := NewMap("a")

	snapshot := theMap.Snapshot()

	if len(snapshot.Add) != 0 || len(snapshot.Remove) != 0 {
		t.Fatalf(`snapshot length should be 0`)
	}
}

func TestSnapShotFromStateReturnsIntersection(t *testing.T) {
	map1 := NewMap("a")
	map1.Put("a", "heyy")

	map2 := NewMap("b")
	map2.Put("b", "hmmm")

	message := map2.GenerateMessage()

	snapShot := map1.SnapShotFromState(message)

	if len(snapShot.Add) != 1 {
		t.Fatalf(`add length should be 1`)
	}

	if len(snapShot.Remove) != 0 {
		t.Fatalf(`remove length should be 0`)
	}
}

func TestGetHashDifferentOnChange(t *testing.T) {
	theMap := NewMap("a")

	prevHash := theMap.GetHash()

	theMap.Put("b", "hmmhmhmh")

	if prevHash == theMap.GetHash() {
		t.Fatalf(`hashes should not be the same`)
	}
}

func TestGenerateMessageReturnsClocks(t *testing.T) {
	theMap := NewMap("a")
	theMap.Put("a", "hmm")
	theMap.Put("b", "hmm")
	theMap.Remove("a")

	message := theMap.GenerateMessage()

	if len(message.AddContents) != 2 {
		t.Fatalf(`two items added add should be 2`)
	}

	if len(message.RemoveContents) != 1 {
		t.Fatalf(`a was removed remove map should be length 1`)
	}
}

func TestDifferenceReturnsDifferenceOfMaps(t *testing.T) {
	map1 := NewMap("a")
	map1.Put("a", "ssms")
	map1.Put("b", "sdmdsmd")

	map2 := NewMap("b")
	map2.Put("d", "eek")
	map2.Put("c", "meh")

	message1 := map1.GenerateMessage()
	message2 := map2.GenerateMessage()

	difference := message1.Difference(0, message2)

	if len(difference.AddContents) != 2 {
		t.Fatalf(`d and c are not in map1 they should be in add contents`)
	}

	if len(difference.RemoveContents) != 0 {
		t.Fatalf(`remove should be empty`)
	}
}

func TestMergeMergesValuesThatAreGreaterThanCurrentClock(t *testing.T) {
	map1 := NewMap("a")
	map1.Put("a", "ssms")
	map1.Put("b", "sdmdsmd")

	map2 := NewMap("b")
	map2.Put("d", "eek")
	map2.Put("c", "meh")

	message1 := map1.GenerateMessage()
	message2 := map2.GenerateMessage()

	difference := message1.Difference(0, message2)
	state := map2.SnapShotFromState(difference)

	map1.Merge(*state)

	if !map1.Contains("d") {
		t.Fatalf(`d should be in the map`)
	}

	if !map2.Contains("c") {
		t.Fatalf(`c should be in the map`)
	}
}
