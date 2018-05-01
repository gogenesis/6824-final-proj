package raft

import (
	"bytes"
	"fmt"
	"labgob"
)

// An abstraction of a []LogEntry that can forget about past Entries.
// This log is one-indexed and is not threadsafe.
// Providing zero as an index will cause any method to panic.
type LogOne struct {
	FirstUncompressedIndex int // The index of the first uncompressed entry stored in Entries
	LastCompressedTerm     int // The term of the last compressed entry
	UncompressedEntries    []LogEntry
}

// Constructor ================================================================

func makeEmptyLogOne() LogOne {
	var log LogOne
	log.FirstUncompressedIndex = 1 // because we're 1-indexing
	log.LastCompressedTerm = 0
	// Entries does not need to be initialized because it defaults to empty.
	return log
}

// Mutator methods ============================================================

// Append an entry to the log.
func (log *LogOne) append(entry LogEntry) {
	log.UncompressedEntries = append(log.UncompressedEntries, entry)
}

// Append several entries to the log.
func (log *LogOne) appendAll(entries []LogEntry) {
	for _, entry := range entries {
		log.append(entry)
	}
}

// Compress all Entries up to and including a specified index.
// Compressed Entries cannot be accessed, but still count for the purposes of indexing with get().
// If index > lastIndex, compresses all entries.
// Panics if index is already compressed.
func (log *LogOne) compressEntriesUpTo(index int) {
	log.assertNotCompressed(index)
	if index < log.lastIndex() {
		// guaranteed positive because of the assertion, guaranteed not out of bounds by the check in the previous line
		numEntriesToCompress := index - log.FirstUncompressedIndex + 1
		if numEntriesToCompress <= 0 {
			fmt.Printf("log=%+v\n", log)
			fmt.Printf("numEntriesToCompress=%v\n", numEntriesToCompress)
			fmt.Printf("index=%v\n", index)
			panic("AssertionError")
		}
		log.LastCompressedTerm = log.UncompressedEntries[numEntriesToCompress-1].Term
		log.UncompressedEntries = log.UncompressedEntries[numEntriesToCompress:]
	} else {
		if len(log.UncompressedEntries) > 0 {
			// won't fail because there are some uncompressed entries, so the last one must be uncompressed
			log.LastCompressedTerm = log.get(log.lastIndex()).Term
			log.UncompressedEntries = make([]LogEntry, 0)
		}
	}
	log.FirstUncompressedIndex = index + 1
}

// Delete all Entries after, but not including, lastIndex.
// Panics if lastIndex + 1 has already been compressed.
// Does nothing if index > log.lastIndex().
// Call truncateAfter(0) to clear the whole log (assuming nothing has been compressed)
func (log *LogOne) truncateAfter(lastIndexToKeep int) {
	if lastIndexToKeep >= log.length() {
		return
	}
	if lastIndexToKeep == log.lastCompressedIndex() {
		// Simply throw out the uncompressed entries
		log.UncompressedEntries = make([]LogEntry, 0)
		return
	}
	log.assertNotCompressed(lastIndexToKeep)
	log.UncompressedEntries = log.UncompressedEntries[:lastIndexToKeep-log.FirstUncompressedIndex+1]
}

// Observer methods ===========================================================

// Get the entry at a specified index.
// Panics if that entry has already been compressed or if index <= 0.
func (log *LogOne) get(index int) LogEntry {
	log.assertNotCompressed(index)
	indexIntoEntries := index - log.FirstUncompressedIndex
	if indexIntoEntries >= len(log.UncompressedEntries) {
		panic(fmt.Sprintf("Index %d is out of bounds, lastIndex=%d", index, log.lastIndex()))
	}
	return log.UncompressedEntries[indexIntoEntries]
}

// Return a slice starting at index index and all entries after that.
// Panics if index > log.length() or index has been compressed.
func (log *LogOne) getIndicesIncludingAndAfter(index int) []LogEntry {
	log.assertNotCompressed(index)
	var toReturn []LogEntry
	for i := index; i <= log.lastIndex(); i++ {
		toReturn = append(toReturn, log.get(i))
	}
	return toReturn
}

// Get the length of the log, including compressed Entries.
func (log *LogOne) length() int {
	return log.FirstUncompressedIndex + len(log.UncompressedEntries) - 1 // for the starter 1
}

// Get the last index of the log.
// Returns 0 if the log is empty.
func (log *LogOne) lastIndex() int {
	return log.length()
}

// Returns true iff an index has been compressed.
// If false, that index can be gotten with get().
func (log *LogOne) indexIsCompressed(index int) bool {
	return index < log.FirstUncompressedIndex
}

// Returns true iff an index is not compressed.
// If true, that index can be gotten with get().
func (log *LogOne) indexIsUncompressed(index int) bool {
	return !log.indexIsCompressed(index)
}

// Return the size of the uncompressed Entries, in bytes.
func (log *LogOne) sizeBytes() int {
	byteBuffer := new(bytes.Buffer)
	encoder := labgob.NewEncoder(byteBuffer)
	encoder.Encode(log.FirstUncompressedIndex)
	encoder.Encode(log.LastCompressedTerm)
	encoder.Encode(log.UncompressedEntries)
	return len(byteBuffer.Bytes())
}

func (log *LogOne) assertNotCompressed(index int) {
	if index < 1 {
		panic(fmt.Sprintf("Illegal index %d!", index))
	}
	if log.indexIsCompressed(index) {
		panic(fmt.Sprintf("Index %d has been compressed already!", index))
	}
}

// Returns the index of the last compressed log entry,
// or 0 if no entries have been compressed.
func (log *LogOne) lastCompressedIndex() int {
	return log.FirstUncompressedIndex - 1
}

// Returns the term of the last compressed log entry,
// or 0 if no entries have been compressed.
func (log *LogOne) lastCompressedTerm() int {
	return log.LastCompressedTerm
}

func (log *LogOne) lastTerm() int {
	if log.indexIsCompressed(log.lastIndex()) {
		return log.lastCompressedTerm()
	}
	return log.get(log.lastIndex()).Term
}
