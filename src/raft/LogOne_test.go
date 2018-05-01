package raft

import (
	"fmt"
	"testing"
)

func TestLogOneEmpty(t *testing.T) {
	log := makeEmptyLogOne()
	assertEquals(0, log.length())
	assertEquals(0, log.lastIndex())
	assertEquals(0, log.lastCompressedTerm())
	assertEquals(0, log.lastCompressedIndex())
}

func TestLogOneGetFromEmpty(t *testing.T) {
	func() {
		defer func() {
			recover()
		}()
		log := makeEmptyLogOne()
		log.get(4)
		t.Fatalf("Calling out of bounds get(4) did not panic.")
	}()
}

func TestLogOneLengthOne(t *testing.T) {
	log := makeEmptyLogOne()
	e := LogEntry{}
	log.append(e)
	assertEquals(1, log.length())
	assertEquals(1, log.lastIndex())
	assertEquals(e, log.get(1))
}

func TestLogOneGetIndicesIncludingAndAfterLengthOne(t *testing.T) {
	log := makeEmptyLogOne()
	e := LogEntry{}
	log.append(e)
	assertSliceEquals([]LogEntry{e}, log.getIndicesIncludingAndAfter(1))
}

func TestLogOneLengthTwo(t *testing.T) {
	log := makeEmptyLogOne()
	e1 := LogEntry{Index: 1}
	e2 := LogEntry{Index: 2}
	log.append(e1)
	log.append(e2)

	assertEquals(e1, log.get(1))
	assertEquals(e2, log.get(2))
	assertSliceEquals([]LogEntry{e1, e2}, log.getIndicesIncludingAndAfter(1))
	assertSliceEquals([]LogEntry{e2}, log.getIndicesIncludingAndAfter(2))
}

func TestLogOneTruncate(t *testing.T) {
	log := makeEmptyLogOne()
	e1 := LogEntry{Index: 1}
	e2 := LogEntry{Index: 2}
	log.append(e1)
	log.append(e2)

	assertEquals(2, log.length())
	log.truncateAfter(1)
	assertEquals(1, log.length())
	assertEquals(e1, log.get(1))
}

func TestLogOneTruncateOutOfBounds(t *testing.T) {
	log := makeEmptyLogOne()
	assertEquals(0, log.length())
	log.truncateAfter(5)
	assertEquals(0, log.length())

	log.append(LogEntry{})
	assertEquals(1, log.length())
	log.truncateAfter(5)
	assertEquals(1, log.length())
}

func getTestLog() (LogOne, []LogEntry) {
	toReturn := makeEmptyLogOne()
	entries := make([]LogEntry, 0)
	numEntries := 5
	for i := 1; i <= numEntries; i++ {
		e := LogEntry{Index: i, Term: i}
		entries = append(entries, e)
		toReturn.append(e)
	}
	return toReturn, entries
}

func TestLogOneCompressEntriesDoesNotChangeLaterEntries(t *testing.T) {
	log, entries := getTestLog()
	assertSliceEquals(entries, log.getIndicesIncludingAndAfter(1))

	log.compressEntriesUpTo(2)

	assertEquals(len(entries), log.length())
	// they're different because of one-indexing
	assertEquals(entries[3], log.get(4))
	assertEquals(entries[4], log.get(5))
}

func TestLogOneIndexIsCompressed(t *testing.T) {
	log, _ := getTestLog()
	log.compressEntriesUpTo(2)

	assertEquals(true, log.indexIsCompressed(1))
	assertEquals(true, log.indexIsCompressed(2))
	assertEquals(false, log.indexIsCompressed(3))
	assertEquals(false, log.indexIsCompressed(4))
	assertEquals(false, log.indexIsCompressed(5))
}

func TestLogOneLastCompressedIndex(t *testing.T) {
	log, _ := getTestLog()

	assertEquals(0, log.lastCompressedIndex())
	log.compressEntriesUpTo(3)
	assertEquals(3, log.lastCompressedIndex())
}

func TestLogOneCompressReducesSize(t *testing.T) {
	log, _ := getTestLog()

	initialSize := log.sizeBytes()
	originalString := fmt.Sprintf("%+v", log)

	log.compressEntriesUpTo(log.lastIndex())

	assertEquals(0, len(log.UncompressedEntries))
	if log.sizeBytes() >= initialSize {
		panic(fmt.Sprintf("Compressing all entries in a log changed it from %v (%d bytes) to %+v (%d bytes)!",
			originalString, initialSize, log, log.sizeBytes()))
	}
}

func TestLogOneCannotGetCompressedEntry(t *testing.T) {
	log, _ := getTestLog()

	numEntries := 5
	entryToCompressUpTo := 3
	log.compressEntriesUpTo(entryToCompressUpTo)

	for i := 1; i <= numEntries; i++ {
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

func TestLogOneCannotGetSliceCompressed(t *testing.T) {
	log, _ := getTestLog()
	log.compressEntriesUpTo(3)
	defer func() { recover() }()
	log.getIndicesIncludingAndAfter(1)
	t.Fatalf("Able to get slice that included compressed elements.")
}

func TestLogOneCompressOneIndex(t *testing.T) {
	log, _ := getTestLog()

	assertEquals(5, log.length())
	assertEquals(5, log.lastIndex())
	assertEquals(0, log.lastCompressedIndex())
	assertEquals(0, log.lastCompressedTerm())
	assertEquals(false, log.indexIsCompressed(1))

	log.compressEntriesUpTo(1) // should compress the first entry

	assertEquals(5, log.length())
	assertEquals(5, log.lastIndex())
	assertEquals(1, log.lastCompressedIndex())
	assertEquals(1, log.lastCompressedTerm())
	assertEquals(true, log.indexIsCompressed(1))
}

func TestLogOneCannotCompressNegativeIndex(t *testing.T) {
	log := makeEmptyLogOne()
	defer func() { recover() }()
	log.compressEntriesUpTo(-1)
	t.Fatalf("Able to compress indices up to and including -1.")
}

func TestLogOneCannotCallZeroIndex(t *testing.T) {
	log, _ := getTestLog()

	func() {
		defer func() { recover() }()
		log.get(0)
		t.Fatalf("Able to get 0!")
	}()
	func() {
		defer func() { recover() }()
		log.compressEntriesUpTo(0)
		t.Fatalf("Able to compressEntriesUpTo 0!")
	}()
	func() {
		defer func() { recover() }()
		log.getIndicesIncludingAndAfter(0)
		t.Fatalf("Able to getIndicesIncludingAndAfter 0!")
	}()
}

func TestLogOneTruncateAfterZero(t *testing.T) {
	log, _ := getTestLog()
	log.truncateAfter(0)
	assertEquals(0, log.length())

	defer func() { recover() }()
	log2, _ := getTestLog()
	log2.compressEntriesUpTo(1)
	log2.truncateAfter(0)
	t.Fatalf("Able to truncateAfter(0) when entries have been compressed!")
}

func TestLogOneLastCompressedTermEmpty(t *testing.T) {
	log := makeEmptyLogOne()
	assertEquals(0, log.lastCompressedTerm())

}

func TestLogOneLastCompressedTerm(t *testing.T) {
	log, _ := getTestLog()
	log.compressEntriesUpTo(1)
	assertEquals(1, log.lastCompressedTerm())

	log.compressEntriesUpTo(2)
	assertEquals(2, log.lastCompressedTerm())
}

func TestLogOneCannotCompressAlreadyCompressed(t *testing.T) {
	log, _ := getTestLog()
	log.compressEntriesUpTo(3)

	defer func() { recover() }() // error expected
	log.compressEntriesUpTo(2)
	t.Fatalf("Able to compress already-compressed entry!\n")
}

func TestLogOneCompressAllEntriesEmpty(t *testing.T) {
	log := makeEmptyLogOne()
	big := 999

	log.compressEntriesUpTo(big)
	assertEquals(big, log.length())
	assertEquals(big, log.lastCompressedIndex())
	assertEquals(0, log.lastCompressedTerm())

	e := LogEntry{}
	log.append(e)
	assertEquals(e, log.get(big+1))
}

func TestLogOneCompressAllEntriesLastTerm(t *testing.T) {
	log, _ := getTestLog()

	big := 999
	log.compressEntriesUpTo(big)
	assertEquals(5, log.lastCompressedTerm())
}

func TestLogOneCompressAllEntries(t *testing.T) {
	log, _ := getTestLog()

	big := 999
	log.compressEntriesUpTo(big)

	assertEquals(0, len(log.UncompressedEntries))
	assertEquals(big, log.lastCompressedIndex())
	assertEquals(big, log.length())
}

func TestLogOneCompressAllEntriesMultipleTimes(t *testing.T) {
	log, _ := getTestLog()

	big := 999
	log.compressEntriesUpTo(big)
	bigger := 9999999999
	log.compressEntriesUpTo(bigger)

	assertEquals(bigger, log.length())
	assertEquals(0, len(log.UncompressedEntries))
	assertEquals(bigger, log.lastCompressedIndex())
	assertEquals(5, log.lastCompressedTerm())
}

func TestLogOneCompressAllEntriesLength(t *testing.T) {
	log, _ := getTestLog()

	assertEquals(5, log.length())
	assertEquals(5, log.lastIndex())

	log.compressEntriesUpTo(5)
	assertEquals(0, len(log.UncompressedEntries))
	assertEquals(5, log.length())
	assertEquals(5, log.lastIndex())

	big := 999
	log.compressEntriesUpTo(big)
	assertEquals(big, log.length())
	assertEquals(big, log.lastIndex())
}

func TestLogOneTruncateNoncompressed(t *testing.T) {
	log, _ := getTestLog()

	log.compressEntriesUpTo(3)
	log.truncateAfter(3)
	assertEquals(3, log.length())
	assertEquals(0, len(log.UncompressedEntries))
}
