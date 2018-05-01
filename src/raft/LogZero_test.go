package raft

import (
	"fmt"
	"reflect"
	"testing"
)

func assertSliceEquals(expected, actual []LogEntry) {
	if !reflect.DeepEqual(expected, actual) {
		panic(fmt.Sprintf("Expected %+v, got %+v\n", expected, actual))
	}
}

func TestLogZeroEmpty(t *testing.T) {
	log := makeEmptyLogZero()
	assertEquals(0, log.length())
	assertEquals(-1, log.lastIndex())
}

func TestLogZeroGetFromEmpty(t *testing.T) {
	func() {
		defer func() {
			recover()
		}()
		log := makeEmptyLogZero()
		log.get(0)
		t.Fatalf("Calling out of bounds get(1) did not panic.")
	}()
}

func TestLogZeroLengthOne(t *testing.T) {
	log := makeEmptyLogZero()
	e := LogEntry{}
	log.append(e)
	assertEquals(1, log.length())
	assertEquals(0, log.lastIndex())
	assertEquals(e, log.get(0))
}

func TestLogZeroGetSliceLengthOne(t *testing.T) {
	log := makeEmptyLogZero()
	e := LogEntry{}
	log.append(e)
	assertSliceEquals([]LogEntry{e}, log.getSlice(0, 1))
}

func TestLogZeroLengthTwo(t *testing.T) {
	log := makeEmptyLogZero()
	e1 := LogEntry{Index: 1}
	e2 := LogEntry{Index: 2}
	log.append(e1)
	log.append(e2)

	assertEquals(e1, log.get(0))
	assertEquals(e2, log.get(1))
	assertSliceEquals([]LogEntry{e1, e2}, log.getSlice(0, 2))
}

func TestLogZeroTruncate(t *testing.T) {
	log := makeEmptyLogZero()
	e1 := LogEntry{Index: 1}
	e2 := LogEntry{Index: 2}
	log.append(e1)
	log.append(e2)

	assertEquals(2, log.length())
	log.truncate(0)
	assertEquals(1, log.length())
	assertEquals(e1, log.get(0))
}

func TestLogZeroTruncateOutOfBounds(t *testing.T) {
	log := makeEmptyLogZero()
	assertEquals(0, log.length())
	log.truncate(2)
	assertEquals(0, log.length())

	log.append(LogEntry{})
	assertEquals(1, log.length())
	log.truncate(2)
	assertEquals(1, log.length())
}

func TestLogZeroCompressEntriesDoesNotChangeLaterEntries(t *testing.T) {
	log := makeEmptyLogZero()
	entries := make([]LogEntry, 0)
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i}
		entries = append(entries, e)
		log.append(e)
	}
	assertSliceEquals(entries, log.getSlice(0, numEntries))

	log.compressEntriesUpTo(2)

	assertEquals(len(entries), log.length())
	assertEquals(entries[3], log.get(3))
	assertEquals(entries[4], log.get(4))
}

func TestLogZeroIndexIsCompressed(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i}
		log.append(e)
	}

	log.compressEntriesUpTo(2)

	assertEquals(true, log.indexIsCompressed(0))
	assertEquals(true, log.indexIsCompressed(1))
	assertEquals(true, log.indexIsCompressed(2))
	assertEquals(false, log.indexIsCompressed(3))
	assertEquals(false, log.indexIsCompressed(4))
}

func TestLogZeroLastCompressedIndex(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i}
		log.append(e)
	}

	assertEquals(-1, log.lastCompressedIndex())
	log.compressEntriesUpTo(3)
	assertEquals(3, log.lastCompressedIndex())
}

func TestLogZeroCompressReducesSize(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i}
		log.append(e)
	}

	initialSize := log.sizeBytes()
	originalString := fmt.Sprintf("%+v", log)

	log.compressEntriesUpTo(log.lastIndex())

	assertEquals(0, len(log.UncompressedEntries))
	if log.sizeBytes() >= initialSize {
		fmt.Printf("Compressing all entries in a log changed it from %v (%d bytes) to %+v (%d bytes)!", originalString,
			initialSize, log, log.sizeBytes())
		panic("")
	}
}

func TestLogZeroCannotGetCompressedEntry(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i}
		log.append(e)
	}
	entryToCompressUpTo := 3
	log.compressEntriesUpTo(entryToCompressUpTo)

	for i := 0; i < numEntries; i++ {
		// need a separate func so that panic/recover doesn't lose track of our count
		func(i int) {
			if i <= entryToCompressUpTo {
				// expect a failure
				defer func() { recover() }()
				assertEquals(true, log.indexIsCompressed(i))
				log.get(i)
				t.Fatalf("get() on compressed entry with index %d did not panic.\n", i)
			} else {
				assertEquals(false, log.indexIsCompressed(i))
				e := log.get(i)
				assertEquals(i, e.Index) // make sure they're not out of order
			}
		}(i)
	}
}

func TestLogZeroCannotGetSliceCompressed(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i}
		log.append(e)
	}
	log.compressEntriesUpTo(2)
	defer func() { recover() }()
	log.getSlice(1, 4)
	t.Fatalf("Able to get slice that included compressed elements.")
}

func TestCompressZeroIndex(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}

	assertEquals(5, log.length())
	assertEquals(4, log.lastIndex())
	assertEquals(-1, log.lastCompressedIndex())
	assertEquals(-1, log.lastCompressedTerm())
	assertEquals(false, log.indexIsCompressed(0))

	log.compressEntriesUpTo(0) // should compress the first entry

	assertEquals(5, log.length())
	assertEquals(4, log.lastIndex())
	assertEquals(0, log.lastCompressedIndex())
	assertEquals(0, log.lastCompressedTerm())
	assertEquals(true, log.indexIsCompressed(0))
}

func TestCannotCompressNegativeIndex(t *testing.T) {
	log := makeEmptyLogZero()
	defer func() { recover() }()
	log.compressEntriesUpTo(-1)
	t.Fatalf("Able to compress indices up to and including -1.")
}

func TestLogZeroLastCompressedTermEmpty(t *testing.T) {
	log := makeEmptyLogZero()
	assertEquals(-1, log.lastCompressedTerm())

}

func TestLogZeroLastCompressedTerm(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}
	log.compressEntriesUpTo(1)
	assertEquals(1, log.lastCompressedTerm())

	log.compressEntriesUpTo(2)
	assertEquals(2, log.lastCompressedTerm())
}

func TestLogZeroCannotCompressAlreadyCompressed(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}

	log.compressEntriesUpTo(3)

	defer func() { recover() }() // error expected
	log.compressEntriesUpTo(2)
	t.Fatalf("Able to compress already-compressed entry!\n")
}

func TestLogZeroCompressAllEntriesEmpty(t *testing.T) {
	log := makeEmptyLogZero()
	big := 4

	log.compressEntriesUpTo(big)
	assertEquals(big+1, log.length())
	assertEquals(big, log.lastCompressedIndex())
	assertEquals(-1, log.lastCompressedTerm())

	e := LogEntry{}
	log.append(e)
	assertEquals(e, log.get(big+1))
}

func TestLogZeroCompressAllEntriesLastTerm(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}

	big := 999
	log.compressEntriesUpTo(big)
	assertEquals(4, log.lastCompressedTerm())
}

func TestLogZeroCompressAllEntries(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}

	big := 999
	log.compressEntriesUpTo(big)

	assertEquals(0, len(log.UncompressedEntries))
	assertEquals(big, log.lastCompressedIndex())
	assertEquals(big+1, log.length())
}

func TestLogZeroCompressAllEntriesMultipleTimes(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}

	big := 999
	log.compressEntriesUpTo(big)
	bigger := 9999999999
	log.compressEntriesUpTo(bigger)

	assertEquals(bigger+1, log.length())
	assertEquals(0, len(log.UncompressedEntries))
	assertEquals(bigger, log.lastCompressedIndex())
	assertEquals(4, log.lastCompressedTerm())
}

func TestLogZeroCompressAllEntriesLength(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}

	assertEquals(5, log.length())
	assertEquals(4, log.lastIndex())

	log.compressEntriesUpTo(4)
	assertEquals(0, len(log.UncompressedEntries))
	assertEquals(5, log.length())
	assertEquals(4, log.lastIndex())

	big := 999
	log.compressEntriesUpTo(big)
	assertEquals(big, log.lastIndex())
}

func TestLogZeroTruncateNoncompressed(t *testing.T) {
	log := makeEmptyLogZero()
	numEntries := 5
	for i := 0; i < numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		log.append(e)
	}

	log.compressEntriesUpTo(2)
	log.truncate(2)
	assertEquals(3, log.length())
	assertEquals(0, len(log.UncompressedEntries))
}
