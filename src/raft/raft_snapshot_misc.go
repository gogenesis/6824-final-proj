package raft

// Return the size of this Raft's persisted state in bytes.
func (rf *Raft) StateSizeBytes() int {
	rf.lock()
	defer rf.unlock()

	return rf.persister.RaftStateSize()
}

func (rf *Raft) Snapshot(stateMachineState []byte, lastIncludedIndex int) {
	rf.lock()
	defer rf.unlock()

	if lastIncludedIndex > rf.Log.lastCompressedIndex() {
		rf.snapshotWithLock(stateMachineState, lastIncludedIndex)
	} else {
		debug(rf, TRACE, "Ignoring snapshot request because lastIncludedIndex %d <= my last snapshot index %d",
			lastIncludedIndex, rf.Log.lastCompressedIndex())
	}
}

func (rf *Raft) snapshotWithLock(stateMachineState []byte, lastIncludedIndex int) {
	debug(rf, RPC, "Snapshotting. lastIncludedIndex=%d, lastIncludedTerm=%d", lastIncludedIndex, rf.Log.lastCompressedTerm())
	assert(lastIncludedIndex > rf.Log.lastCompressedIndex()) // can't snapshot a subset of the existing snapshot
	assert(lastIncludedIndex <= rf.lastApplied)              // can't snapshot something that hasn't been sent to the state machine yet

	rf.Log.compressEntriesUpTo(lastIncludedIndex) // automatically handles lastIncludedTerm.
	rf.assertInvariants()
	rf.persister.SaveStateAndSnapshot(rf.getPersistState(), stateMachineState)
	debug(rf, TRACE, "Done with snapshot.")
}
