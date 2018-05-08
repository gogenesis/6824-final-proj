package raft

import (
	"fmt"
)

type InstallSnapshotArgs struct {
	Term              int    // leader's term
	LeaderId          int    //so follower can redirect clients
	LastIncludedIndex int    //the snapshot replaces all entries up through and including this index
	LastIncludedTerm  int    //term of lastIncludedIndex
	Data              []byte //raw bytes of the snapshot chunk, starting at offset
}

type InstallSnapshotReply struct {
	Term int // follower's term, for leader to update itself
}

func (rf *Raft) sendInstallSnapshot(peerNum int) {
	rf.lock()

	if !rf.isAlive || rf.CurrentElectionState != Leader {
		rf.unlock()
		return
	}

	args := InstallSnapshotArgs{}
	args.Term = rf.CurrentTerm
	args.LeaderId = rf.me
	args.LastIncludedIndex = rf.lastIndexInSnapshot()
	args.LastIncludedTerm = rf.Log.lastCompressedTerm()
	args.Data = rf.persister.ReadSnapshot()
	reply := InstallSnapshotReply{}

	//if rf.matchIndex[peerNum] >= rf.lastIndexInSnapshot() {
	//	ad.DebugObj(rf, ad.TRACE, "Would send InstallSnapshot to %d, LastIncludedIndex=%d, but they already match indices through %d, "+
	//		"so there's no point.", peerNum, args.LastIncludedIndex, rf.matchIndex[peerNum])
	//	rf.unlock()
	//	return
	//}

	ad.DebugObj(rf, ad.RPC, "Sending InstallSnapshot to %d, LastIncludedIndex = %d", peerNum, args.LastIncludedIndex)
	rf.unlock()

	ok := rf.peers[peerNum].Call("Raft.InstallSnapshot", &args, &reply)

	rf.lock()
	defer rf.unlock()

	if !(rf.isAlive && rf.CurrentElectionState == Leader) {
		return
	}
	if ok {
		rf.updateTermIfNecessary(reply.Term)
		ad.DebugObj(rf, ad.TRACE, "Received successful response to InstallSnapshot sent to %d with LastIncludedIndex=%d", peerNum, args.LastIncludedIndex)
		rf.matchIndex[peerNum] = args.LastIncludedIndex
		rf.nextIndex[peerNum] = args.LastIncludedIndex + 1
	} else {
		ad.DebugObj(rf, ad.TRACE, "Received failed response to InstallSnapshot sent to %d with LastIncludedIndex=%d", peerNum, args.LastIncludedIndex)
	}
}

func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.lock()

	// 1. Reply immediately if term < currentTerm
	reply.Term = rf.CurrentTerm
	if args.Term < rf.CurrentTerm {
		rf.unlock()
		return
	}

	debugStr := fmt.Sprintf("InstallSnapshot from %d, LastIncludedIndex=%d", args.LeaderId, args.LastIncludedIndex)
	ad.DebugObj(rf, ad.RPC, "Received %v", debugStr)
	rf.updateTermIfNecessary(args.Term)
	rf.resetElectionTimeout()
	if args.Term == rf.CurrentTerm && rf.CurrentElectionState == Leader {
		panic("Received InstallSnapshot from another leader in the same term?!")
	}

	// 2. Create new snapshot file if first chunk (offset is 0)
	// offset is always 0
	snapshotInProgress := make([]byte, len(args.Data))

	// 3. Write data into snapshot file at given offset
	copy(snapshotInProgress, args.Data)

	// 4. Reply and wait for more data chunks if done is false
	// N/A, done is always true

	//5. Save snapshot file, discard any existing or partial snapshot with a smaller index
	//if args.LastIncludedIndex <= rf.lastIndexInSnapshot() {
	//	ad.DebugObj(rf, ad.RPC, "Ignoring %v because my own snapshot already includes indices up to %d", debugStr, rf.lastIndexInSnapshot())
	//	return
	//}
	//ad.DebugObj(rf, ad.TRACE, "Applying Snapshot from %d, LastIncludedIndex=%d", args.LeaderId, args.LastIncludedIndex)
	//rf.lastApplied = max(rf.lastApplied, args.LastIncludedIndex)
	//rf.commitIndex = max(rf.lastApplied, args.LastIncludedIndex)
	//rf.snapshotWithLock(snapshotInProgress, args.LastIncludedIndex)
	//
	//// 6. If existing log entry has same index and term as snapshot’s last included entry, retain log entries following it and reply
	//if rf.Log.lastIndex() > args.LastIncludedIndex {
	//	// Make sure the log has already been compressed to the correct index when applying the snapshot
	//	assertEquals(args.LastIncludedIndex, rf.Log.lastCompressedIndex())
	//	// Since we have entries after LastIncludedIndex, those should still be uncompressed
	//	assert(len(rf.Log.UncompressedEntries) > 0)
	//	ad.DebugObj(rf, ad.TRACE, "Finished InstallSnapshot from %d, LastIncludedIndex=d, not changing state machine state "+
	//		"because my log already contains those entries", args.LeaderId, args.LastIncludedIndex)
	//	return
	//}

	//These are normal invariants, but assert them anyway, just to be sure
	assert(rf.lastIndexInSnapshot() <= rf.lastApplied)
	assert(rf.lastApplied <= rf.commitIndex)
	assert(rf.commitIndex <= rf.lastLogIndex())
	switch {
	case args.LastIncludedIndex <= rf.lastIndexInSnapshot():
		// This is redundant with the snapshot we already have.
		ad.DebugObj(rf, ad.RPC, "Snapshot ends with already compressed entries, ignoring it because "+
			"my own snapshot already includes indices up to %d", rf.lastIndexInSnapshot())
		rf.unlock()
		return

	case rf.lastIndexInSnapshot() < args.LastIncludedIndex &&
		args.LastIncludedIndex <= rf.lastApplied:
		// This is a slightly newer snapshot, but we don't need to tell the state machine about it.
		ad.DebugObj(rf, ad.RPC, "Snapshot ends with applied but not compressed entries, updating stored snapshot with %v "+
			"and changing nothing else", debugStr)
		rf.snapshotWithLock(args.Data, args.LastIncludedIndex)
		rf.unlock()
		return

	case rf.lastApplied < args.LastIncludedIndex &&
		args.LastIncludedIndex <= rf.commitIndex:
		// Update lastApplied so that the next command applied is the one that follows this snapshot.
		ad.DebugObj(rf, ad.CURRENT, "Snapshot ends with committed but not applied entries, Updating LastApplied to %d", args.LastIncludedIndex)
		rf.lastApplied = args.LastIncludedIndex
		rf.snapshotWithLock(args.Data, args.LastIncludedIndex)

	case rf.commitIndex < args.LastIncludedIndex &&
		args.LastIncludedIndex < rf.lastLogIndex():
		// This snapshot includes noncommitted entries.
		ad.DebugObj(rf, ad.CURRENT, "Snapshot ends with stored but not committed entries, updating lastApplied=commitIndex=%d",
			args.LastIncludedIndex)
		rf.lastApplied = args.LastIncludedIndex
		rf.commitIndex = args.LastIncludedIndex
		rf.snapshotWithLock(args.Data, args.LastIncludedIndex)

	case rf.lastLogIndex() <= args.LastIncludedIndex:
		// Discard the entire log because it is obselete at this point.
		ad.DebugObj(rf, ad.CURRENT, "Snapshot ends with entries after the end of my log, replacing entire log.")
		rf.lastApplied = args.LastIncludedIndex
		rf.commitIndex = args.LastIncludedIndex
		rf.snapshotWithLock(args.Data, args.LastIncludedIndex) // automatically handles compression and log replacement
		assertEquals(0, len(rf.Log.UncompressedEntries))
	}

	// 8. Reset state machine using snapshot contents
	ad.DebugObj(rf, ad.RPC, "Finished InstallSnapshot from %d, LastIncludedIndex=%d, resetting state machine state", args.LeaderId,
		args.LastIncludedIndex)

	rf.unlock()
	rf.applyCh <- ApplyMsg{true, snapshotInProgress, args.LastIncludedIndex, rf.CurrentTerm, STATE_RESET}
}
