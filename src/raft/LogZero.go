package raft

import (
	"bytes"
	"fmt"
	"labgob"
)

// An abstraction of a []LogEntry that can forget about past Entries.
// This log is zero-indexed and is not threadsafe.
type LogZero struct {
	FirstUncompressedIndex int // The index of the first uncompressed entry stored in Entries
	LastCompressedTerm     int // The term of the last compressed entry
	UncompressedEntries    []LogEntry
}

// Constructor ================================================================

func makeEmptyLogZero() LogZero {
	var log LogZero
	log.FirstUncompressedIndex = 0
	log.LastCompressedTerm = -1
	// Entries does not need to be initialized because it defaults to empty.
	return log
}

// Mutator methods ============================================================

// Append an entry to the log.
func (log *LogZero) append(entry LogEntry) {
	log.UncompressedEntries = append(log.UncompressedEntries, entry)
}

// Append several entries to the log.
func (log *LogZero) appendAll(entries []LogEntry) {
	for _, entry := range entries {
		log.append(entry)
	}
}

// Compress all Entries up to and including a specified index.
// Compressed Entries cannot be accessed, but still count for the purposes of indexing with get().
// If index > lastIndex, compresses all entries.
// Panics if index is already compressed.
func (log *LogZero) compressEntriesUpTo(index int) {
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
func (log *LogZero) truncate(lastIndex int) {
	if lastIndex >= log.length() {
		return
	}
	if lastIndex == log.lastCompressedIndex() {
		// Simply throw out the uncompressed entries
		log.UncompressedEntries = make([]LogEntry, 0)
		return
	}
	log.assertNotCompressed(lastIndex)
	log.UncompressedEntries = log.UncompressedEntries[:lastIndex-log.FirstUncompressedIndex+1]
}

// Observer methods ===========================================================

// Get the entry at a specified index.
// Panics if that entry has already been compressed.
func (log *LogZero) get(index int) LogEntry {
	log.assertNotCompressed(index)
	indexIntoEntries := index - log.FirstUncompressedIndex
	if indexIntoEntries >= len(log.UncompressedEntries) {
		panic(fmt.Sprintf("Index %d is out of bounds, lastIndex=%d", index, log.lastIndex()))
	}
	return log.UncompressedEntries[indexIntoEntries]
}

// Get a slice of the log.
// Panics if any Entries that would be included in the slice have already been compressed.
func (log *LogZero) getSlice(start, endNotInclusive int) []LogEntry {
	log.assertNotCompressed(start)
	startIndexIntoEntries := start - log.FirstUncompressedIndex
	return log.UncompressedEntries[startIndexIntoEntries : endNotInclusive-log.FirstUncompressedIndex]
}

// Get the length of the log, including compressed Entries.
func (log *LogZero) length() int {
	return log.FirstUncompressedIndex + len(log.UncompressedEntries)
}

// Get the last index of the log.
// Returns -1 if the log is empty.
func (log *LogZero) lastIndex() int {
	return log.length() - 1
}

// Returns true iff an index has been compressed.
// If false, that index can be gotten with get().
func (log *LogZero) indexIsCompressed(index int) bool {
	return index < log.FirstUncompressedIndex
}

// Returns true iff an index is not compressed.
// If true, that index can be gotten with get().
func (log *LogZero) indexIsUncompressed(index int) bool {
	return !log.indexIsCompressed(index)
}

// Return the size of the uncompressed Entries, in bytes.
func (log *LogZero) sizeBytes() int {
	byteBuffer := new(bytes.Buffer)
	encoder := labgob.NewEncoder(byteBuffer)
	encoder.Encode(log.FirstUncompressedIndex)
	encoder.Encode(log.UncompressedEntries)
	return len(byteBuffer.Bytes())
}

func (log *LogZero) assertNotCompressed(index int) {
	if log.indexIsCompressed(index) {
		panic(fmt.Sprintf("Index %d has been compressed already!", index))
	}
}

// Returns the index of the last compressed log entry,
// or -1 if no entries have been compressed.
func (log *LogZero) lastCompressedIndex() int {
	return log.FirstUncompressedIndex - 1
}

// Returns the term of the last compressed log entry,
// or -1 if no entries have been compressed.
func (log *LogZero) lastCompressedTerm() int {
	return log.LastCompressedTerm
}

func (log *LogZero) lastTerm() int {
	if log.indexIsCompressed(log.lastIndex()) {
		return log.lastCompressedTerm()
	}
	return log.get(log.lastIndex()).Term
}
