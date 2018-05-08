package raft

import (
	"bytes"
	"labgob"
)

// Get the state that is persisted.
// ONLY CALL WITH THE LOCK
func (rf *Raft) getPersistState() []byte {
	byteBuffer := new(bytes.Buffer)
	encoder := labgob.NewEncoder(byteBuffer)
	encoder.Encode(rf.CurrentTerm)
	encoder.Encode(rf.VotedFor)
	encoder.Encode(rf.Log)
	return byteBuffer.Bytes()
}

// save Raft's persistent state to stable storage, where it can later be retrieved after a crash and restart.
// ONLY call when you have the lock!
func (rf *Raft) writePersist() {
	persistentState := rf.getPersistState()
	rf.persister.SaveRaftState(persistentState)
}

// restore previously persisted state.
// ONLY call with the lock!
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 {
		return
	}

	byteBuffer := bytes.NewBuffer(data)
	decoder := labgob.NewDecoder(byteBuffer)

	var currentTerm int
	if decoder.Decode(&currentTerm) != nil {
		panic("Error decoding currentTerm!")
	} else {
		rf.CurrentTerm = currentTerm
	}

	var votedFor int
	if decoder.Decode(&votedFor) != nil {
		panic("Error decoding votedFor!")
	} else {
		rf.VotedFor = votedFor
	}

	//log := LogOne{}
	//var LastCompressedTerm int
	//decoder.Decode(&LastCompressedTerm)
	//log.LastCompressedTerm = LastCompressedTerm
	//var FirstUncompressedIndex int
	//decoder.Decode(&FirstUncompressedIndex)
	//log.FirstUncompressedIndex = LastCompressedTerm
	//var UncompressedEntries []LogEntry
	//decoder.Decode(&UncompressedEntries)
	//log.UncompressedEntries = UncompressedEntries
	//rf.Log = log

	var log LogOne
	if decoder.Decode(&log) != nil {
		panic("Error decoding log!")
	} else {
		rf.Log = log
	}

	// These will be 0 if this is a newly created raft, but if reading from storage and there are compressed entries,
	// these lines are needed to maintain the invariant that lastIndexInSnapshot<=lastApplied<=commitIndex.
	rf.lastApplied = rf.lastIndexInSnapshot()
	rf.commitIndex = rf.lastApplied

	ad.DebugObj(rf, ad.RPC, "State read from stable storage. CurrentTerm=%d, VotedFor=%d, len(log)=%d",
		rf.CurrentTerm, rf.VotedFor, rf.Log.length())
	rf.assertInvariants()
}
